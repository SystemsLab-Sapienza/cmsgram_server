username = document.querySelector("#name-field");
username.oninput = function(e) {
	if (this.value.length >= 3) {
		checkUsername(this.value);
	}
}

function checkUsername(name) {
	var payload = {
		Key: "username",
		Value: name,
	};
	var req = new XMLHttpRequest();

	req.open("POST", "/isNameTaken", true);
	req.setRequestHeader("Content-Type", "application/json");
	req.send(JSON.stringify(payload));
	req.onload = function() {
		var res = JSON.parse(req.responseText);
		if (res.Value == "true") {
			username.classList.add("taken");
		} else {
			username.classList.remove("taken");
		}
	}
}
