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

type OpenRouterRequest struct {
	Model          string          `json:"model"`
	Messages       []OpenRouterMsg `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
	Type       string      `json:"type"`
	JSONSchema *JSONSchema `json:"json_schema,omitempty"`
}

type JSONSchema struct {
	Name   string                 `json:"name"`
	Strict bool                   `json:"strict"`
	Schema map[string]interface{} `json:"schema"`
}

type OpenRouterMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func openRouterWebhookHandler(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("[%s] Received OpenRouter webhook:\n", time.Now().Format(time.RFC3339))
	fmt.Printf("  Word:     %s\n", req.Word)
	fmt.Printf("  Context:  %s\n", req.Context)
	fmt.Printf("  Source:   %s\n", req.Source)
	fmt.Printf("  Language: %s\n", req.Language)
	fmt.Printf("  Time:     %s\n", req.Timestamp)
	fmt.Println("------------------------------------------------")

	// Get OpenRouter API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENROUTER_API_KEY environment variable not set")
		http.Error(w, "API key missing", http.StatusInternalServerError)
		return
	}

	// Determine target language (default to Traditional Chinese)
	targetLang := req.Language
	if targetLang == "" || targetLang == "en" || targetLang == "en-US" {
		targetLang = "Traditional Chinese"
	}

	// Prepare the prompt (include structure since we're using json_object mode)
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

	// Prepare OpenRouter API request with JSON mode
	// Using json_object mode instead of json_schema for broader model compatibility
	// See: https://openrouter.ai/models?order=newest&supported_parameters=tools&max_price=0
	openRouterReq := OpenRouterRequest{
		// Model: "meta-llama/llama-3.3-70b-instruct:free", // Supports json_object mode
		// Model: "google/gemini-2.0-flash-exp:free", // Supports structured output via response_format
		Model: "openrouter/auto:free", // Automatically selects free models that support structured outputs
		Messages: []OpenRouterMsg{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		ResponseFormat: &ResponseFormat{
			Type: "json_object", // Simpler format, more widely supported than json_schema
		},
	}

	jsonData, err := json.Marshal(openRouterReq)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return
	}

	// Make request to OpenRouter API with increased timeout
	fmt.Println("Calling OpenRouter API...")
	client := &http.Client{Timeout: 60 * time.Second}
	apiReq, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	apiReq.Header.Set("Content-Type", "application/json")

	// Use configurable base URL for HTTP-Referer
	baseURL := os.Getenv("WEBHOOK_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default fallback
	}
	apiReq.Header.Set("HTTP-Referer", baseURL)
	apiReq.Header.Set("X-Title", "Inwen Webhook")

	// Log request details
	fmt.Printf("Request to OpenRouter: Model=%s, Prompt length=%d chars\n", openRouterReq.Model, len(prompt))
	start := time.Now()

	apiResp, err := client.Do(apiReq)
	duration := time.Since(start)
	fmt.Printf("OpenRouter API call took: %v\n", duration)
	if err != nil {
		fmt.Printf("Error calling OpenRouter API: %v\n", err)
		http.Error(w, "Failed to call OpenRouter API", http.StatusInternalServerError)
		return
	}
	defer apiResp.Body.Close()

	respBody, err := io.ReadAll(apiResp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}

	fmt.Printf("OpenRouter API response status: %d\n", apiResp.StatusCode)

	if apiResp.StatusCode != http.StatusOK {
		fmt.Printf("OpenRouter API error (status %d): %s\n", apiResp.StatusCode, string(respBody))
		http.Error(w, fmt.Sprintf("OpenRouter API error: %d", apiResp.StatusCode), http.StatusInternalServerError)
		return
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(respBody, &openRouterResp); err != nil {
		fmt.Printf("Error parsing OpenRouter response: %v\n", err)
		http.Error(w, "Failed to parse response", http.StatusInternalServerError)
		return
	}

	var explanation string
	if len(openRouterResp.Choices) > 0 {
		explanation = openRouterResp.Choices[0].Message.Content
		fmt.Printf("OpenRouter Response received: %d characters\n", len(explanation))
		fmt.Println("Response content:")
		fmt.Println(explanation)
	} else {
		fmt.Println("Warning: No choices in OpenRouter response")
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
