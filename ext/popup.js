document.addEventListener('DOMContentLoaded', () => {
    const toggle = document.getElementById('enabledToggle');
    const statusText = document.getElementById('statusText');
    const providerSelect = document.getElementById('providerSelect');

    // Load saved state (default to true for enabled, 'gemini' for provider)
    chrome.storage.local.get(['enabled', 'provider'], (result) => {
        const isEnabled = result.enabled !== undefined ? result.enabled : true;
        const provider = result.provider || 'gemini';

        toggle.checked = isEnabled;
        providerSelect.value = provider;
        updateStatusText(isEnabled);
    });

    // Save enabled state on change
    toggle.addEventListener('change', () => {
        const isEnabled = toggle.checked;
        chrome.storage.local.set({ enabled: isEnabled }, () => {
            updateStatusText(isEnabled);
            console.log('Inwen enabled state saved:', isEnabled);
        });
    });

    // Save provider on change
    providerSelect.addEventListener('change', () => {
        const provider = providerSelect.value;
        chrome.storage.local.set({ provider: provider }, () => {
            console.log('Inwen provider saved:', provider);
        });
    });

    function updateStatusText(enabled) {
        statusText.textContent = enabled ? 'Enabled' : 'Disabled';
        statusText.style.color = enabled ? '#2ecc71' : '#95a5a6';
    }
});
