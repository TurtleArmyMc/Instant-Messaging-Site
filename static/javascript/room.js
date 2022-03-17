window.onload = function () {
    let newMessageForm = document.getElementById("newMessageForm");
    newMessageForm.focus();
    newMessageForm.scrollIntoView();

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
        // Just using "/stream" as a relative path seems to work fine in
        // Firefox, but not in Chrome
        let streamUrl = new URL(document.URL +
            (document.URL.charAt(document.URL.length-1) === '/' ? "stream" : "/stream")
        );
        // Stream of new messages
        var source = new EventSource(streamUrl);
        source.addEventListener(
            "text_message",
            // Handle new message event
            function (e) {
                let p = document.createElement("p");
                let message = JSON.parse(e.data);
                p.textContent = message.content;
                let messagesDisplay = document.getElementById("messages");
                messagesDisplay.appendChild(p);
                newMessageForm.scrollIntoView();
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