package main

type WebhookRequest struct {
	Word      string `json:"word"`
	Context   string `json:"context"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}
