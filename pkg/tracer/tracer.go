package tracer

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"sync"
	"time"
)

type EventType int

const (
	EventKnotReceived EventType = iota
	EventTemplateMatched
	EventTemplateSkipped
	EventContainerFound
	EventContainerForward
	EventContainerReject
	EventResultProduced
)

func (e EventType) String() string {
	switch e {
	case EventKnotReceived:
		return "KnotReceived"
	case EventTemplateMatched:
		return "TemplateMatched"
	case EventTemplateSkipped:
		return "TemplateSkipped"
	case EventContainerFound:
		return "ContainerFound"
	case EventContainerForward:
		return "ContainerForward"
	case EventContainerReject:
		return "ContainerReject"
	case EventResultProduced:
		return "ResultProduced"
	default:
		return "Unknown"
	}
}

type Event struct {
	Type          EventType
	Timestamp     time.Time
	NodeValue     string
	State         string
	Detail        map[string]string
	KnotSeq       int // unique seq for the knot being processed (0 = none)
	ParentKnotSeq int // seq of the parent knot that produced this one (0 = root)
}

type Collector struct {
	mu     sync.RWMutex
	events []Event
}

func (c *Collector) Record(e Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *Collector) Events() []Event {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cp := make([]Event, len(c.events))
	copy(cp, c.events)
	return cp
}

func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = nil
}

func (c *Collector) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.events)
}

// tree types for HTML rendering
type treeKnot struct {
	seq       int
	parentSeq int
	value     string
	state     string
	templates []treeTemplate
}

type treeTemplate struct {
	value      string
	skipped    bool
	skipReason string
	containers []treeContainer
}

type treeContainer struct {
	content   string // T-node values, newline-separated
	forwarded bool
	results   []treeResult
}

type treeResult struct {
	value string
	state string
	raw   string
}

// HTML builds a self-contained HTML page with a collapsible tree view and event timeline.
// Each knot expands to show templates → containers → results → child knots.
// Opens directly in any browser — no external dependencies.
func (c *Collector) HTML() string {
	events := c.Events()
	if len(events) == 0 {
		return ""
	}

	// Group events by KnotSeq
	type knotGroup struct {
		knotEv Event
		events []Event
	}
	groups := map[int]*knotGroup{}
	var seqs []int
	for _, ev := range events {
		seq := ev.KnotSeq
		if seq == 0 {
			continue
		}
		if g, ok := groups[seq]; ok {
			g.events = append(g.events, ev)
		} else {
			groups[seq] = &knotGroup{knotEv: ev, events: nil}
			seqs = append(seqs, seq)
		}
	}

	// Build knot tree nodes
	knots := make([]treeKnot, 0, len(groups))
	knotBySeq := map[int]*treeKnot{}
	for _, seq := range seqs {
		g := groups[seq]
		tk := treeKnot{
			seq:       g.knotEv.KnotSeq,
			parentSeq: g.knotEv.ParentKnotSeq,
			value:     g.knotEv.NodeValue,
			state:     g.knotEv.State,
		}
		var curTpl *treeTemplate
		var curCont *treeContainer
		for _, ev := range g.events {
			switch ev.Type {
			case EventTemplateMatched:
				tk.templates = append(tk.templates, treeTemplate{value: ev.NodeValue})
				curTpl = &tk.templates[len(tk.templates)-1]
			case EventTemplateSkipped:
				tk.templates = append(tk.templates, treeTemplate{
					value: ev.NodeValue, skipped: true,
					skipReason: ev.Detail["reason"],
				})
				curTpl = nil
			case EventContainerFound:
				if curTpl != nil {
					curTpl.containers = append(curTpl.containers, treeContainer{})
					curCont = &curTpl.containers[len(curTpl.containers)-1]
				}
			case EventContainerForward:
				if curCont != nil {
					curCont.forwarded = true
					if tn := ev.Detail["t_nodes"]; tn != "" {
						curCont.content = tn
					}
				}
			case EventContainerReject:
				if curCont != nil {
					curCont.forwarded = false
					if tn := ev.Detail["t_nodes"]; tn != "" {
						curCont.content = tn
					}
				}
			case EventResultProduced:
				if curCont != nil {
					curCont.results = append(curCont.results, treeResult{
						value: ev.NodeValue,
						state: ev.State,
						raw:   ev.Detail["result_raw"],
					})
				}
			}
		}
		knots = append(knots, tk)
		knotBySeq[tk.seq] = &knots[len(knots)-1]
	}

	// Build child links: for each knot, find which result in parent produced it
	children := map[int][]*treeKnot{} // parentSeq → child knots
	for i := range knots {
		k := &knots[i]
		if k.parentSeq > 0 {
			children[k.parentSeq] = append(children[k.parentSeq], k)
		}
	}

	return buildHTMLPage(c, knots, children, knotBySeq)
}

func buildHTMLPage(c *Collector, knots []treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot) string {
	var page strings.Builder
	page.WriteString("<!DOCTYPE html>\n")
	page.WriteString("<html lang=\"en\">\n")
	page.WriteString("<head>\n")
	page.WriteString("<meta charset=\"UTF-8\">\n")
	page.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	page.WriteString("<title>Execution Trace</title>\n")
	page.WriteString("<style>\n")
	page.WriteString("  :root { --knot: #6366f1; --tpl: #ca8a04; --cont-fwd: #16a34a; --cont-rej: #dc2626; --result: #0891b2; }\n")
	page.WriteString("  body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f8fafc; color: #1e293b; font-size: 12px; }\n")
	page.WriteString("  .header { padding: 8px 12px 0; }\n")
	page.WriteString("  h1 { font-size: 16px; margin: 0 0 2px; }\n")
	page.WriteString("  .sub { color: #94a3b8; font-size: 11px; margin-bottom: 6px; }\n")
	page.WriteString("  .legend { display: flex; gap: 10px; flex-wrap: wrap; font-size: 11px; padding: 0 12px 8px; border-bottom: 1px solid #e2e8f0; }\n")
	page.WriteString("  .legend span { display: flex; align-items: center; gap: 4px; }\n")
	page.WriteString("  .legend .dot { width: 8px; height: 8px; border-radius: 2px; flex-shrink: 0; }\n")
	page.WriteString("  .tree-wrap { padding: 4px 0 12px; }\n")
	page.WriteString("  .tree-root { margin: 0; padding: 0 0 0 16px; }\n")
	page.WriteString("  details { margin: 0; }\n")
	page.WriteString("  details > summary { cursor: pointer; list-style: none; }\n")
	page.WriteString("  details > summary::-webkit-details-marker { display: none; }\n")
	page.WriteString("  details > summary::before { content: '▸ '; font-size: 9px; color: #94a3b8; transition: transform .15s; display: inline-block; }\n")
	page.WriteString("  details[open] > summary::before { content: '▾ '; }\n")
	page.WriteString("  /* knot node */\n")
	page.WriteString("  .knot-block { margin: 0; }\n")
	page.WriteString("  .knot-block > summary { display: flex; align-items: center; gap: 4px; padding: 2px 6px; background: #eef2ff; border-left: 3px solid var(--knot); border-radius: 3px; margin: 2px 0; font-size: 12px; }\n")
	page.WriteString("  .knot-block > summary .badge { background: var(--knot); color: #fff; font-size: 9px; padding: 0px 5px; border-radius: 8px; font-weight: 600; flex-shrink: 0; }\n")
	page.WriteString("  .knot-block > summary .knot-val { font-weight: 500; }\n")
	page.WriteString("  .knot-block > summary .knot-state { color: #94a3b8; font-size: 10px; }\n")
	page.WriteString("  .knot-block > summary .knot-parent { color: #94a3b8; font-size: 9px; margin-left: auto; }\n")
	page.WriteString("  .knot-body { margin-left: 10px; padding-left: 6px; border-left: 1px dashed #cbd5e1; }\n")
	page.WriteString("  /* template */\n")
	page.WriteString("  .tpl-block { margin: 0; }\n")
	page.WriteString("  .tpl-block > summary { display: flex; align-items: center; gap: 4px; padding: 1px 5px; background: #fefce8; border-left: 3px solid var(--tpl); border-radius: 2px; margin: 1px 0; font-size: 11px; }\n")
	page.WriteString("  .tpl-block > summary .badge { background: var(--tpl); color: #fff; font-size: 9px; padding: 0px 5px; border-radius: 8px; font-weight: 600; flex-shrink: 0; }\n")
	page.WriteString("  .tpl-block.skipped > summary { opacity: .55; background: #f1f5f9; border-left-color: #94a3b8; }\n")
	page.WriteString("  .tpl-block.skipped > summary .badge { background: #94a3b8; }\n")
	page.WriteString("  .tpl-body { margin-left: 6px; padding-left: 4px; border-left: 1px solid #e2e8f0; }\n")
	page.WriteString("  /* container */\n")
	page.WriteString("  .cont-item { margin: 1px 0; border-radius: 2px; font-size: 11px; }\n")
	page.WriteString("  .cont-item summary { display: flex; align-items: flex-start; gap: 4px; padding: 1px 5px; }\n")
	page.WriteString("  .cont-item.fwd { background: #f0fdf4; border-left: 3px solid var(--cont-fwd); }\n")
	page.WriteString("  .cont-item.rej { background: #fef2f2; border-left: 3px solid var(--cont-rej); opacity: .65; }\n")
	page.WriteString("  .cont-item .badge { font-size: 9px; padding: 0px 5px; border-radius: 8px; font-weight: 600; color: #fff; flex-shrink: 0; }\n")
	page.WriteString("  .cont-item.fwd .badge { background: var(--cont-fwd); }\n")
	page.WriteString("  .cont-item.rej .badge { background: var(--cont-rej); }\n")
	page.WriteString("  .cont-content { font-family: monospace; font-size: 10px; color: #475569; white-space: pre-line; line-height: 1.3; }\n")
	page.WriteString("  /* result */\n")
	page.WriteString("  .res-block { margin: 1px 0 1px 6px; }\n")
	page.WriteString("  .res-block > summary { display: flex; align-items: center; gap: 4px; padding: 1px 5px; background: #ecfeff; border-left: 3px solid var(--result); border-radius: 2px; font-size: 11px; }\n")
	page.WriteString("  .res-block > summary .badge { background: var(--result); color: #fff; font-size: 9px; padding: 0px 5px; border-radius: 8px; font-weight: 600; flex-shrink: 0; }\n")
	page.WriteString("  .res-block > summary .child-hint { font-size: 9px; color: #94a3b8; }\n")
	page.WriteString("  .res-body { margin-left: 4px; padding-left: 4px; border-left: 1px solid #cbd5e1; }\n")
	page.WriteString("  /* empty state */\n")
	page.WriteString("  .empty-children { font-size: 10px; color: #94a3b8; font-style: italic; padding: 1px 4px; }\n")
	page.WriteString("  /* timeline table */\n")
	page.WriteString("  .timeline { margin: 16px 12px 16px; }\n")
	page.WriteString("  .timeline h2 { font-size: 14px; margin-bottom: 4px; }\n")
	page.WriteString("  table { border-collapse: collapse; width: 100%; font-size: 11px; }\n")
	page.WriteString("  th, td { text-align: left; padding: 3px 8px; border-bottom: 1px solid #e2e8f0; }\n")
	page.WriteString("  th { background: #f1f5f9; font-weight: 600; color: #475569; }\n")
	page.WriteString("  tr.event-knot { border-left: 3px solid var(--knot); }\n")
	page.WriteString("  tr.event-template { border-left: 3px solid var(--tpl); }\n")
	page.WriteString("  tr.event-container-fwd { border-left: 3px solid var(--cont-fwd); }\n")
	page.WriteString("  tr.event-container-rej { border-left: 3px solid var(--cont-rej); }\n")
	page.WriteString("  tr.event-result { border-left: 3px solid var(--result); }\n")
	page.WriteString("  .detail { font-size: 10px; color: #94a3b8; }\n")
	page.WriteString("  tr:hover { background: #f8fafc; }\n")
	page.WriteString("</style>\n")
	page.WriteString("</head>\n")
	page.WriteString("<body>\n")
	page.WriteString("<div class=\"header\">\n")
	page.WriteString("<h1>Execution Trace — Tree View</h1>\n")
	page.WriteString(fmt.Sprintf("<div class=\"sub\">%d events &middot; %d knots &middot; %s</div>\n",
		c.Len(), len(knots), time.Now().Format(time.RFC3339)))
	page.WriteString("</div>\n")
	page.WriteString("<div class=\"legend\">\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:var(--knot)\"></span> Knot</span>\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:var(--tpl)\"></span> Template</span>\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:var(--cont-fwd)\"></span> Container forward</span>\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:var(--cont-rej)\"></span> Container reject</span>\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:var(--result)\"></span> Result</span>\n")
	page.WriteString("  <span><span class=\"dot\" style=\"background:#94a3b8\"></span> Skipped</span>\n")
	page.WriteString("</div>\n")

	// Only render root knots (no parent) at top level; children are nested recursively
	var roots []treeKnot
	for i := range knots {
		if knots[i].parentSeq == 0 {
			roots = append(roots, knots[i])
		}
	}

	page.WriteString("<div class=\"tree-wrap\">\n")
	page.WriteString("<ul class=\"tree-root\">\n")
	renderTree(&page, roots, children, knotBySeq, "  ")
	page.WriteString("</ul>\n")
	page.WriteString("</div>\n")

	// Timeline table
	page.WriteString("<div class=\"timeline\">\n")
	page.WriteString("<h2>Event Timeline</h2>\n")
	page.WriteString("<table>\n")
	page.WriteString("<thead><tr><th>#</th><th>Time</th><th>Event</th><th>Seq</th><th>Node</th><th>Details</th></tr></thead>\n")
	page.WriteString("<tbody>\n")

	for i, ev := range c.Events() {
		rowClass := ""
		switch ev.Type {
		case EventKnotReceived:
			rowClass = "event-knot"
		case EventTemplateMatched:
			rowClass = "event-template"
		case EventContainerForward:
			rowClass = "event-container-fwd"
		case EventContainerReject:
			rowClass = "event-container-rej"
		case EventResultProduced:
			rowClass = "event-result"
		}

		detailParts := []string{}
		if ev.State != "" {
			detailParts = append(detailParts, "state="+truncate(ev.State, 40))
		}
		for k, v := range ev.Detail {
			if k == "t_nodes" {
				continue // shown in tree, skip in table
			}
			detailParts = append(detailParts, k+"="+truncate(v, 30))
		}
		if ev.ParentKnotSeq > 0 {
			detailParts = append(detailParts, fmt.Sprintf("parent=K%d", ev.ParentKnotSeq))
		}

		seqStr := ""
		if ev.KnotSeq > 0 {
			seqStr = fmt.Sprintf("K%d", ev.KnotSeq)
		}

		page.WriteString(fmt.Sprintf(
			"<tr class=\"%s\"><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td class=\"detail\">%s</td></tr>\n",
			rowClass,
			i+1,
			ev.Timestamp.Format("15:04:05.000"),
			ev.Type.String(),
			seqStr,
			html.EscapeString(truncate(ev.NodeValue, 60)),
			html.EscapeString(strings.Join(detailParts, " · ")),
		))
	}

	page.WriteString("</tbody>\n</table>\n</div>\n")
	page.WriteString("</body>\n</html>")
	return page.String()
}

// renderTree writes <li> elements for knot nodes. Root knots are rendered directly;
// child knots are rendered recursively under their parent's result.
func renderTree(page *strings.Builder, knots []treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot, indent string) {
	for i := range knots {
		k := &knots[i]
		childKnots := children[k.seq]

		// Build a lookup from (value,state) → child knot for matching results
		childByResult := map[string]*treeKnot{}
		for _, ck := range childKnots {
			key := ck.value + "\x00" + ck.state
			childByResult[key] = ck
		}

		page.WriteString(fmt.Sprintf("%s<li>\n", indent))
		renderKnot(page, k, childByResult, children, knotBySeq, indent+"  ")
		page.WriteString(fmt.Sprintf("%s</li>\n", indent))
	}
}

func renderKnot(page *strings.Builder, k *treeKnot, childByResult map[string]*treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot, indent string) {
	hasBody := len(k.templates) > 0
	label := truncate(k.value, 80)
	parentStr := ""
	if k.parentSeq > 0 {
		parentStr = fmt.Sprintf("← K%d", k.parentSeq)
	}

	if hasBody {
		page.WriteString(fmt.Sprintf("%s<details class=\"knot-block\" open>\n", indent))
		page.WriteString(fmt.Sprintf("%s  <summary><span class=\"badge\">K%d</span><span class=\"knot-val\">%s</span>", indent, k.seq, html.EscapeString(label)))
		if k.state != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"knot-state\">[%s]</span>", html.EscapeString(truncate(k.state, 30))))
		}
		if parentStr != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"knot-parent\">%s</span>", parentStr))
		}
		page.WriteString("</summary>\n")
		page.WriteString(fmt.Sprintf("%s  <div class=\"knot-body\">\n", indent))

		for _, tpl := range k.templates {
			renderTemplate(page, &tpl, childByResult, children, knotBySeq, indent+"  ")
		}

		page.WriteString(fmt.Sprintf("%s  </div>\n", indent))
		page.WriteString(fmt.Sprintf("%s</details>\n", indent))
	} else {
		// leaf knot (no templates expanded)
		page.WriteString(fmt.Sprintf("%s<div class=\"knot-block\">\n", indent))
		page.WriteString(fmt.Sprintf("%s  <span style=\"display:inline-flex;align-items:center;gap:4px;padding:1px 5px;font-size:11px;\"><span class=\"badge\">K%d</span>%s</span>", indent, k.seq, html.EscapeString(label)))
		if k.state != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"knot-state\">[%s]</span>", html.EscapeString(truncate(k.state, 30))))
		}
		if parentStr != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"knot-parent\">%s</span>", parentStr))
		}
		page.WriteString("</div>\n")
	}
}

func renderTemplate(page *strings.Builder, tpl *treeTemplate, childByResult map[string]*treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot, indent string) {
	if tpl.skipped {
		page.WriteString(fmt.Sprintf("%s<details class=\"tpl-block skipped\">\n", indent))
		reason := tpl.skipReason
		if reason == "" {
			reason = "skipped"
		}
		page.WriteString(fmt.Sprintf("%s  <summary><span class=\"badge\">SKIP</span>%s &mdash; %s</summary>\n", indent, html.EscapeString(truncate(tpl.value, 60)), reason))
		page.WriteString(fmt.Sprintf("%s</details>\n", indent))
		return
	}

	hasContainers := len(tpl.containers) > 0
	if hasContainers {
		page.WriteString(fmt.Sprintf("%s<details class=\"tpl-block\">\n", indent))
	} else {
		page.WriteString(fmt.Sprintf("%s<div class=\"tpl-block\">\n", indent))
	}
	page.WriteString(fmt.Sprintf("%s  <summary><span class=\"badge\">TPL</span>%s</summary>\n", indent, html.EscapeString(truncate(tpl.value, 60))))

	if hasContainers {
		page.WriteString(fmt.Sprintf("%s  <div class=\"tpl-body\">\n", indent))
		for _, cont := range tpl.containers {
			renderContainer(page, &cont, childByResult, children, knotBySeq, indent+"  ")
		}
		page.WriteString(fmt.Sprintf("%s  </div>\n", indent))
		page.WriteString(fmt.Sprintf("%s</details>\n", indent))
	} else {
		page.WriteString(fmt.Sprintf("%s</div>\n", indent))
	}
}

func renderContainer(page *strings.Builder, cont *treeContainer, childByResult map[string]*treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot, indent string) {
	cls := "fwd"
	tag := "FWD"
	if !cont.forwarded {
		cls = "rej"
		tag = "REJ"
	}
	hasResults := len(cont.results) > 0

	// Show T-node content lines
	contentHTML := html.EscapeString(cont.content)

	if hasResults {
		page.WriteString(fmt.Sprintf("%s<details class=\"cont-item %s\">\n", indent, cls))
		page.WriteString(fmt.Sprintf("%s  <summary><span class=\"badge\">%s</span><span class=\"cont-content\">%s</span></summary>\n", indent, tag, contentHTML))
		for _, res := range cont.results {
			renderResult(page, &res, childByResult, children, knotBySeq, indent+"  ")
		}
		page.WriteString(fmt.Sprintf("%s</details>\n", indent))
	} else {
		page.WriteString(fmt.Sprintf("%s<div class=\"cont-item %s\"><span class=\"badge\">%s</span><span class=\"cont-content\">%s</span></div>\n", indent, cls, tag, contentHTML))
	}
}

func renderResult(page *strings.Builder, res *treeResult, childByResult map[string]*treeKnot, children map[int][]*treeKnot, knotBySeq map[int]*treeKnot, indent string) {
	key := res.value + "\x00" + res.state
	child := childByResult[key]
	display := res.raw
	if display == "" {
		display = res.value
	}

	if child != nil {
		// Build childByResult for the child's children
		grandChildren := children[child.seq]
		grandByResult := map[string]*treeKnot{}
		for _, gc := range grandChildren {
			gkey := gc.value + "\x00" + gc.state
			grandByResult[gkey] = gc
		}

		page.WriteString(fmt.Sprintf("%s<details class=\"res-block\">\n", indent))
		page.WriteString(fmt.Sprintf("%s  <summary><span class=\"badge\">RES</span>%s", indent, html.EscapeString(truncate(display, 60))))
		if res.state != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"child-hint\">[%s]</span>", html.EscapeString(truncate(res.state, 20))))
		}
		page.WriteString(fmt.Sprintf(" <span class=\"child-hint\">→ K%d</span>", child.seq))
		page.WriteString("</summary>\n")
		page.WriteString(fmt.Sprintf("%s  <div class=\"res-body\">\n", indent))
		renderKnot(page, child, grandByResult, children, knotBySeq, indent+"  ")
		page.WriteString(fmt.Sprintf("%s  </div>\n", indent))
		page.WriteString(fmt.Sprintf("%s</details>\n", indent))
	} else {
		page.WriteString(fmt.Sprintf("%s<div class=\"res-block\"><span style=\"display:inline-flex;align-items:center;gap:4px;padding:1px 5px;font-size:11px;background:#ecfeff;border-left:3px solid var(--result);border-radius:2px;\"><span class=\"badge\">RES</span>%s</span>", indent, html.EscapeString(truncate(display, 60))))
		if res.state != "" {
			page.WriteString(fmt.Sprintf(" <span class=\"child-hint\">[%s]</span>", html.EscapeString(truncate(res.state, 20))))
		}
		page.WriteString("</div>\n")
	}
}

func (c *Collector) StringSummary() string {
	events := c.Events()
	if len(events) == 0 {
		return "no events"
	}
	var b strings.Builder
	for _, ev := range events {
		b.WriteString(fmt.Sprintf("[%s] %s: %s", ev.Timestamp.Format("15:04:05.000"), ev.Type, ev.NodeValue))
		if ev.State != "" {
			b.WriteString(fmt.Sprintf(" (state=%s)", ev.State))
		}
		for k, v := range ev.Detail {
			b.WriteString(fmt.Sprintf(" %s=%s", k, truncate(v, 30)))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (c *Collector) JSON() string {
	events := c.Events()
	if len(events) == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteString("[\n")
	for i, ev := range events {
		b.WriteString(fmt.Sprintf("  {\"type\":\"%s\",\"time\":\"%s\",\"node\":\"%s\"",
			ev.Type, ev.Timestamp.Format(time.RFC3339Nano), strings.ReplaceAll(ev.NodeValue, "\"", "\\\"")))
		if ev.State != "" {
			b.WriteString(fmt.Sprintf(",\"state\":\"%s\"", ev.State))
		}
		if len(ev.Detail) > 0 {
			b.WriteString(",\"detail\":{")
			first := true
			for k, v := range ev.Detail {
				if !first {
					b.WriteString(",")
				}
				b.WriteString(fmt.Sprintf("\"%s\":\"%s\"", k, v))
				first = false
			}
			b.WriteString("}")
		}
		b.WriteString("}")
		if i < len(events)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("]")
	return b.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (c *Collector) SortByTime() {
	c.mu.Lock()
	defer c.mu.Unlock()
	sort.Slice(c.events, func(i, j int) bool {
		return c.events[i].Timestamp.Before(c.events[j].Timestamp)
	})
}
