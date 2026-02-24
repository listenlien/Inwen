package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Ollama exposes an OpenAI-compatible API at /v1/chat/completions.
type OllamaRequest struct {
	Model          string            `json:"model"`
	Messages       []OllamaMsg       `json:"messages"`
	Stream         bool              `json:"stream"`
	ResponseFormat *OllamaRespFormat `json:"response_format,omitempty"`
}

type OllamaRespFormat struct {
	Type string `json:"type"` // "json_object" forces valid JSON output
}

type OllamaMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func ollamaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Log the received data
	fmt.Printf("[%s] Received Ollama webhook:\n", time.Now().Format(time.RFC3339))
	fmt.Printf("  Word:     %s\n", req.Word)
	fmt.Printf("  Context:  %s\n", req.Context)
	fmt.Printf("  Source:   %s\n", req.Source)
	fmt.Printf("  Language: %s\n", req.Language)
	fmt.Printf("  Time:     %s\n", req.Timestamp)
	fmt.Println("------------------------------------------------")

	// Determine target language (default to Traditional Chinese)
	targetLang := req.Language
	if targetLang == "" || targetLang == "en" || targetLang == "en-US" {
		targetLang = "Traditional Chinese"
	}

	// Build prompt
	prompt := fmt.Sprintf(`Explain the word "%s" in %s. Return a JSON object with these exact fields:
{
  "word": "the word itself",
  "etymology": "etymology of the word in context (if applicable)",
  "synonyms": [{"word": "synonym in original language", "translation": "translation"}],
  "antonyms": [{"word": "antonym in original language", "translation": "translation"}],
  "context_meaning": "meaning of the word in this context",
  "translation": "translation of the sentence"
}

Context sentence: "%s"
Target language for translations: %s
Keep the total response under 300 words.`,
		req.Word, targetLang, req.Context, targetLang)

	// Get Ollama base URL — defaults to the docker-compose service name on the shared network.
	// Override via OLLAMA_URL env var (e.g. http://192.168.0.81:11434 for an external host).
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "gemma3:4b" // change to whatever model you have pulled in Ollama
	}

	ollamaReq := OllamaRequest{
		Model: model,
		Messages: []OllamaMsg{
			{Role: "user", Content: prompt},
		},
		Stream:         false,
		ResponseFormat: &OllamaRespFormat{Type: "json_object"},
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Calling Ollama API at %s, model=%s...\n", ollamaURL, model)
	client := &http.Client{Timeout: 120 * time.Second}
	apiReq, err := http.NewRequest("POST", ollamaURL+"/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	apiReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	apiResp, err := client.Do(apiReq)
	fmt.Printf("Ollama API call took: %v\n", time.Since(start))
	if err != nil {
		fmt.Printf("Error calling Ollama API: %v\n", err)
		http.Error(w, "Failed to call Ollama API", http.StatusInternalServerError)
		return
	}
	defer apiResp.Body.Close()

	respBody, err := io.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Ollama API response status: %d\n", apiResp.StatusCode)

	if apiResp.StatusCode != http.StatusOK {
		fmt.Printf("Ollama API error (status %d): %s\n", apiResp.StatusCode, string(respBody))
		http.Error(w, fmt.Sprintf("Ollama API error: %d", apiResp.StatusCode), http.StatusInternalServerError)
		return
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		fmt.Printf("Error parsing Ollama response: %v\n", err)
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	var explanation string
	if len(ollamaResp.Choices) > 0 {
		explanation = ollamaResp.Choices[0].Message.Content
		fmt.Printf("Ollama Response received: %d characters\n", len(explanation))
		fmt.Println("Response content:")
		fmt.Println(explanation)
	} else {
		fmt.Println("Warning: No choices in Ollama response")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"explanation": explanation}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
	fmt.Println("Response sent to client successfully")
	fmt.Println("================================================")
}
