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
	http.HandleFunc("/health", healthCheckHandler)

	port := 8080
	fmt.Printf("Server is running on http://localhost:%d\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  - /webhook/gemini (Gemini API)")
	fmt.Println("  - /webhook/openrouter (OpenRouter API)")
	fmt.Println("  - /health (Health check)")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"inwen-webhook"}`))
}
