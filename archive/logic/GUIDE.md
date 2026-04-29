# Data Authoring Guide

This directory stores natural-language rule data consumed by the inference pipeline. Each `.json` file contains one or more rule entries, each with a `description` and a `commands` array. The pipeline matches `[input]` commands (triggers, T-nodes) against incoming knots and produces `[output]` commands (results, R-nodes).

## Directory layout

```
archive/logic/
  <category>/                  # e.g. make_attribution, make_connection
    safe_teach/
      src/                     # Rule templates (use [input] / [output] tags)
        <rule_name>.json
      test/                    # Test fixtures (raw natural-language sentences, no tags)
        <test_name>.json
  exclusion/                   # Exclusion rules
  safe_teach/                  # Global teach-mode start/end rules
  utils.json                   # Shared utility rules (keyed by "in"/"out")
  GUIDE.md
```

- **src** files: every command must start with `[input]` or `[output]`.
- **test** files: contain plain sentences (no tags); the runner feeds them directly as input knots.

## Command format

Each command is a string in the `"commands"` array:

```
"[input]  <natural language template>"
"[output] <natural language template>"
```

- `[input]`  — trigger: when this pattern matches, the container fires.
- `[output]` — result: produced when all triggers of a container are satisfied.

A command may carry multiple bracket tags (e.g. `[input] [repo]`, `[output] [time]`).

## Heads (bracket tags)

### Rule: `->` requires a head

**Any command that contains `->` must also contain at least one bracket head** — one of `[repo]`, `[time]`, `[speak]`, `[ST]`, `[compute]`, `[phrase]`, `[word]`, or a custom head like `[Get]`.

This is because `->` denotes a key-value mapping, and the head tells the system *which* plugin should handle it.

### Plugin-registered heads

These heads are registered by plugins via `OnRegisterDataGen` and map to executor/checker/extractor handlers:

| Head | Plugin | Purpose | Example |
|------|--------|---------|---------|
| `[repo]` | repoaddon | Key-value storage (set/get/check/SSet) | `[repo] set A -> B`, `[repo] get $1` |
| `[time]` | timer | Fetch current time/date components | `[time] get time`, `[time] time -> $1` |
| `[speak]` | speaker | Produce spoken output | `[speak] $1 -> $2` |
| `[ST]` | safeTeach | Toggle teach/ask state | `[ST] teach -> on`, `[ST] teach -> off` |
| `[compute]` | calculator | Arithmetic and comparisons | `[compute] check $1 < $2`, `[compute] get $1 + $2` |
| `[phrase]` | word | Phrase-level operations (length, indexing) | `[phrase] len $1`, `[phrase] get $1 @ $3` |
| `[word]` | word | Character-level string operations | `[word] len $1`, `[word] AddPrefix $1 # $2` |
| `[input]` | builder | Declare a trigger line | `[input] $1` or `[input] Zero says to Susie : $1` |
| `[output]` | builder | Declare a result line | `[output] $1` or `[output] Zero says to Susie : $1` |

### Custom heads

You may introduce custom heads that are NOT registered by any plugin datagen. Example: `[Get]` used in `utils.json` and `res/A_do_B_V2.json` as `[Get] they`.

Custom heads are allowed but should be used sparingly — they must be handled by some consumer in the pipeline to be meaningful.

### Checker-only heads

`check [teach]` and `check [ask]` are special checker forms (no `->`) registered by the safeTeach plugin. They evaluate whether teach/ask mode is active.

## Placeholder system

Template placeholders use `$1`, `$2`, … for positional matching. In addition, letter placeholders A–F denote semantic roles:

| Placeholder | Role | Example usage |
|-------------|------|---------------|
| `A` | First verb | `the verb A can do+sb` |
| `B` | First object (not human) | `how to B someone` |
| `C` | First subject (often a person) | `C says to D : ...` |
| `D` | Second subject (often a person) | `D should tell C` |
| `E` | Second verb | `then you should E them` |
| `F` | Second object (not human) | `then you should E them F` |

Placeholders are matched via `variable.MatchTemplate` using regex patterns like `$1`, `$2`, and single-letter variables `A`–`F`.

## `->` syntax

`->` separates a key from its value in repo-like contexts:

```
[repo] key -> value
[repo] grammar usage of verb A -> do+sb+sth
[time] time -> $1
[ST] teach -> on
```

The right-hand side may be empty (e.g. `[repo] key ->`), which denotes a key with no value — used to check for absence.

## Special constructions

- **`check [teach]` / `check [ask]`** — checker nodes from safeTeach; evaluate to true when the corresponding mode is active.
- **`[repo] SSet key -> value`** — "single set": clears existing entries for the key before setting (unlike plain `set` which adds to a hash set).
- **`[last]`** — a qualifier used inside repo keys, e.g. `[repo] [last] Susie says to Zero -> A` records the most recent thing Susie said.
- **`[Get]`** — a custom extractor head used in `utils.json` to retrieve the pronoun "they" and its referent.

## `utils.json` fields

`utils.json` has an extended schema with `"in"` and `"out"` fields documenting the input/output signature of each utility rule, in addition to `"commands"`.

## Validation checklist

Before committing data, ensure:

1. Every command containing `->` also has a bracket head (`[repo]`, `[time]`, `[ST]`, etc.).
2. All commands in `src/` files start with `[input]` or `[output]`.
3. All commands in `test/` files are plain sentences (no `[input]`/`[output]` tags).
4. JSON is valid (parseable).
5. Placeholder letters A–F are used according to their semantic roles.
6. Custom heads are intentional and documented.
