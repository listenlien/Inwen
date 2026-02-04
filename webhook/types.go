package main

type WebhookRequest struct {
	Word      string `json:"word"`
	Context   string `json:"context"`
	Source    string `json:"source"`
	Language  string `json:"language"`
	Timestamp string `json:"timestamp"`
}
