// Just using "path" as a relative path seems to work fine in Firefox,
// but not in Chrome
getRelativeUrl = function (url) {
	if (document.URL.charAt(document.URL.length - 1) === "/") {
		return new URL(document.URL + url);
	}
	return new URL(document.URL + "/" + url);
};

// TODO?: Update timestamps at midnight?
formatDate = function (date) {
	let now = new Date();
	if (
		date.getDate() == now.getDate() &&
		date.getMonth() == now.getMonth() &&
		date.getYear() == now.getYear()
	) {
		// Return time without seconds for messages sent today
		let match = /(\d{1,2}:\d{2}):\d{2}(.*)/.exec(date.toLocaleTimeString());
		return match[1] + match[2];
	} else {
		// Return date for messages sent on a different day
		return date.toLocaleDateString();
	}
};

newTimestampDiv = function (time) {
	// timestampDiv wrapped in order to center timestamp vertically
	let timestampWrapper = document.createElement("div");
	timestampWrapper.classList.add("timestamp");
	let timestampDiv = document.createElement("div");
	timestampDiv.textContent = formatDate(time);
	// TODO: Replace with toggletip
	// Show full date on hover
	timestampDiv.setAttribute("title", time.toLocaleString());
	timestampWrapper.appendChild(timestampDiv);
	return timestampWrapper;
};

getImageUrl = function (content) {
	let url_re =
		/^(https?:\/\/.*?\.(?:apng|avif|gif|ico|jfif|jpg|jpeg|png|svg|webp)(?:\?.*)?)$/gi;
	let match = url_re.exec(content);
	return match != null ? match[1] : null;
};

newTextContentDiv = function (content) {
	let contentDiv = document.createElement("div");
	contentDiv.textContent = content;
	return contentDiv;
};

tryEmbedImg = function (msgDiv, src) {
	let img = document.createElement("img");
	// Replace the link with the image if it is successfully loaded
	img.onload = () => {
		msgDiv.removeChild(msgDiv.lastChild);
		msgDiv.appendChild(img);
		msgDiv.scrollIntoView();
	};
	img.alt = src;
	img.src = src;
};

getYoutubeId = function (content) {
	let id_re =
		/^(?:https:\/\/)?(?:www\.)?youtu(?:\.be\/|be\.com\/(?:embed\/|watch\?v\=))([A-Za-z0-9_-]{11})\S*$/gi;
	let match = id_re.exec(content);
	return match != null ? match[1] : null;
};

newYoutubeEmbed = function (id) {
	let embed = document.createElement("iframe");
	embed.setAttribute(
		"allow",
		"accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
	);
	embed.setAttribute("allowfullscreen", "");
	embed.setAttribute("src", "https://www.youtube.com/embed/" + id);
	return embed;
};

displayMessage = function (msg) {
	let msgDiv = document.createElement("div");
	msgDiv.classList.add("message");
	msgDiv.appendChild(newTimestampDiv(new Date(msg.time)));

	// If the content is a link to a Youtube video, embed it
	let youtubeId = getYoutubeId(msg.content);
	if (youtubeId != null) msgDiv.appendChild(newYoutubeEmbed(youtubeId));
	else {
		// Otherwise, display the content
		msgDiv.appendChild(newTextContentDiv(msg.content));
		// If the content appears to be a link to an image, try to embed it
		let imgUrl = getImageUrl(msg.content);
		if (imgUrl != null) tryEmbedImg(msgDiv, imgUrl);
	}

	let msgDisplay = document.getElementById("messages");
	msgDisplay.insertBefore(msgDiv, msgDisplay.firstChild);
	msgDiv.scrollIntoView();
};

window.onload = function () {
	// Display initial message history, if there is any
	let historyUrl = getRelativeUrl("history");
	fetch(historyUrl)
		.then((response) =>
			response.ok
				? response.json()
				: new Promise(() => {
						messages: [];
				  })
		)
		.then((data) => data.messages.forEach(displayMessage))
		.catch((error) => {
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
		}).catch(function (err) {
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
		};
	} else {
		alert("NOT SUPPORTED");
	}
};
