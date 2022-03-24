// Just using "path" as a relative path seems to work fine in Firefox,
// but not in Chrome
getRelativeUrl = function(url) {
    if (document.URL.charAt(document.URL.length-1) === '/') {
        return new URL(document.URL + url);
    }
    return new URL(document.URL + '/' + url);
}

displayMessage = function(message) {
    let p = document.createElement("p");
    p.classList.add('message');
    p.textContent = message.content;
    let messagesDisplay = document.getElementById("messages");
    messagesDisplay.insertBefore(p, messagesDisplay.firstChild);
    p.scrollIntoView();
}

window.onload = function () {
    // Display initial message history, if there is any
    let historyUrl = getRelativeUrl("history");
    fetch(historyUrl)
        .then(response => response.ok ? response.json() : new Promise(() => {messages: []}))
        .then(data => data.messages.forEach(displayMessage))
        .catch(error => {
            console.log(error);
            alert(error);
        });

    let newMessageForm = document.getElementById("newMessageForm");
    newMessageForm.focus();

    // Handler func for submitting new messages
    newMessageForm.onsubmit = function (e) {
        e.preventDefault();
        /** @type {HTMLFormElement} */
        let form = e.target;
        let data = new FormData(form);
        fetch(form.action, {
            method: form.method,
            body: data,
        })
            .catch(function (err) {
                //Failure
                alert("Error " + err);
            });
        // Clear field after sending a message
        document.getElementById("messageInput").value = "";
    };

    if (!!window.EventSource) {
        let streamUrl = getRelativeUrl("stream");
        // Stream of new messages
        var source = new EventSource(streamUrl);
        source.addEventListener(
            "text_message",
            // Handle new message event
            function (e) {
                let message = JSON.parse(e.data);
                displayMessage(message);
            },
            false
        );

        // Required in Chrome to ensure that the stream is closed when the
        // page is
        window.onbeforeunload = function (e) {
            source.close();
        }

    } else {
        alert("NOT SUPPORTED");
    }
};