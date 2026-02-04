// Import configuration
importScripts('config.js');

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'SEND_TO_WEBHOOK') {
        // Get provider preference from storage
        chrome.storage.local.get('provider', (result) => {
            const provider = result.provider || 'gemini';
            const webhookUrl = provider === 'openrouter'
                ? 'http://localhost:8080/webhook/openrouter'
                : 'http://localhost:8080/webhook/gemini';

            console.log('Inwen: Using provider:', provider, 'URL:', webhookUrl);

            // Detect browser language
            const language = navigator.language || navigator.userLanguage || 'en-US';
            console.log('Inwen: Detected language:', language);

            fetch(webhookUrl, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    word: message.word,
                    context: message.context,
                    source: message.url,
                    language: language,
                    timestamp: new Date().toISOString()
                })
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! status: ${response.status}`);
                    }
                    return response.json();
                })
                .then(data => {
                    console.log('Inwen: Received data from webhook:', data);

                    // Check if we have an explanation field
                    if (!data.explanation) {
                        throw new Error('No explanation field in response');
                    }

                    try {
                        // The explanation field contains a JSON string, parse it
                        const parsedData = JSON.parse(data.explanation);
                        console.log('Inwen: Parsed explanation:', parsedData);
                        sendResponse({ success: true, data: parsedData });
                    } catch (e) {
                        console.error('Inwen: Failed to parse explanation JSON:', e);
                        console.error('Inwen: Raw explanation:', data.explanation);

                        // If parsing fails, send the raw explanation
                        sendResponse({
                            success: false,
                            error: `Failed to parse response: ${e.message}. Raw: ${data.explanation.substring(0, 100)}...`
                        });
                    }
                })
                .catch(error => {
                    console.error('Inwen: Webhook error:', error);
                    sendResponse({ success: false, error: error.message });
                });
        }); // Close chrome.storage.local.get
        return true; // Keep the message channel open for async response
    }
});