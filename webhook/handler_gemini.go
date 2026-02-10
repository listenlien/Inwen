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
	fmt.Printf("  Word:     %s\n", req.Word)
	fmt.Printf("  Context:  %s\n", req.Context)
	fmt.Printf("  Source:   %s\n", req.Source)
	fmt.Printf("  Language: %s\n", req.Language)
	fmt.Printf("  Time:     %s\n", req.Timestamp)
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

	// Determine target language (default to Traditional Chinese)
	targetLang := req.Language
	if targetLang == "" || targetLang == "en" || targetLang == "en-US" {
		targetLang = "Traditional Chinese"
	}

	// The model can be referred to https://ai.google.dev/gemini-api/docs/models. For free tier, each model has different rate limits.
	// model := client.GenerativeModel("gemini-3-flash-preview")
	// model := client.GenerativeModel("gemini-2.5-flash")
	model := client.GenerativeModel("gemini-2.5-flash-lite")

	// Define structured output schema
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"word":      {Type: genai.TypeString, Description: "The word itself"},
			"etymology": {Type: genai.TypeString, Description: "The etymology of the word in context"},
			"synonyms": {
				Type:        genai.TypeArray,
				Description: "A list of synonyms",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"word":        {Type: genai.TypeString, Description: "Synonym in original language"},
						"translation": {Type: genai.TypeString, Description: "Translation of the synonym"},
					},
					Required: []string{"word", "translation"},
				},
			},
			"antonyms": {
				Type:        genai.TypeArray,
				Description: "A list of antonyms",
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"word":        {Type: genai.TypeString, Description: "Antonym in original language"},
						"translation": {Type: genai.TypeString, Description: "Translation of the antonym"},
					},
					Required: []string{"word", "translation"},
				},
			},
			"context_meaning": {Type: genai.TypeString, Description: "The meaning of the word in the context"},
			"translation":     {Type: genai.TypeString, Description: "The translation of the sentence"},
		},
		Required: []string{"word", "context_meaning", "translation"},
	}

	// Configure the model with the schema
	model.GenerationConfig = genai.GenerationConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
	}

	prompt := fmt.Sprintf(`Explain the word "%s" in %s. Provide:
	- The word itself
	- Etymology (if English word)
	- Synonyms with translations to %s
	- Antonyms with translations to %s
	- Use %s to explain the meaning of the word "%s" in the context: "%s"
	- Translation of the sentence to %s
	Keep the total response under 300 words.`,
		req.Word, targetLang, targetLang, targetLang, targetLang, req.Word, req.Context, targetLang)

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
