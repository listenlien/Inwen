# CLAUDE.md - AI Assistant Guide for Inwen

This document provides comprehensive guidance for AI assistants working on the Inwen codebase.

## Project Overview

**Inwen** is a Chrome extension that provides AI-powered word explanations. When users select text on any webpage, a popup appears with structured dictionary information including definitions, etymology, synonyms, antonyms, and Traditional Chinese translations.

**Architecture**: Client-server model
- **Frontend**: Chrome extension (Manifest V3, vanilla JavaScript)
- **Backend**: Local Go webhook server (port 8080)
- **AI APIs**: Google Gemini and OpenRouter (with free models)

## Directory Structure

```
Inwen/
├── ext/                        # Chrome extension (frontend)
│   ├── manifest.json           # Extension config (Manifest V3)
│   ├── background.js           # Service worker - handles API requests
│   ├── content.js              # Content script - text selection & popup UI
│   ├── popup.html              # Settings UI (enable/disable toggle)
│   └── popup.js                # Popup toggle logic
├── webhook/                    # Go backend server
│   ├── main.go                 # Server entry point (HTTP on :8080)
│   ├── types.go                # Shared data types (WebhookRequest)
│   ├── gemini_handler.go       # Google Gemini API integration
│   ├── openrouter_handler.go   # OpenRouter API integration
│   ├── go.mod                  # Go dependencies
│   └── go.sum                  # Dependency checksums (gitignored)
├── .gitignore                  # Ignores .env, binaries, node_modules
└── README.md                   # User documentation
```

## Development Commands

### Backend (Go Server)

```bash
# Install dependencies
cd webhook && go mod download

# Run the server (development)
cd webhook && go run .

# Build binary (production)
cd webhook && go build -o inwen-server .
```

Server runs at `http://localhost:8080` with endpoints:
- `POST /webhook/gemini` - Gemini API
- `POST /webhook/openrouter` - OpenRouter API

### Frontend (Chrome Extension)

No build step required. To reload after changes:
1. Go to `chrome://extensions`
2. Click reload button on Inwen extension
3. Refresh open web pages

### No Testing Framework

This project currently has no automated tests. Manual testing is required:
- Test text selection on various websites
- Verify popup displays correctly
- Check server logs for API responses

## Code Conventions

### Go Conventions

**Error Handling**:
```go
if err != nil {
    http.Error(w, "Error message", http.StatusInternalServerError)
    return
}
```

**HTTP Patterns**:
- Validate method: `if r.Method != http.MethodPost`
- Read body: `io.ReadAll(r.Body)` with `defer r.Body.Close()`
- JSON responses with `json.Marshal()`
- Status codes: 200, 400, 405, 500

**Logging**: Standard `log` and `fmt` packages to stdout

**File Organization**:
- One handler per API in separate files
- Shared types in `types.go`
- Entry point in `main.go`

### JavaScript Conventions

**Event Handling**:
- Use `addEventListener()` pattern
- Async/await with `.then().catch()` chains
- Clean up listeners with `removeEventListener()`

**DOM Manipulation**:
- `Object.assign()` for bulk style assignment
- Template strings for HTML content
- Dynamic element creation with `document.createElement()`

**Storage**:
- `chrome.storage.local.get()` / `chrome.storage.local.set()`
- Callback-based API

**Code Style**:
- camelCase for variables and functions
- Inline styles (no external CSS files)
- Comments for non-obvious logic only

## Key Files Reference

### `webhook/main.go`
Server entry point. Loads `.env`, registers handlers, starts HTTP server.

### `webhook/types.go`
Defines `WebhookRequest` struct used by all handlers:
```go
type WebhookRequest struct {
    Word      string `json:"word"`
    Context   string `json:"context"`
    Source    string `json:"source"`
    Timestamp string `json:"timestamp"`
}
```

### `webhook/gemini_handler.go`
Gemini API integration using `github.com/google/generative-ai-go` SDK. Default model: `gemini-3-flash-preview`.

### `webhook/openrouter_handler.go`
OpenRouter HTTP API integration. Default model: `meta-llama/llama-3.3-70b-instruct:free`. 30-second timeout.

### `ext/background.js`
Service worker handling API requests. **Line 3** contains the endpoint URL - modify to switch between Gemini and OpenRouter.

### `ext/content.js`
Content script injected into all pages. Handles:
- Text selection detection (`mouseup` event)
- Popup creation and positioning
- Response parsing and rendering
- Close handlers (ESC key, outside click)

### `ext/manifest.json`
Chrome extension configuration. Permissions: `storage`, `contextMenus`. Matches all URLs.

## Configuration

### Environment Variables

Create `webhook/.env` (never commit this file):
```
GEMINI_API_KEY=your_gemini_key
OPENROUTER_API_KEY=your_openrouter_key
```

### Switching API Endpoints

Edit `ext/background.js` line 3:
```javascript
// For Gemini:
const WEBHOOK_URL = 'http://localhost:8080/webhook/gemini';

// For OpenRouter:
const WEBHOOK_URL = 'http://localhost:8080/webhook/openrouter';
```

### Changing AI Models

**Gemini** (`webhook/gemini_handler.go`):
```go
model := client.GenerativeModel("gemini-3-flash-preview")
// Alternatives: gemini-2.5-flash, gemini-2.5-flash-lite
```

**OpenRouter** (`webhook/openrouter_handler.go`):
```go
"model": "meta-llama/llama-3.3-70b-instruct:free"
// Alternatives: qwen/qwen-2.5-72b-instruct:free, mistralai/mistral-small-3.1-24b:free
```

## Common Tasks

### Adding a New API Handler

1. Create `webhook/newapi_handler.go`
2. Define handler function: `func HandleNewAPI(w http.ResponseWriter, r *http.Request)`
3. Register in `main.go`: `http.HandleFunc("/webhook/newapi", HandleNewAPI)`
4. Update `ext/background.js` with new endpoint option

### Modifying the Popup UI

Edit `ext/content.js`:
- `createPopupBase()` - popup structure and base styles
- `renderStructuredData()` - content formatting
- Inline styles use `Object.assign(element.style, {...})`

### Adding New Response Fields

1. Update the prompt in handler files (Gemini/OpenRouter)
2. Update `renderStructuredData()` in `ext/content.js` to display new fields

### Changing Popup Position

In `ext/content.js`, find `createPopupBase()` and modify:
```javascript
position: 'fixed',
bottom: '20px',
right: '20px',
// Change to top/left for different positioning
```

## API Request/Response Format

### Request Body (to webhook)
```json
{
  "word": "selected_text",
  "context": "surrounding_sentence",
  "source": "https://page-url.com",
  "timestamp": "2026-01-27T12:00:00Z"
}
```

### Response Body (from webhook)
```json
{
  "explanation": "{\"word\":\"...\",\"etymology\":\"...\",\"synonyms\":[...],\"antonyms\":[...],\"context_meaning\":\"...\",\"translation\":\"...\"}"
}
```

The `explanation` field contains a JSON string that gets parsed in `content.js`.

## Dependencies

### Go Dependencies (`go.mod`)
- `github.com/google/generative-ai-go v0.20.1` - Gemini SDK
- `github.com/joho/godotenv v1.5.1` - .env file loading
- `google.golang.org/api v0.260.0` - Google API client

### JavaScript Dependencies
None - uses vanilla JS and Chrome Extension APIs only.

## Security Considerations

- API keys stored in local `.env` file (gitignored)
- Backend acts as proxy to hide keys from frontend
- CORS enabled for localhost only
- No authentication (local-only server assumption)
- Input validation through JSON parsing

## Troubleshooting

### Server not starting
- Check if port 8080 is in use: `lsof -i :8080`
- Verify `.env` file exists in `webhook/` directory
- Check Go version compatibility

### Extension not detecting text selection
- Verify extension is enabled in popup
- Check `chrome.storage.local` state in DevTools
- Reload extension and refresh page

### API errors
- Check server logs for detailed error messages
- Verify API keys are valid and have quota
- For rate limiting, try a different model or wait

### Popup not appearing
- Check browser console for JavaScript errors
- Verify content script is loaded (check Sources in DevTools)
- Some sites may block content scripts (CSP restrictions)

## Architecture Notes

### Communication Flow
1. User selects text on webpage
2. `content.js` captures selection via `mouseup` event
3. Checks enabled state in `chrome.storage.local`
4. Sends message to `background.js` via `chrome.runtime.sendMessage()`
5. `background.js` makes fetch request to local webhook
6. Handler processes request, calls external AI API
7. Response flows back through the same chain
8. `content.js` renders popup with formatted data

### Why Local Server?
- Keeps API keys secure (not exposed in extension code)
- Enables server-side logging and debugging
- Allows switching AI providers without extension updates
- Provides single point for request/response modification

## Extension Manifest V3 Notes

This extension uses Manifest V3 (latest Chrome standard):
- Service worker instead of background page (`background.js`)
- `chrome.storage.local` for persistence (service workers don't persist state)
- `host_permissions` for localhost access
- Content scripts auto-injected via `matches: ["<all_urls>"]`
