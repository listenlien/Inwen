package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func geminiWebhookHandler(w http.ResponseWriter, r *http.Request) {
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
	fmt.Printf("[%s] Received webhook:\n", time.Now().Format(time.RFC3339))
	fmt.Printf("  Word:    %s\n", req.Word)
	fmt.Printf("  Context: %s\n", req.Context)
	fmt.Printf("  Source:  %s\n", req.Source)
	fmt.Printf("  Time:    %s\n", req.Timestamp)
	fmt.Println("------------------------------------------------")

	// Call Gemini API
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("GEMINI_API_KEY environment variable not set")
		http.Error(w, "API key missing", http.StatusInternalServerError)
		return
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		fmt.Printf("Error creating Gemini client: %v\n", err)
		http.Error(w, "Failed to create Gemini client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// The model can be referred to https://ai.google.dev/gemini-api/docs/models. For free tier, each model has different rate limits.
	model := client.GenerativeModel("gemini-3-flash-preview")
	// model := client.GenerativeModel("gemini-2.5-flash")
	// model := client.GenerativeModel("gemini-2.5-flash-lite")
	prompt := fmt.Sprintf(`Explain the word "%s" in Traditional Chinese. Return a JSON object with the following keys:
	- "word": The word itself.
	- "etymology": The etymology of the word in Traditional Chinese.
	- "synonyms": A list of synonyms, where each item is an object with "word" (English) and "translation" (Traditional Chinese).
	- "antonyms": A list of antonyms, where each item is an object with "word" (English) and "translation" (Traditional Chinese).
	- "context_meaning": The meaning of the word in the context of this sentence: "%s" in Traditional Chinese.
	- "translation": The translation of the sentence into Traditional Chinese.
	The total context words should be less than 300. Do not include markdown code blocks.`, req.Word, req.Context)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		fmt.Printf("Error generating content: %v\n", err)
		http.Error(w, "Failed to generate content", http.StatusInternalServerError)
		return
	}

	var geminiResponse string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		fmt.Println("Gemini Response:")
		for _, part := range resp.Candidates[0].Content.Parts {
			fmt.Println(part)
			geminiResponse += fmt.Sprintf("%v\n", part)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"explanation": geminiResponse,
	})
}
