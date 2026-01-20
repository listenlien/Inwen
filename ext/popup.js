document.addEventListener('DOMContentLoaded', () => {
    const toggle = document.getElementById('enabledToggle');
    const statusText = document.getElementById('statusText');

    // Load saved state (default to true)
    chrome.storage.local.get('enabled', (result) => {
        const isEnabled = result.enabled !== undefined ? result.enabled : true;
        toggle.checked = isEnabled;
        updateStatusText(isEnabled);
    });

    // Save state on change
    toggle.addEventListener('change', () => {
        const isEnabled = toggle.checked;
        chrome.storage.local.set({ enabled: isEnabled }, () => {
            updateStatusText(isEnabled);

            // Optional: Notify active tabs to update immediately if needed,
            // but for this simple use case, checking storage on mouseup is sufficient.
            console.log('Inwen enabled state saved:', isEnabled);
        });
    });

    function updateStatusText(enabled) {
        statusText.textContent = enabled ? 'Enabled' : 'Disabled';
        statusText.style.color = enabled ? '#2ecc71' : '#95a5a6';
    }
});
