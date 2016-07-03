function addEmail() {
	btn = document.getElementById("email-field-aside");
	btn.parentNode.removeChild(btn);

	child = document.createElement("div");
	child.setAttribute("class", "form-field");
	child.innerHTML = emailTemplate.textContent;
	emailSection.appendChild(child);
}
function addSite() {
	btn = document.getElementById("website-field-aside");
	btn.parentNode.removeChild(btn);

	child = document.createElement("div");
	child.setAttribute("class", "form-field");
	child.innerHTML = siteTemplate.textContent;
	addressSection.appendChild(child);
}

emailTemplate = document.getElementById("info-email-template");
emailSection = document.getElementById("email-info"); //change

siteTemplate = document.getElementById("info-address-template");
addressSection = document.getElementById("info-address");
