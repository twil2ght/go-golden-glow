package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goldenglow/pkg/container/fetcher"
	"goldenglow/pkg/database"
	"goldenglow/pkg/node"
)

var (
	repo database.Repository
	f    fetcher.Interface
)

func main() {
	repo = database.DefaultJSONRepo()
	if err := repo.Init(); err != nil {
		panic(fmt.Sprintf("failed to init database: %v", err))
	}
	defer repo.Shutdown()

	f = fetcher.New(repo, node.DefaultFactory)

	http.HandleFunc("/", serveHTML)
	http.HandleFunc("/api/nodes", handleNodes)
	http.HandleFunc("/api/containers", handleContainers)

	fmt.Println("HTML tool listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleNodes(w http.ResponseWriter, r *http.Request) {
	nodeSet, err := repo.HGet("nodeSet")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	nodes := make([]string, 0, len(nodeSet))
	for k := range nodeSet {
		nodes = append(nodes, k)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nodes)
}

type containerDetail struct {
	Hash   string   `json:"hash"`
	TNodes []string `json:"tNodes"`
	RNodes []string `json:"rNodes"`
}

func handleContainers(w http.ResponseWriter, r *http.Request) {
	nodeValue := r.URL.Query().Get("node")
	if nodeValue == "" {
		http.Error(w, "missing node query param", http.StatusBadRequest)
		return
	}

	containerHashes := make(map[string]struct{})

	if t2c, err := repo.HGet("T->C:" + nodeValue); err == nil {
		for h := range t2c {
			containerHashes[h] = struct{}{}
		}
	}
	if r2c, err := repo.HGet("R->C:" + nodeValue); err == nil {
		for h := range r2c {
			containerHashes[h] = struct{}{}
		}
	}

	containers := make([]containerDetail, 0, len(containerHashes))
	for hash := range containerHashes {
		cd := containerDetail{Hash: hash}

		tNodes := f.T(hash)
		for _, n := range tNodes {
			cd.TNodes = append(cd.TNodes, n.Value())
		}

		rNodes := f.R(hash)
		for _, n := range rNodes {
			cd.RNodes = append(cd.RNodes, n.Value())
		}

		containers = append(containers, cd)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(containers)
}

func serveHTML(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlPage))
}

const htmlPage = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Golden Glow — Node Explorer</title>
<style>
:root {
	--bg: #0d1117;
	--bg-surface: #161b22;
	--bg-card: #161b22;
	--bg-card-header: #1c2128;
	--bg-tag: #0d1117;
	--bg-input: #0d1117;
	--border: #30363d;
	--text: #c9d1d9;
	--text-heading: #f0f6fc;
	--text-secondary: #8b949e;
	--text-placeholder: #484f58;
	--accent: #58a6ff;
	--tag-green: #7ee787;
	--ns-purple: #d2a8ff;
	--match-bg: #1f3a2f;
	--match-color: #7ee787;
	--match-outline: #238636;
	--hover-bg: #161b22;
	--active-bg: #1f2937;
}
[data-theme="light"] {
	--bg: #ffffff;
	--bg-surface: #f6f8fa;
	--bg-card: #ffffff;
	--bg-card-header: #f6f8fa;
	--bg-tag: #f6f8fa;
	--bg-input: #ffffff;
	--border: #d0d7de;
	--text: #24292f;
	--text-heading: #1f2328;
	--text-secondary: #656d76;
	--text-placeholder: #8c959f;
	--accent: #0969da;
	--tag-green: #1a7f37;
	--ns-purple: #8250df;
	--match-bg: #dafbe1;
	--match-color: #1a7f37;
	--match-outline: #1a7f37;
	--hover-bg: #f3f4f6;
	--active-bg: #ddf4ff;
}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,Segoe UI,Roboto,sans-serif;background:var(--bg);color:var(--text);height:100vh;display:flex;flex-direction:column}
header{background:var(--bg-surface);border-bottom:1px solid var(--border);padding:12px 20px;display:flex;align-items:center;gap:16px}
header h1{font-size:18px;font-weight:600;color:var(--text-heading);white-space:nowrap}
header .stats{font-size:14px;color:var(--text-secondary);flex:1}
.theme-btn{width:34px;height:34px;border-radius:6px;border:1px solid var(--border);background:var(--bg-input);color:var(--text);cursor:pointer;font-size:16px;display:flex;align-items:center;justify-content:center;flex-shrink:0}
.theme-btn:hover{background:var(--hover-bg)}
main{display:flex;flex:1;overflow:hidden}
.panel{display:flex;flex-direction:column}
.panel-left{width:440px;min-width:300px;border-right:1px solid var(--border);background:var(--bg)}
.panel-right{flex:1;background:var(--bg)}
.panel-header{padding:12px 16px;border-bottom:1px solid var(--border);background:var(--bg-surface)}
.panel-header h2{font-size:14px;font-weight:600;color:var(--text-heading)}
.search-box{width:100%;padding:8px 12px;background:var(--bg-input);border:1px solid var(--border);border-radius:6px;color:var(--text);font-size:13px;outline:none;margin:8px 0}
.search-box:focus{border-color:var(--accent)}
.search-box::placeholder{color:var(--text-placeholder)}
.node-list{flex:1;overflow-y:auto;padding:4px 0}
.node-item{padding:8px 16px;font-size:15px;font-family:ui-monospace,SFMono-Regular,Consolas,monospace;cursor:pointer;border-left:3px solid transparent;transition:background .1s;word-break:break-all}
.node-item:hover{background:var(--hover-bg)}
.node-item.active{background:var(--active-bg);border-left-color:var(--accent);color:var(--text-heading)}
.node-item .tag{color:var(--tag-green);font-weight:600}
.node-item .ns{color:var(--ns-purple)}
.empty-state{display:flex;align-items:center;justify-content:center;height:100%;color:var(--text-secondary);font-size:14px;flex-direction:column;gap:8px}
.container-list{flex:1;overflow-y:auto;padding:16px}
.container-card{background:var(--bg-card);border:1px solid var(--border);border-radius:8px;margin-bottom:12px;overflow:hidden}
.container-card-header{padding:10px 16px;background:var(--bg-card-header);border-bottom:1px solid var(--border);font-size:14px;font-weight:600;color:var(--text-heading);display:flex;align-items:center;gap:8px}
.container-card-header .hash{font-family:ui-monospace,SFMono-Regular,Consolas,monospace;font-size:13px;color:var(--text-secondary);font-weight:400}
.container-card-body{padding:12px 16px}
.node-section{margin-bottom:8px}
.node-section:last-child{margin-bottom:0}
.node-section-label{font-size:13px;font-weight:600;color:var(--text-secondary);text-transform:uppercase;letter-spacing:.5px;margin-bottom:6px}
.node-tag{font-size:15px;font-family:ui-monospace,SFMono-Regular,Consolas,monospace;padding:4px 10px;background:var(--bg-tag);border-radius:4px;margin:2px 0;display:block;word-break:break-all}
.node-tag.match{background:var(--match-bg);color:var(--match-color);outline:1px solid var(--match-outline)}
.node-stats{font-size:13px;color:var(--text-secondary);padding:4px 16px}
.hidden{display:none}
</style>
</head>
<body>
<header>
	<h1>Golden Glow &mdash; Node Explorer</h1>
	<span class="stats" id="headerStats"></span>
	<button class="theme-btn" id="themeBtn" title="Toggle theme" onclick="toggleTheme()">&#9788;</button>
</header>
<main>
	<div class="panel panel-left">
		<div class="panel-header">
			<h2>Nodes</h2>
			<input class="search-box" id="search" type="text" placeholder="Filter nodes&hellip;" autofocus>
		</div>
		<div class="node-stats" id="nodeStats"></div>
		<div class="node-list" id="nodeList"></div>
	</div>
	<div class="panel panel-right">
		<div class="panel-header">
			<h2 id="detailTitle">Containers &mdash; click a node to inspect</h2>
		</div>
		<div class="container-list" id="containerList">
			<div class="empty-state">
				<span style="font-size:32px;opacity:.3">&#8598;</span>
				<span>Select a node from the list to see its containers</span>
			</div>
		</div>
	</div>
</main>
<script>
let allNodes = [];
let activeNode = null;

async function init() {
	const resp = await fetch('/api/nodes');
	allNodes = await resp.json();
	allNodes.sort();
	document.getElementById('headerStats').textContent = allNodes.length + ' total nodes';
	renderNodes(allNodes);
}

function renderNodes(nodes) {
	const list = document.getElementById('nodeList');
	const stats = document.getElementById('nodeStats');
	stats.textContent = 'Showing ' + nodes.length + ' of ' + allNodes.length + ' nodes';

	if (nodes.length === 0) {
		list.innerHTML = '<div class="empty-state"><span>No nodes match the filter</span></div>';
		return;
	}

	list.innerHTML = nodes.map(n => {
		let display = escapeHTML(n);
		// Colorize [tag] prefixes
		display = display.replace(/^\[(\w+)(:[^\]]+)?\]/, '<span class="tag">[$1$2]</span>');
		// Colorize [namespace:...] inside
		display = display.replace(/\[(\w+):([^\]]+)\]/g, '<span class="ns">[$1:$2]</span>');
		const activeClass = (n === activeNode) ? ' active' : '';
		return '<div class="node-item' + activeClass + '" data-value="' + escapeHTML(n) + '" onclick="selectNode(this)">' + display + '</div>';
	}).join('');
}

function escapeHTML(s) {
	const d = document.createElement('div');
	d.textContent = s;
	return d.innerHTML;
}

function selectNode(el) {
	const nodeValue = el.dataset.value;
	document.querySelectorAll('.node-item').forEach(i => i.classList.remove('active'));
	el.classList.add('active');
	activeNode = nodeValue;
	loadContainers(nodeValue);
}

async function loadContainers(nodeValue) {
	const title = document.getElementById('detailTitle');
	const list = document.getElementById('containerList');
	title.textContent = 'Loading...';
	list.innerHTML = '<div class="empty-state"><span>Loading containers&hellip;</span></div>';

	const resp = await fetch('/api/containers?node=' + encodeURIComponent(nodeValue));
	const containers = await resp.json();

	title.textContent = 'Containers for “' + nodeValue + '”  (' + containers.length + ' found)';

	if (containers.length === 0) {
		list.innerHTML = '<div class="empty-state"><span>This node is not referenced in any container</span></div>';
		return;
	}

	list.innerHTML = containers.map((c, i) => {
		let tHTML = c.tNodes.length === 0
			? '<span style="color:#484f58;font-size:11px">(none)</span>'
			: c.tNodes.map(n => {
				const cls = n === nodeValue ? ' node-tag match' : ' node-tag';
				return '<span class="' + cls + '">' + escapeHTML(n) + '</span>';
			}).join('');

		let rHTML = c.rNodes.length === 0
			? '<span style="color:#484f58;font-size:11px">(none)</span>'
			: c.rNodes.map(n => {
				const cls = n === nodeValue ? ' node-tag match' : ' node-tag';
				return '<span class="' + cls + '">' + escapeHTML(n) + '</span>';
			}).join('');

		return '<div class="container-card">' +
			'<div class="container-card-header">' +
				'Container ' + (i + 1) +
				'<span class="hash">' + escapeHTML(c.hash) + '</span>' +
			'</div>' +
			'<div class="container-card-body">' +
				'<div class="node-section">' +
					'<div class="node-section-label">T (triggers) &mdash; ' + c.tNodes.length + '</div>' +
					tHTML +
				'</div>' +
				'<div class="node-section">' +
					'<div class="node-section-label">R (results) &mdash; ' + c.rNodes.length + '</div>' +
					rHTML +
				'</div>' +
			'</div>' +
		'</div>';
	}).join('');
}

document.getElementById('search').addEventListener('input', function(e) {
	const q = e.target.value.toLowerCase();
	const filtered = q === '' ? allNodes : allNodes.filter(n => n.toLowerCase().includes(q));
	renderNodes(filtered);
});

function toggleTheme() {
		const html = document.documentElement;
		const btn = document.getElementById('themeBtn');
		const next = html.dataset.theme === 'dark' ? 'light' : 'dark';
		html.dataset.theme = next;
		btn.innerHTML = next === 'dark' ? '&#9788;' : '&#9790;';
		localStorage.setItem('theme', next);
	}

	(function restoreTheme() {
		const saved = localStorage.getItem('theme') || 'dark';
		document.documentElement.dataset.theme = saved;
		document.getElementById('themeBtn').innerHTML = saved === 'dark' ? '&#9788;' : '&#9790;';
	})();

init();
</script>
</body>
</html>`
