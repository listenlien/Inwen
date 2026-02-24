# Inwen

Inwen is a Chrome extension that provides instant, AI-powered explanations for selected text on any webpage. It uses a local Go server to interface with multiple AI backends (Gemini, OpenRouter, and a self-hosted Ollama instance), returning structured dictionary definitions with etymology, synonyms, antonyms, and context-aware translations. **The output language automatically adapts to your browser's language settings** (defaults to Traditional Chinese).

## Features

- **Three AI Providers**: Choose between Google Gemini, OpenRouter, or a locally-hosted Ollama model
- **Structured Output**: All providers enforce JSON output for reliable, type-safe responses
  - Gemini: Native `genai.Schema` strict validation
  - OpenRouter: `json_object` mode for broad model compatibility
  - Ollama: `json_object` response format via OpenAI-compatible API
- **Contextual Analysis**: Explains words based on the sentence they appear in
- **Auto Language Detection**: Output language matches your browser's language settings (defaults to Traditional Chinese)
- **Double-Click Activation**: Simply double-click any word to trigger the analysis
- **Structured Display**: Clean popup showing:
  - Word & Etymology
  - Synonyms & Antonyms (with translations)
  - Meaning in Context
  - Sentence Translation
- **Interactive UI**:
  - Loading state indicator
  - Close via ESC key or clicking outside
  - Enable/Disable toggle via extension icon
- **Privacy-First Design**: API keys stay on your machine; Ollama runs 100% locally

## Prerequisites

- **Google Chrome**: Or any Chromium-based browser (Brave, Edge, etc.)
- **Docker & Docker Compose**: For running the webhook server and Ollama (recommended)
- Or **Go 1.21+**: For running the webhook server natively
- **API Keys** (only needed for cloud providers):
  - **Gemini API Key** (optional): From [Google AI Studio](https://aistudio.google.com/)
  - **OpenRouter API Key** (optional): From [OpenRouter](https://openrouter.ai/keys)

## Project Structure

```
Inwen/
├── webhook/                   # Go backend server
│   ├── main.go                # Server entry point (port 8088)
│   ├── handler_gemini.go      # Gemini API handler (structured schema output)
│   ├── handler_openrouter.go  # OpenRouter handler (json_object mode)
│   ├── handler_ollama.go      # Ollama handler (local LLM, json_object mode)
│   ├── types.go               # Shared request types
│   ├── docker-compose.yml     # Runs webhook + open-webui/Ollama together
│   ├── Dockerfile             # Webhook server image
│   ├── .env                   # Environment variables (not in git)
│   ├── go.mod
│   └── go.sum
└── ext/                       # Chrome extension
    ├── manifest.json
    ├── background.js          # Routes requests to the selected provider
    ├── content.js             # Double-click handler & popup display
    ├── config.js              # Endpoint reference
    ├── popup.html             # Provider selector UI
    └── popup.js               # Popup state management
```

## Setup

### Option A — Docker Compose (recommended)

This runs the Go webhook server and the Ollama+Open WebUI container together on a shared internal network. Ollama is **not** exposed to the host on port 11434; the webhook reaches it via the Docker service name.

```bash
cd webhook
```

1. Edit `.env` with your cloud API keys (leave blank if using Ollama only):
   ```dotenv
   GEMINI_API_KEY=your_key_here
   OPENROUTER_API_KEY=your_key_here

   # Ollama (resolved internally via Docker network — do not change OLLAMA_URL here)
   OLLAMA_URL=http://open-webui:11434
   OLLAMA_MODEL=phi3:3.8b
   ```

2. Start everything:
   ```bash
   docker compose up -d
   ```

3. Pull a model inside the Ollama container (first run only):
   ```bash
   docker exec inwen-openwebui ollama pull phi3:3.8b
   # or any other model, e.g.: ollama pull gemma3:4b
   ```

4. Services started:
   - **Webhook server**: `http://localhost:8088`
   - **Open WebUI** (optional UI for Ollama): `http://localhost:3003`

### Option B — Run the webhook server natively

```bash
cd webhook
go mod download
go run .
```

Server will start on `http://localhost:8088`.

> When running natively, set `OLLAMA_URL=http://localhost:11434` in `.env` and run Ollama separately.

### Install the Extension

1. Open Chrome and navigate to `chrome://extensions`
2. Enable **Developer mode** (top-right toggle)
3. Click **Load unpacked** and select the `Inwen/ext` directory

## Usage

1. **Enable the Extension**: Click the Inwen icon and toggle to **Enabled**
2. **Select Provider**: Choose your AI provider from the popup:
   - **Google Gemini** — cloud, fast, structured schema
   - **OpenRouter** — cloud, access to many free models
   - **Ollama (Local)** — fully local, private, no API key needed
3. **Double-Click a Word**: On any webpage to trigger the analysis
4. **View Results**: A popup shows etymology, synonyms, antonyms, contextual meaning, and sentence translation

## API Endpoints

All endpoints accept the same `POST` request body:

```json
{
  "word": "arise",
  "context": "Scenarios often arise where primary resolution mechanisms fail.",
  "source": "https://example.com",
  "language": "zh-TW",
  "timestamp": "2026-02-24T09:00:00Z"
}
```

> The `language` field is auto-populated from browser settings. If empty or English (`en`, `en-US`), it defaults to Traditional Chinese.

| Endpoint | Provider | Notes |
|---|---|---|
| `POST /webhook/gemini` | Google Gemini | Structured `genai.Schema`, requires `GEMINI_API_KEY` |
| `POST /webhook/openrouter` | OpenRouter | `json_object` mode, requires `OPENROUTER_API_KEY` |
| `POST /webhook/ollama` | Ollama (local) | `json_object` mode, no API key, model set via `OLLAMA_MODEL` |
| `GET /health` | — | Returns `{"status":"healthy"}` |

**Response** (all endpoints):
```json
{
  "explanation": "{\"word\":\"arise\",\"etymology\":\"...\",\"synonyms\":[...],\"antonyms\":[...],\"context_meaning\":\"...\",\"translation\":\"...\"}"
}
```

> The `explanation` field is a JSON string — the extension parses it with `JSON.parse()`.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GEMINI_API_KEY` | — | Google Gemini API key |
| `OPENROUTER_API_KEY` | — | OpenRouter API key |
| `OLLAMA_URL` | `http://open-webui:11434` | Ollama base URL (Docker internal) |
| `OLLAMA_MODEL` | `phi3:3.8b` | Model name as pulled in Ollama |
| `WEBHOOK_BASE_URL` | `http://localhost:8088` | Used as HTTP-Referer for OpenRouter |

## Docker Compose Architecture

```
┌─────────────────────────────────────────────────────┐
│                  inwen-network (bridge)              │
│                                                     │
│  ┌──────────────┐        ┌───────────────────────┐  │
│  │    inwen     │──────▶│    inwen-openwebui     │  │
│  │ (webhook)    │ :11434 │  (open-webui:ollama)   │  │
│  │ port 8088    │        │  Ollama: 127→0.0.0.0   │  │
│  └──────────────┘        │  WebUI: port 3003      │  │
│        ▲                 └───────────────────────┘  │
└────────┼────────────────────────────────────────────┘
         │
    Chrome Extension
    localhost:8088
```

The webhook container uses the **Docker service name** `open-webui` to reach Ollama at port 11434 on the internal network. `OLLAMA_HOST=0.0.0.0:11434` is set on the Ollama container so it binds to all interfaces (not just loopback).

## Development

### Backend Changes
1. Modify Go files in `webhook/`
2. With Docker: `docker compose up --build -d`
3. Natively: restart `go run .`

### Extension Changes
1. Modify files in `ext/`
2. Go to `chrome://extensions` → click **Reload** on Inwen
3. Refresh any open web pages

## Troubleshooting

| Symptom | Fix |
|---|---|
| `connection refused` on port 11434 | Add `OLLAMA_HOST=0.0.0.0:11434` to the `open-webui` service env (already in `docker-compose.yml`) |
| `No choices in Ollama response` | The model may not be pulled yet: `docker exec inwen-openwebui ollama pull phi3:3.8b` |
| `GEMINI_API_KEY not set` | Add the key to `webhook/.env` and restart |
| Rate limited (OpenRouter) | Switch to a different free model in `handler_openrouter.go` |
| Extension not working | Check the server is running on `:8088`, reload the extension at `chrome://extensions` |

## License

This project is open source and available for personal and educational use.
