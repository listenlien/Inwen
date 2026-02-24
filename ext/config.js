// Extension configuration
const CONFIG = {
    // Webhook server URL - change this for production deployment
    WEBHOOK_URL: 'http://localhost:8080/webhook/gemini',

    // Alternative endpoints:
    // WEBHOOK_URL: 'http://localhost:8080/webhook/gemini',
    // WEBHOOK_URL: 'http://localhost:8080/webhook/openrouter',
    // WEBHOOK_URL: 'http://localhost:8080/webhook/openwebui',
};

// Make config available globally
if (typeof window !== 'undefined') {
    window.INWEN_CONFIG = CONFIG;
}
