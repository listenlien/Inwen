# Inwen

Inwen is a Chrome extension that provides instant, AI-powered explanations for selected text on any webpage. It uses a local Go server to interface with Google's Gemini API, providing structured dictionary definitions, etymology, synonyms, antonyms, and context-aware translations.

## Features

-   **Contextual Analysis**: Explains words based on the sentence they appear in.
-   **Structured Display**: clean popup showing:
    -   Word & Etymology
    -   Synonyms & Antonyms (with translations)
    -   Meaning in Context
    -   Sentence Translation
-   **Interactive UI**:
    -   Loading state indicator.
    -   Close via ESC key or clicking outside.
    -   Enable/Disable toggle via extension icon.
-   **Privacy-First Design**: Uses a local proxy server to handle API keys securely (key stays on your machine).

## Prerequisites

-   **Go**: [Download and install Go](https://go.dev/dl/).
-   **Google Chrome**: Or any Chromium-based browser (Brave, Edge, etc.).
-   **Gemini API Key**: Get one from [Google AI Studio](https://aistudio.google.com/).

## Setup

### 1. Start the Webhook Server

The extension communicates with this local server to process requests.

1.  Navigate to the `webhook` directory:
    ```bash
    cd webhook
    ```
2.  Set your Gemini API key:
    ```bash
    export GEMINI_API_KEY="your_api_key_here"
    ```
3.  Run the server:
    ```bash
    go run main.go
    ```
    You should see: `Server is running on http://localhost:8080`

### 2. Install the Extension

1.  Open Chrome and go to `chrome://extensions`.
2.  Enable **Developer mode** (toggle in the top right).
3.  Click **Load unpacked**.
4.  Select the `Inwen/ext` directory.

## Usage

1.  **Enable the Extension**: Click the Inwen icon in your browser toolbar and ensure the switch is set to "Enabled".
2.  **Select Text**: Highlight any word or phrase on a webpage.
3.  **View Results**: A popup will appear with the analysis.

## Development

-   **Backend**: `webhook/main.go` (Go)
-   **Frontend**: `ext/content.js` (JavaScript), `ext/background.js`, `ext/popup.html`

### Making Changes
If you modify the extension code (`js`, `html`, `manifest`), you must reload it in `chrome://extensions` and refresh your web pages.
