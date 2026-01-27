chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'SEND_TO_WEBHOOK') {
        fetch('http://localhost:8080/webhook/openrouter', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                word: message.word,
                context: message.context,
                source: message.url,
                timestamp: new Date().toISOString()
            })
        })
            .then(response => response.json())
            .then(data => {
                console.log('Data sent successfully', data);
                try {
                    // Extract JSON from the potential Markdown code block
                    let jsonStr = data.explanation.trim();
                    if (jsonStr.startsWith('```json')) {
                        jsonStr = jsonStr.replace(/^```json/, '').replace(/```$/, '').trim();
                    } else if (jsonStr.startsWith('```')) {
                        jsonStr = jsonStr.replace(/^```/, '').replace(/```$/, '').trim();
                    }

                    const parsedData = JSON.parse(jsonStr);
                    sendResponse({ success: true, data: parsedData });
                } catch (e) {
                    console.error('Failed to parse Gemini JSON:', e);
                    // Fallback to sending raw text if parsing fails (or alert user)
                    sendResponse({ success: false, error: 'Failed to parse API response: ' + e.message });
                }
            })
            .catch(error => {
                console.error('Error:', error);
                sendResponse({ success: false, error: error.message });
            });
        return true; // Keep the message channel open for async response
    }
});