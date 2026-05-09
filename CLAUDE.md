# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / Test / Run

```bash
# Run the data validator
go run ./cmd/validator/

# Browse nodes and containers in a web UI (http://localhost:8080)
go run ./cmd/tool/html/

# Query containers containing a node from the CLI
go run ./cmd/tool/ <node-value>

# Build and run the interactive (keyboard) mode
go run ./cmd/V2/keyboard

# Build and run the file/batch mode
go run ./cmd/V2/file

# Verify a run was successful (check all expected containers saved)
go run ./cmd/verify/                                          # check generated data only
go run ./cmd/verify/ archive/logic/make_attribution/safe_teach/src  # also check builder containers from archive

# Run all tests
go test ./...

# Run a single package's tests
go test ./pkg/registry/

# Run a specific test
go test ./pkg/registry/ -run TestRegistry_Register

# Trace execution: generates trace/execution.html + trace/execution.json
go run ./cmd/V2/file
# Open trace/execution.html in any browser to view the flow graph and event timeline
```

There is no Makefile or task runner. `go run` and `go test` are the only entry points.

## Architecture

This is a local, non-ML natural language system inspired by programming encapsulation. It matches user input against rule templates and chains results through containers that produce further inferences — essentially a forward-chaining, pattern-matching engine.

### Pipeline (both modes)

```
UserInput → MessageQueue → Runner (worker pool) → knotQueue
  → Template matching → Container (T/R resolution) → new knots → …
```

**Keyboard mode** (`cmd/V2/keyboard`): Reads lines from stdin, pushes them into the message queue, runs until context cancel.
**File mode** (`cmd/V2/file`): Reads directories of natural-language sentences from `archive/`, feeds them as strings into the same pipeline.

### Core types

- **Node** (`pkg/node`): A string value paired with variable sets (keyed by "state" string). Nodes come in four flavors: plain `Node`, `[node:executor]` (performs actions), `[node:checker]` (boolean conditions), `[node:extractor]` (pulls values — e.g., current time).

- **Variable** (`pkg/variable`): Template variables use `$1`, `$2`… syntax. `variable.Set` is `map[string]Item`. `MatchTemplate` in `pkg/node/template/util.go` regex-matches a concrete sentence (like "the time is 14:30 now") against a template (like "the time is $1 now") and extracts variable bindings.

- **Knot** (`pkg/knot`): A `(trigger Node, state string)` pair — the unit of work flowing through the system.

- **Container** (`pkg/container`): The core inference step. Each container is identified by a hash of its T-nodes (triggers) + R-nodes (results), stored in the JSON database. When a knot triggers, the runner finds matching templates, looks up which containers those templates belong to (via `positioner`), fetches the container's full T/R nodes (via `fetcher`), merges variables across T-nodes (handling checker/extractor nodes specially), and produces result knots from R-nodes.

- **Registry** (`pkg/registry`): A generic, thread-safe `map[string]T` with ordered iteration. Used pervasively throughout the codebase as both a DI container and a data structure.

### Storage

Two `database.Repository` implementations, both file-backed:
- **JSON repo** (`database.DefaultJSONRepo`): Hash sets at `archive/Data/json/hash_data.json`. Stores node sets, container T/R mappings, and reverse indexes (T→C, R→C).
- **Redis repo** (`database.DefaultRedisRepo`): Expirable hash sets at `archive/Data/redis/hash_data.json`. Simulates Redis with TTL support.

`pkg/brainsaver` is the service layer on top: `Save(t, r)` hashes T+R node sets, persists forward (C→T, C→R) and reverse (T→C, R→C) mappings, and maintains the global `nodeSet`.

### Plugin system

Plugins live in `plugin/builtin/`. Each plugin's `init()` registers itself with `plugin.DefaultManager`. Plugins implement hook interfaces (all in `pkg/.../hook.go`) to inject behavior:

| Hook interface | Purpose |
|---|---|
| `handler.ExecuteHook` | Register executor handlers (perform actions) |
| `handler.CheckHook` | Register checker handlers (boolean conditions) |
| `handler.ExtractHook` | Register extractor handlers (pull external values) |
| `messageQueue.MsgQueueHook` | Register message providers (e.g., keyboard input) |
| `datagen.Hook` | Register training data generators |
| `template.Hook` | Register conflict rules for template filtering |
| `runner.Hook` | Register idle handlers (called when knot queue drains) |
| `runner.TraceHook` | Register trace handlers (receive execution events for debugging) |

`plugin/mount/doc.go` blank-imports all active plugins. `setup.Init()` (`pkg/setup/setup.go`) wires everything together: it creates all core services, iterates over registered plugins calling their hook registration methods, runs data generation, loads persistent data, and returns the `Background` struct with the wired message queue manager and node factory.

### Data generation and loading

`pkg/datagen` generates JSON files into `generated/` from plugin-provided data specs (T/R sentences + parameter maps). `pkg/dataloader` scans all JSON files under a directory, deserializes them, and feeds them into `brainsaver.Save()`.

### Runner idle detection

When the knot queue AND pending counter are both zero for a configurable timeout, the runner calls `onIdle()`, which resets the node factory, invokes all registered idle handlers, and pulls the next item from the external (message) queue to restart the cycle.

### Debug tracing (`pkg/tracer`)

The runner emits trace events at every key point in the pipeline. The `plugin/builtin/tracer` plugin collects them and writes output files on shutdown:

| Event | When |
|---|---|
| `KnotReceived` | Knot pulled from queue for processing |
| `TemplateMatched` | Template node matched the trigger |
| `TemplateSkipped` | Template filtered out (no container, type mismatch, conflict rule) |
| `ContainerFound` | Container hash found for a template node |
| `ContainerForward` | `container.Forward()` returned true |
| `ContainerReject` | `container.Forward()` returned false (checker failed, var merge failed, etc.) |
| `ResultProduced` | Result knot added to knot queue |

Output files (written to `trace/`):
- `execution.html` — Self-contained HTML page with inline SVG flow graph + event timeline table. Open in any browser.
- `execution.json` — Structured JSON trace with timestamps, node values, variable states, and detail maps

Plugin config constants in `plugin/builtin/tracer/tracer.go`:
- `resetOnIdle` — clear events each idle cycle (default: false, accumulate across cycles)
- `writeOnIdle` — write files each idle cycle (default: false)
- `writeOnShutdown` — write files on plugin shutdown (default: true)

To use programmatically, implement `runner.TraceHook` and call `OnRegisterTraceHandler(mgr)`. The `tracer.Collector` is thread-safe and provides `HTML()`, `JSON()`, and `StringSummary()` output methods.

### Configuration

`config/config.go` holds two constants: `GG` ("Susie") and `User` ("Zero") — the system persona and user name used in natural language templates.

### Rules

- when visualizing or making a graph, use HTML only