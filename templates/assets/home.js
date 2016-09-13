const max_len = 4096;

chars = document.getElementById("chars-count");

textarea = document.querySelector("textarea");
textarea.oninput = function(e) {
	if (this.value.length > max_len) {
		this.value = this.value.substring(0, max_len);
	}
	chars.innerText = this.value.length + "/"+max_len;
}
