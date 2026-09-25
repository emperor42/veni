"use strict";

(function () {
	var form = document.getElementById("crawl-form");
	var results = document.getElementById("results");
	var breadcrumb = document.getElementById("breadcrumb");

	if (!form || !results || !breadcrumb) return;

	form.addEventListener("submit", async function (event) {
		event.preventDefault();
		var url = document.getElementById("url-input").value.trim();
		var depth = document.getElementById("depth-input").value || "2";
		clearBreadcrumb();
		results.replaceChildren(textElement("p", "loading", "Crawling " + url + "…"));

		try {
			var data = await fetchCrawl(url, Number(depth), []);
			if (!data) throw new Error("empty crawl response");
			renderBreadcrumb(data);
			results.replaceChildren(renderNode(data));
		} catch (error) {
			results.replaceChildren(textElement("p", "error", "Failed to crawl URL."));
		}
	});

	async function fetchCrawl(url, depth, path) {
		var query = new URLSearchParams({ url: url, depth: String(depth) });
		if (path && path.length) query.set("path", JSON.stringify(path));
		var response = await fetch("/crawl?" + query.toString(), { headers: { Accept: "application/json" } });
		if (!response.ok) throw new Error("crawl request failed");
		return response.json();
	}

	function renderBreadcrumb(node) {
		breadcrumb.replaceChildren();
		breadcrumb.hidden = false;
		breadcrumb.style.display = "block";
		breadcrumb.appendChild(textElement("span", "", "Crawl Path:"));
		var path = Array.isArray(node.path) ? node.path : [];
		path.forEach(function (entry, index) {
			if (index > 0) breadcrumb.appendChild(document.createTextNode(" → "));
			var link = document.createElement("a");
			link.textContent = String(entry);
			var safeURL = safeCrawlURL(entry);
			if (safeURL) {
				link.href = safeURL;
				link.target = "_blank";
				link.rel = "noopener noreferrer";
			}
			breadcrumb.appendChild(link);
		});
	}

	function clearBreadcrumb() {
		breadcrumb.replaceChildren();
		breadcrumb.hidden = true;
		breadcrumb.style.display = "none";
	}

	function renderNode(node) {
		if (!node || typeof node !== "object") return document.createTextNode("");

		var details = document.createElement("details");
		if (Number(node.depth) <= 1) details.open = true;
		var summary = document.createElement("summary");
		var titleText = String(node.title == null ? "No Title" : node.title);
		var title = textElement("span", titleText.indexOf("Error:") === 0 ? "error" : "", titleText);
		summary.appendChild(title);
		summary.appendChild(textElement("small", "", " (" + String(node.url == null ? "" : node.url) + ")"));
		details.appendChild(summary);

		var content = textElement("div", "node-content", String(node.content == null ? "" : node.content));
		details.appendChild(content);

		if (node.url && Number(node.depth) > 0) {
			var continueButton = document.createElement("button");
			continueButton.type = "button";
			continueButton.className = "continue-btn";
			continueButton.textContent = "Continue down this path";
			continueButton.addEventListener("click", function () {
				continueCrawl(String(node.url), Number(node.depth) - 1, Array.isArray(node.path) ? node.path : []);
			});
			details.appendChild(continueButton);
		}

		if (Array.isArray(node.links)) {
			node.links.forEach(function (link) {
				details.appendChild(renderNode(link));
			});
		}
		return details;
	}

	async function continueCrawl(url, depth, path) {
		if (depth <= 0) {
			window.alert("Max depth reached");
			return;
		}
		results.replaceChildren(textElement("p", "loading", "Crawling " + url + "…"));
		try {
			var data = await fetchCrawl(url, depth, path);
			if (!data) throw new Error("empty crawl response");
			renderBreadcrumb(data);
			results.replaceChildren(renderNode(data));
		} catch (error) {
			results.replaceChildren(textElement("p", "error", "Failed to continue crawl."));
		}
	}

	function textElement(tagName, className, text) {
		var element = document.createElement(tagName);
		if (className) element.className = className;
		element.textContent = text;
		return element;
	}

	function safeCrawlURL(value) {
		try {
			var parsed = new URL(String(value));
			if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return "";
			return parsed.href;
		} catch (error) {
			return "";
		}
	}
})();
