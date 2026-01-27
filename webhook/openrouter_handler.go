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
	Model    string          `json:"model"`
	Messages []OpenRouterMsg `json:"messages"`
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
	fmt.Printf("  Word:    %s\n", req.Word)
	fmt.Printf("  Context: %s\n", req.Context)
	fmt.Printf("  Source:  %s\n", req.Source)
	fmt.Printf("  Time:    %s\n", req.Timestamp)
	fmt.Println("------------------------------------------------")

	// Get OpenRouter API key
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENROUTER_API_KEY environment variable not set")
		http.Error(w, "API key missing", http.StatusInternalServerError)
		return
	}

	// Prepare the prompt
	prompt := fmt.Sprintf(`Explain the word "%s" in Traditional Chinese. Return a JSON object with the following keys:
	- "word": The word itself.
	- "etymology": The etymology of the word in Traditional Chinese.
	- "synonyms": A list of synonyms, where each item is an object with "word" (English) and "translation" (Traditional Chinese).
	- "antonyms": A list of antonyms, where each item is an object with "word" (English) and "translation" (Traditional Chinese).
	- "context_meaning": The meaning of the word in the context of this sentence: "%s" in Traditional Chinese.
	- "translation": The translation of the sentence into Traditional Chinese.
	The total context words should be less than 300. Do not include markdown code blocks.`, req.Word, req.Context)

	// Prepare OpenRouter API request
	// Free models available: meta-llama/llama-3.3-70b-instruct:free, meta-llama/llama-3.1-405b-instruct:free
	// qwen/qwen-2.5-72b-instruct:free, mistralai/mistral-small-3.1-24b:free, etc.
	// See: https://openrouter.ai/models?order=newest&supported_parameters=tools&max_price=0
	openRouterReq := OpenRouterRequest{
		Model: "meta-llama/llama-3.3-70b-instruct:free", // High-quality free model
		Messages: []OpenRouterMsg{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(openRouterReq)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return
	}

	// Make request to OpenRouter API
	client := &http.Client{Timeout: 30 * time.Second}
	apiReq, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	apiReq.Header.Set("Authorization", "Bearer "+apiKey)
	apiReq.Header.Set("Content-Type", "application/json")
	apiReq.Header.Set("HTTP-Referer", "http://localhost:8080") // Optional
	apiReq.Header.Set("X-Title", "Inwen Webhook")              // Optional

	apiResp, err := client.Do(apiReq)
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

	if apiResp.StatusCode != http.StatusOK {
		fmt.Printf("OpenRouter API error (status %d): %s\n", apiResp.StatusCode, string(respBody))
		http.Error(w, "OpenRouter API error", http.StatusInternalServerError)
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
		fmt.Println("OpenRouter Response:")
		fmt.Println(explanation)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"explanation": explanation,
	})
}
