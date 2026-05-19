# Quality-Gate Investigation Findings

## Summary

The gastown pack hardcodes quality-gate commands (`go test`, `golangci-lint`) in exactly one
location: `crew.template.md`. All other formula references are either configurable (refinery) or
descriptive (polecat mentions tests in passing). The upstream audit confirms this is a known parity
gap.

## Locations Where Quality-Gate Guidance Appears

### 1. crew.template.md — HARDCODED (PRIMARY TARGET)

**File:** `examples/gastown/packs/gastown/assets/prompts/crew.template.md`

- **Lines 333-338** — Session-end "Landing the Plane" protocol:
  ```bash
  go test ./...             # or: make test
  golangci-lint run ./...   # or: make lint
  ```
  File P0 beads if quality gates are broken.

- **Line 379** — Session-end summary checklist item:
  ```
  Status of quality gates (all passing / issues filed)
  ```

This is the **only location in the gastown pack** where `go test` and `golangci-lint` are
hardcoded. These assume a Go project.

### 2. mol-polecat-work.toml — NO hardcoded gates, references template vars

**File:** `examples/gastown/packs/gastown/formulas/mol-polecat-work.toml`

- Lines 85-86: Mentions "self-review + tests" in formula description but no specific commands
- Lines 171-175: Uses `{{setup_command}}` template var (empty = skip) — no quality-gate commands
- Line 184: submit-and-exit step depends on `["self-review"]` — no self-review step defined locally

No `test_command`, `lint_command`, `typecheck_command`, or `build_command` vars exist in this
formula. The polecat formula does NOT configure quality gates at all — it relies on the refinery to
run them after submission.

### 3. mol-refinery-patrol.toml — CONFIGURABLE (already parameterized)

**File:** `examples/gastown/packs/gastown/formulas/mol-refinery-patrol.toml`

- **Lines 40-62:** Template vars for quality gates:
  - `run_tests` (default: "true")
  - `setup_command` (default: "")
  - `typecheck_command` (default: "")
  - `lint_command` (default: "")
  - `test_command` (default: "")
  - `build_command` (default: "")

- **Lines 220-247 (run-tests step):** Each configured check runs in order. Empty commands are
  skipped silently.

- **Lines 250-301 (handle-failures step):** Failed checks cause work bead rejection.

**Problem:** When these vars are empty, the refinery silently skips ALL quality checks.

### 4. Template Fragments — NO quality-gate content

All 8 template fragments in `examples/gastown/packs/gastown/template-fragments/` were checked.
None contain quality-gate, `go test`, or `golangci-lint` references.

### 5. Agent Prompt Templates — NO quality-gate content

All 6 agent prompt templates were checked. None contain specific quality-gate commands.

## Repo Instruction Files (Fallback Targets)

- **AGENTS.md (root):** Lines 264-283 — "Code quality gates" section with `make test`, `go vet`,
  `make dashboard-check`, etc.
- **CLAUDE.md (root):** Line 1 — Contains only `@AGENTS.md` (a reference/include directive).

## Template Rendering Mechanism

- **Template engine:** Go `text/template` package
- **Rendering function:** `cmd/gc/prompt.go` — `renderPrompt()` (line 77)
- **Template data:** `PromptContext` struct with `.AgentName`, `.RigName`, `.CityRoot`, `.WorkDir`
- **Fragment resolution order:** Pack fragment dirs → city-root `template-fragments/` → prompt
  template's own definitions

## What Needs to Change

1. **crew.template.md (lines 333-338):** Replace hardcoded `go test ./...` and
   `golangci-lint run ./...` with a fallback mechanism that reads repo instructions (CLAUDE.md or
   AGENTS.md) when pack-specific guidance is missing/empty.

2. **mol-refinery-patrol.toml (lines 220-247):** The run-tests step silently skips empty commands.
   Add fallback logic so when gate vars are empty, the agent reads the repo's CLAUDE.md / AGENTS.md
   for the real Definition of Done instead of skipping entirely.

3. **mol-polecat-work.toml:** No change needed for quality gates — the polecat delegates quality
   checks to the refinery.

## Upstream Audit Context

`engdocs/archive/analysis/gastown-upstream-audit.md` lines 1255-1256 confirm this is a known gap:

> "Polecats should consult repo CLAUDE.md / AGENTS.md when gate vars are unset. Upstream stopped
> treating empty setup_command / typecheck_command / lint_command / build_command / test_command
> vars as a silent skip and explicitly told polecats to read project CLAUDE.md / AGENTS.md for the
> real Definition of Done."

Status: `[~]` Needed — open parity item.

## Implementation Strategy

The fix should stay inside the Gastown pack / prompt flow. Make the fallback provider-aware:

- For Anthropic/Claude → read CLAUDE.md
- For Kilo/Claude Code/other providers → read AGENTS.md
- Extract quality-gate relevant sections from the matched file
- Include them in the rendered prompt

The existing quality-gate intent must be preserved, not weakened.
