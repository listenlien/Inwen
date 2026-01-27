package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	http.HandleFunc("/webhook/gemini", geminiWebhookHandler)
	http.HandleFunc("/webhook/openrouter", openRouterWebhookHandler)

	port := 8080
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  - /webhook/gemini (Gemini API)")
	fmt.Println("  - /webhook/openrouter (OpenRouter API)")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
