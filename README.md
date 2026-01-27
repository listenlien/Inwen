# Inwen

Inwen is a Chrome extension that provides instant, AI-powered explanations for selected text on any webpage. It uses a local Go server to interface with multiple AI APIs (Gemini and OpenRouter), providing structured dictionary definitions, etymology, synonyms, antonyms, and context-aware translations in Traditional Chinese.

## Features

- **Dual API Support**: Choose between Google Gemini API or OpenRouter (with access to multiple free models)
- **Contextual Analysis**: Explains words based on the sentence they appear in
- **Structured Display**: Clean popup showing:
  - Word & Etymology
  - Synonyms & Antonyms (with translations)
  - Meaning in Context
  - Sentence Translation
- **Interactive UI**:
  - Loading state indicator
  - Close via ESC key or clicking outside
  - Enable/Disable toggle via extension icon
- **Privacy-First Design**: Uses a local proxy server to handle API keys securely (keys stay on your machine)

## Prerequisites

- **Go**: [Download and install Go](https://go.dev/dl/)
- **Google Chrome**: Or any Chromium-based browser (Brave, Edge, etc.)
- **API Keys**: 
  - **Gemini API Key** (optional): Get one from [Google AI Studio](https://aistudio.google.com/)
  - **OpenRouter API Key** (optional): Get one from [OpenRouter](https://openrouter.ai/keys)

## Project Structure

```
Inwen/
├── webhook/              # Go backend server
│   ├── main.go          # Server entry point
│   ├── gemini_handler.go    # Gemini API handler
│   ├── openrouter_handler.go # OpenRouter API handler
│   ├── types.go         # Shared data structures
│   ├── .env             # Environment variables (not in git)
│   ├── go.mod           # Go dependencies
│   └── go.sum           # Go dependency checksums
└── ext/                 # Chrome extension
    ├── manifest.json    # Extension configuration
    ├── background.js    # Background service worker
    ├── content.js       # Content script (injected into pages)
    ├── popup.html       # Extension popup UI
    └── popup.js         # Popup logic
```

## Setup

### 1. Configure API Keys

1. Navigate to the `webhook` directory:
   ```bash
   cd webhook
   ```

2. Create/edit the `.env` file with your API keys:
   ```bash
   # Gemini API Configuration
   GEMINI_API_KEY=your_gemini_api_key_here

   # OpenRouter API Configuration
   OPENROUTER_API_KEY=your_openrouter_api_key_here
   ```

   **Note**: You only need to configure the API(s) you plan to use.

### 2. Start the Webhook Server

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the server:
   ```bash
   go run .
   ```

   You should see:
   ```
   Server is running on http://localhost:8080
   Available endpoints:
     - /webhook/gemini (Gemini API)
     - /webhook/openrouter (OpenRouter API)
   ```

### 3. Install the Extension

1. Open Chrome and go to `chrome://extensions`
2. Enable **Developer mode** (toggle in the top right)
3. Click **Load unpacked**
4. Select the `Inwen/ext` directory

### 4. Configure the Extension

By default, the extension uses the OpenRouter endpoint. To change this:

1. Edit `ext/background.js`
2. Change the `fetch` URL to your preferred endpoint:
   - For Gemini: `http://localhost:8080/webhook/gemini`
   - For OpenRouter: `http://localhost:8080/webhook/openrouter`

## Usage

1. **Enable the Extension**: Click the Inwen icon in your browser toolbar and ensure the switch is set to "Enabled"
2. **Select Text**: Highlight any word or phrase on a webpage
3. **View Results**: A popup will appear with the AI-powered analysis

## API Endpoints

### `/webhook/gemini` (POST)
Uses Google's Gemini API directly.

**Request Body:**
```json
{
  "word": "example",
  "context": "This is an example sentence.",
  "source": "https://example.com",
  "timestamp": "2026-01-27T12:00:00Z"
}
```

### `/webhook/openrouter` (POST)
Uses OpenRouter API with access to multiple free models.

**Current Model**: `meta-llama/llama-3.3-70b-instruct:free`

**Other Free Models Available**:
- `meta-llama/llama-3.1-405b-instruct:free`
- `qwen/qwen-2.5-72b-instruct:free`
- `mistralai/mistral-small-3.1-24b:free`

See all free models: https://openrouter.ai/models?max_price=0

**Request Body**: Same as Gemini endpoint

**Response** (both endpoints):
```json
{
  "explanation": "{\"word\":\"example\",\"etymology\":\"...\",\"synonyms\":[...],\"antonyms\":[...],\"context_meaning\":\"...\",\"translation\":\"...\"}"
}
```

## Development

### Backend (Go)
- **Entry Point**: `webhook/main.go`
- **Handlers**: `webhook/gemini_handler.go`, `webhook/openrouter_handler.go`
- **Environment**: Uses `godotenv` to load `.env` file automatically

### Frontend (Chrome Extension)
- **Content Script**: `ext/content.js` - Handles text selection and popup display
- **Background Script**: `ext/background.js` - Manages API requests
- **Popup**: `ext/popup.html` + `ext/popup.js` - Extension settings UI

### Making Changes

**Backend Changes:**
1. Modify the Go files
2. Restart the server (`Ctrl+C` then `go run .`)

**Extension Changes:**
1. Modify the extension files
2. Go to `chrome://extensions`
3. Click the reload button on the Inwen extension
4. Refresh any open web pages

## Troubleshooting

### "OPENROUTER_API_KEY environment variable not set"
- Make sure you've created the `.env` file in the `webhook/` directory
- Ensure the `.env` file contains your API key
- Restart the server after adding the key

### "Rate limited" errors
- OpenRouter free tier has rate limits
- Try switching to a different free model in `openrouter_handler.go`
- Or add credits to your OpenRouter account

### Extension not working
- Check that the server is running on `http://localhost:8080`
- Verify the endpoint URL in `ext/background.js` matches your server
- Check the browser console for errors (F12)
- Reload the extension in `chrome://extensions`

## License

This project is open source and available for personal and educational use.
