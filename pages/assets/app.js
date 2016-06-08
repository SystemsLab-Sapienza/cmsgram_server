username = document.querySelector("#name_field");
username.oninput = function(e) {
	if (this.value.length >= 3) {
		checkUsername(this.value);
	}
}

function checkUsername(name) {
	var req = new XMLHttpRequest();
	req.open("POST", "/checkUsername", true);
	req.send(name);
	req.onload = function() {
		if (req.responseText == "true") {
			username.classList.add("taken");
		} else {
			username.classList.remove("taken");
		}
	}
}
