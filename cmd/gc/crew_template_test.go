package main

import (
	"io"
	"strings"
	"testing"

	"github.com/gastownhall/gascity/internal/fsys"
)

func TestRenderCrewTemplateQualityGateUsesInstructionsFile(t *testing.T) {
	f := fsys.NewFake()

	crewTemplate := `{{ define "approval-fallacy-crew" }}
## The Approval Fallacy

**There is no approval step.** When your work is done, you act - you don't wait.
{{ end }}

{{ define "propulsion-crew" }}
## Theory of Operation

You execute work on your hook immediately. No waiting.
{{ end }}

{{ define "capability-ledger-work" }}
## The Capability Ledger

Every completion is recorded.
{{ end }}

# Crew Worker Context

{{ template "approval-fallacy-crew" . }}

{{ template "propulsion-crew" . }}

{{ template "capability-ledger-work" . }}

---

## Your Role: CREW WORKER ({{ .AgentName }} in {{ .RigName }})

## Session End Protocol

When ending a session, complete ALL steps below.

1. **File beads for remaining work** that needs follow-up

2. **Run quality gates** (only if code changes were made):
   Consult your provider's instructions file for the quality gate command:
   ```
   cat {{ .InstructionsFile }}
   ```
   File P0 beads if quality gates are broken.

3. **PUSH TO REMOTE - NON-NEGOTIABLE:**
   git push

4. **Handoff or close:**
   git status
`

	f.Files["/city/prompts/crew.template.md"] = []byte(crewTemplate)

	ctx := PromptContext{
		CityRoot:          "/city",
		AgentName:         "gastown/crew/alice",
		RigName:           "gastown",
		WorkDir:           "/work/gastown/crew/alice",
		IssuePrefix:       "gt",
		InstructionsFile:   "CLAUDE.md",
		ProviderKey:       "claude",
		ProviderDisplayName: "Claude Code",
	}

	got := renderPrompt(f, "/city", "", "prompts/crew.template.md", ctx, "", io.Discard, nil, nil, nil)

	if !strings.Contains(got, "CLAUDE.md") {
		t.Errorf("CLAUDE.md not found in rendered output for claude provider")
	}

	if strings.Contains(got, "go test ./...") {
		t.Errorf("hardcoded 'go test ./...' found in rendered output - template should use {{ .InstructionsFile }}")
	}

	if strings.Contains(got, "golangci-lint run") {
		t.Errorf("hardcoded 'golangci-lint run' found in rendered output - template should use {{ .InstructionsFile }}")
	}
}

func TestRenderCrewTemplateQualityGateForNonClaudeProvider(t *testing.T) {
	f := fsys.NewFake()

	crewTemplate := `{{ define "approval-fallacy-crew" }}
## The Approval Fallacy

**There is no approval step.** When your work is done, you act - you don't wait.
{{ end }}

{{ define "propulsion-crew" }}
## Theory of Operation

You execute work on your hook immediately. No waiting.
{{ end }}

{{ define "capability-ledger-work" }}
## The Capability Ledger

Every completion is recorded.
{{ end }}

# Crew Worker Context

{{ template "approval-fallacy-crew" . }}

{{ template "propulsion-crew" . }}

{{ template "capability-ledger-work" . }}

---

## Your Role: CREW WORKER ({{ .AgentName }} in {{ .RigName }})

## Session End Protocol

When ending a session, complete ALL steps below.

1. **File beads for remaining work** that needs follow-up

2. **Run quality gates** (only if code changes were made):
   Consult your provider's instructions file for the quality gate command:
   ```
   cat {{ .InstructionsFile }}
   ```
   File P0 beads if quality gates are broken.

3. **PUSH TO REMOTE - NON-NEGOTIABLE:**
   git push

4. **Handoff or close:**
   git status
`

	f.Files["/city/prompts/crew.template.md"] = []byte(crewTemplate)

	ctx := PromptContext{
		CityRoot:          "/city",
		AgentName:         "gastown/crew/bob",
		RigName:           "gastown",
		WorkDir:           "/work/gastown/crew/bob",
		IssuePrefix:       "gt",
		InstructionsFile:   "AGENTS.md",
		ProviderKey:       "codex",
		ProviderDisplayName: "Codex CLI",
	}

	got := renderPrompt(f, "/city", "", "prompts/crew.template.md", ctx, "", io.Discard, nil, nil, nil)

	if !strings.Contains(got, "AGENTS.md") {
		t.Errorf("AGENTS.md not found in rendered output for non-claude provider")
	}

	if strings.Contains(got, "go test ./...") {
		t.Errorf("hardcoded 'go test ./...' found in rendered output - template should use {{ .InstructionsFile }}")
	}

	if strings.Contains(got, "golangci-lint run") {
		t.Errorf("hardcoded 'golangci-lint run' found in rendered output - template should use {{ .InstructionsFile }}")
	}
}

func TestRenderCrewTemplateQualityGateNoHardcodedGoCommands(t *testing.T) {
	f := fsys.NewFake()

	crewTemplate := `{{ define "approval-fallacy-crew" }}
## The Approval Fallacy
{{ end }}

{{ define "propulsion-crew" }}
## Theory of Operation
{{ end }}

{{ define "capability-ledger-work" }}
## The Capability Ledger
{{ end }}

# Crew Worker Context

{{ template "approval-fallacy-crew" . }}

---

## Session End Protocol

When ending a session, complete ALL steps below.

1. **File beads for remaining work** that needs follow-up

2. **Run quality gates** (only if code changes were made):
   Consult your provider's instructions file:
   ```
   cat {{ .InstructionsFile }}
   ```

3. **PUSH TO REMOTE:**
   git push
`

	f.Files["/city/prompts/crew.template.md"] = []byte(crewTemplate)

	ctx := PromptContext{
		CityRoot:        "/city",
		AgentName:       "gastown/crew/test",
		RigName:         "gastown",
		WorkDir:         "/work/gastown/crew/test",
		IssuePrefix:     "gt",
		InstructionsFile: "CLAUDE.md",
	}

	got := renderPrompt(f, "/city", "", "prompts/crew.template.md", ctx, "", io.Discard, nil, nil, nil)

	if strings.Contains(got, "go test ./...") {
		t.Errorf("rendered output contains hardcoded 'go test ./...' - quality gate must use {{ .InstructionsFile }}")
	}

	if strings.Contains(got, "golangci-lint run") {
		t.Errorf("rendered output contains hardcoded 'golangci-lint run' - quality gate must use {{ .InstructionsFile }}")
	}
}

// TestRenderCrewTemplateQualityGateWithRealTemplate verifies the actual
// crew.template.md uses {{ .InstructionsFile }} instead of hardcoded Go commands.
// This test will fail if the template still contains hardcoded quality gates
// (go test ./... or golangci-lint run) - indicating the template needs updating.
func TestRenderCrewTemplateQualityGateWithRealTemplate(t *testing.T) {
	f := fsys.NewFake()

	f.Files["/city/prompts/crew.template.md"] = mustReadFile(t,
		"examples/gastown/packs/gastown/assets/prompts/crew.template.md",
	)

	f.Files["/city/prompts/shared/approval-fallacy.template.md"] = mustReadFile(t,
		"examples/gastown/packs/gastown/template-fragments/approval-fallacy.template.md",
	)
	f.Files["/city/prompts/shared/propulsion.template.md"] = mustReadFile(t,
		"examples/gastown/packs/gastown/template-fragments/propulsion.template.md",
	)
	f.Files["/city/prompts/shared/capability-ledger.template.md"] = mustReadFile(t,
		"examples/gastown/packs/gastown/template-fragments/capability-ledger.template.md",
	)
	f.Files["/city/prompts/shared/architecture.template.md"] = mustReadFile(t,
		"examples/gastown/packs/gastown/template-fragments/architecture.template.md",
	)

	ctx := PromptContext{
		CityRoot:          "/city",
		AgentName:         "gastown/crew/test",
		RigName:           "gastown",
		WorkDir:           "/work/gastown/crew/test",
		IssuePrefix:       "gt",
		InstructionsFile:   "CLAUDE.md",
		ProviderKey:       "claude",
		ProviderDisplayName: "Claude Code",
		DefaultBranch:    "main",
		Branch:           "feature/test",
	}

	got := renderPrompt(f, "/city", "", "prompts/crew.template.md", ctx, "", io.Discard,
		[]string{"examples/gastown/packs/gastown"}, nil, nil)

	if !strings.Contains(got, "CLAUDE.md") {
		t.Errorf("CLAUDE.md not found in rendered output for claude provider")
	}

	if strings.Contains(got, "go test ./...") {
		t.Errorf("rendered output contains hardcoded 'go test ./...' - quality gate must use {{ .InstructionsFile }}")
	}

	if strings.Contains(got, "golangci-lint run") {
		t.Errorf("rendered output contains hardcoded 'golangci-lint run' - quality gate must use {{ .InstructionsFile }}")
	}
}
