document.getElementById("crawl-form").addEventListener("submit", async function(e) {
	e.preventDefault();
	var url = document.getElementById("url-input").value;
	var depth = document.getElementById("depth-input").value || 2;
	var results = document.getElementById("results");
	var breadcrumb = document.getElementById("breadcrumb");
	results.innerHTML = '<p class="loading">Crawling <strong>' + url + '</strong>...</p>';
	breadcrumb.style.display = "none";
	var res = await fetch("/crawl?url=" + encodeURIComponent(url) + "&depth=" + depth);
	var data = await res.json();
	if (!data) {
		results.innerHTML = '<p class="error">Failed to crawl URL.</p>';
		return;
	}
	renderBreadcrumb(data);
	results.innerHTML = renderNode(data);
});

function renderBreadcrumb(node) {
	var el = document.getElementById("breadcrumb");
	el.style.display = "block";
	var html = '<span>Crawl Path:</span> ';
	if (node.path) {
		node.path.forEach(function(p, i) {
			html += '<a href="' + p + '" target="_blank">' + p + '</a>';
			if (i < node.path.length - 1) html += ' → ';
		});
	}
	el.innerHTML = html;
}

function renderNode(node) {
	if (!node) return "";
	var isError = node.title && node.title.startsWith("Error");
	var html = "<details" + (node.depth > 1 ? "" : " open") + ">";
	var title = isError ? '<span class="error">' + node.title + '</span>' : node.title;
	html += "<summary>" + title + " <small>(" + node.url + ")</small></summary>";
	html += '<div class="node-content">' + escapeHtml(node.content) + '</div>';
	if (node.url && node.depth > 0) {
		html += '<button class="continue-btn" onclick="continueCrawl(\'' + escapeJs(node.url) + '\',' + (node.depth - 1) + ')">Continue down this path</button>';
	}
	if (node.links && node.links.length > 0) {
		node.links.forEach(function(link) {
			html += renderNode(link);
		});
	}
	html += "</details>";
	return html;
}

async function continueCrawl(url, depth) {
	if (depth <= 0) { alert("Max depth reached"); return; }
	var results = document.getElementById("results");
	var res = await fetch("/crawl?url=" + encodeURIComponent(url) + "&depth=" + depth);
	var data = await res.json();
	if (!data) return;
	results.innerHTML = renderBreadcrumb(data) || "";
	results.innerHTML = renderNode(data);
}

function escapeHtml(text) {
	if (!text) return "";
	var d = document.createElement("div");
	d.textContent = text;
	return d.innerHTML;
}

function escapeJs(text) {
	return text.replace(/'/g, "\\'").replace(/"/g, "&quot;");
}
