document.addEventListener('dblclick', () => {
    const selection = window.getSelection();
    const selectedText = selection.toString().trim();

    if (selectedText.length > 0) {
        console.log('Inwen: selectedText', selectedText);
        // Check if extension is enabled
        // Check if chrome.storage is available (it might not be in some contexts)
        if (typeof chrome !== 'undefined' && chrome.storage && chrome.storage.local) {
            chrome.storage.local.get('enabled', (result) => {
                const isEnabled = result.enabled !== undefined ? result.enabled : true;
                if (!isEnabled) return;

                // Validate if selected text contains valid English words
                if (typeof nlp !== 'undefined') {
                    const doc = nlp(selectedText);

                    // 1. Check if it's more than one term (e.g., "hello world")
                    const terms = doc.termList();
                    if (terms.length !== 1) return false;

                    // 2. Reject strings with numbers (e.g., GMT+8, v1.0, etc.)
                    if (/[0-9]/.test(selectedText)) return false;

                    // 3. Reject if it's just punctuation or symbols
                    if (!/[a-zA-Z]/.test(selectedText)) return false;

                    // 4. Check if it's a valid word
                    const hasWords = doc.has('#Noun') || doc.has('#Verb') || doc.has('#Adjective') ||
                        doc.has('#Adverb') || doc.has('#Pronoun') || doc.has('#Determiner');

                    console.log('Inwen: hasWords', hasWords);

                    if (!hasWords) {
                        console.log('Inwen: Selected text does not contain valid words, skipping webhook');
                        return;
                    }
                }

                // Get the parent element's text as context
                const context = selection.anchorNode.parentElement.innerText;

                // Validate if context is a valid sentence or paragraph by nlp
                if (typeof nlp !== 'undefined') {
                    const doc = nlp(context);

                    // 1. Check if it's more than one term (e.g., "hello world")
                    const terms = doc.termList();
                    if (terms.length <= 1) return false;

                    // 2. Check if context has at least one sentence with verbs and nouns
                    const hasSentences = doc.sentences().length > 0;
                    const hasContent = doc.has('#Verb') && doc.has('#Noun');

                    if (!hasSentences || !hasContent) {
                        console.log('Inwen: Context is invalid or empty, skipping webhook');
                        return;
                    }

                    console.log('Inwen: Context validation passed:', doc.sentences().length, 'sentences');
                }

                // Show loading popup immediately
                showLoadingPopup();

                // Send data to the background script
                chrome.runtime.sendMessage({
                    type: 'SEND_TO_WEBHOOK',
                    word: selectedText,
                    context: context,
                    url: window.location.href
                }, (response) => {
                    if (response && response.success) {
                        updatePopupContent(response.data);
                    } else {
                        updatePopupError(response ? response.error : 'Unknown error');
                    }
                });
            });
            // } else {
            //     // Fallback: if storage is not available, just proceed as enabled
            //     const context = selection.anchorNode.parentElement.innerText;
            //     showLoadingPopup();

            //     chrome.runtime.sendMessage({
            //         type: 'SEND_TO_WEBHOOK',
            //         word: selectedText,
            //         context: context,
            //         url: window.location.href
            //     }, (response) => {
            //         if (response && response.success) {
            //             updatePopupContent(response.data);
            //         } else {
            //             updatePopupError(response ? response.error : 'Unknown error');
            //         }
            //     });
        }
    }
});

function removePopup() {
    const popup = document.getElementById('inwen-popup');
    if (popup) {
        popup.remove();
    }
    // Remove global event listeners
    document.removeEventListener('keydown', handleEscKey);
    document.removeEventListener('mousedown', handleOutsideClick);
}

function handleEscKey(event) {
    if (event.key === 'Escape') {
        removePopup();
    }
}

function handleOutsideClick(event) {
    const popup = document.getElementById('inwen-popup');
    if (popup && !popup.contains(event.target)) {
        removePopup();
    }
}

function createPopupBase() {
    // Remove existing popup if any
    removePopup();

    const popup = document.createElement('div');
    popup.id = 'inwen-popup';
    Object.assign(popup.style, {
        position: 'fixed',
        bottom: '20px',
        right: '20px',
        width: '320px',
        maxHeight: '500px',
        overflowY: 'auto',
        backgroundColor: '#ffffff',
        border: '1px solid #e0e0e0',
        borderRadius: '12px',
        padding: '0', // Reset padding for better layout
        boxShadow: '0 8px 16px rgba(0,0,0,0.15)',
        zIndex: '2147483647',
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif",
        fontSize: '13px',
        lineHeight: '1.6',
        color: '#333333',
        textAlign: 'left',
        transition: 'all 0.3s ease'
    });

    // Close button container
    const header = document.createElement('div');
    Object.assign(header.style, {
        display: 'flex',
        justifyContent: 'flex-end',
        padding: '10px 15px',
        borderBottom: '1px solid #f0f0f0'
    });

    const closeBtn = document.createElement('button');
    closeBtn.innerHTML = '&times;';
    Object.assign(closeBtn.style, {
        border: 'none',
        background: 'transparent',
        fontSize: '20px',
        lineHeight: '1',
        cursor: 'pointer',
        color: '#999',
        padding: '0'
    });
    closeBtn.onmouseenter = () => closeBtn.style.color = '#333';
    closeBtn.onmouseleave = () => closeBtn.style.color = '#999';
    closeBtn.onclick = removePopup;

    header.appendChild(closeBtn);
    popup.appendChild(header);

    const content = document.createElement('div');
    content.id = 'inwen-popup-content';
    Object.assign(content.style, {
        padding: '15px'
    });

    popup.appendChild(content);
    document.body.appendChild(popup);

    // Add event listeners to close popup
    document.addEventListener('keydown', handleEscKey);
    // Use mousedown instead of click to catch it earlier and avoid conflicts with selection
    document.addEventListener('mousedown', handleOutsideClick);

    return content;
}

function showLoadingPopup() {
    const content = createPopupBase();
    content.innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100px;">
            <div class="inwen-spinner"></div>
            <style>
                .inwen-spinner {
                    border: 3px solid #f3f3f3;
                    border-top: 3px solid #3498db;
                    border-radius: 50%;
                    width: 24px;
                    height: 24px;
                    animation: inwen-spin 1s linear infinite;
                }
                @keyframes inwen-spin {
                    0% { transform: rotate(0deg); }
                    100% { transform: rotate(360deg); }
                }
            </style>
        </div>
        <div style="text-align: center; color: #888; font-size: 11px;">Analyzing...</div>
    `;
}

function updatePopupContent(data) {
    const content = document.getElementById('inwen-popup-content');
    if (content) {
        content.innerHTML = renderStructuredData(data);
    }
}

function updatePopupError(errorMessage) {
    const content = document.getElementById('inwen-popup-content');
    if (content) {
        content.innerHTML = `<div style="color: #d32f2f; background: #ffebee; padding: 10px; border-radius: 4px;"><strong>Error:</strong> ${errorMessage}</div>`;
    }
}

function renderStructuredData(data) {
    if (!data) return '';

    const { word, etymology, synonyms, antonyms, context_meaning, translation } = data;

    const styleLabel = 'font-weight: 600; color: #555; font-size: 10px; text-transform: uppercase; margin-bottom: 4px; display: block;';
    const styleSection = 'margin-bottom: 16px;';
    const styleValue = 'color: #333; font-size: 11px;';

    return `
        <h2 style="margin: 0 0 16px 0; font-size: 24px; font-weight: 700; color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 8px; display: inline-block;">${word}</h2>
        
        <div style="${styleSection} background-color: #f8f9fa; padding: 10px; border-radius: 6px; border-left: 3px solid #3498db;">
            <span style="${styleLabel} color: #2980b9;">Meaning in Context</span>
            <div style="${styleValue}">${context_meaning || '-'}</div>
        </div>

        <div style="${styleSection}">
            <span style="${styleLabel}">Sentence Translation</span>
            <div style="${styleValue} font-style: italic; color: #555;">${translation || '-'}</div>
        </div>

        <div style="${styleSection}">
            <span style="${styleLabel}">Etymology</span>
            <div style="${styleValue}">${etymology || '-'}</div>
        </div>

        <div style="${styleSection}">
            <span style="${styleLabel}">Synonyms</span>
            <div style="${styleValue}">${renderWordList(synonyms)}</div>
        </div>

        <div style="${styleSection}">
            <span style="${styleLabel}">Antonyms</span>
            <div style="${styleValue}">${renderWordList(antonyms)}</div>
        </div>
    `;
}

function renderWordList(list) {
    if (!list || !Array.isArray(list) || list.length === 0) return '-';

    return list.map(item => {
        // Handle both string format (legacy fallback) and object format
        if (typeof item === 'string') return item;
        return `${item.word} <span style="color: #888; font-size: 10px;">(${item.translation})</span>`;
    }).join(', ');
}