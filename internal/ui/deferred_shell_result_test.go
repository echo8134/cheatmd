package ui

import (
	"errors"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func deliverOverlayCommands(m *mainModel, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, next := range batch {
			deliverOverlayCommands(m, next)
		}
		return
	}
	_, next := m.Update(msg)
	deliverOverlayCommands(m, next)
}

func TestDeferredShellResultPreservesOverlayInput(t *testing.T) {
	for _, result := range []struct {
		name string
		msg  shellResultMsg
	}{
		{"multiple", shellResultMsg{options: []string{"chosen", "other"}}},
		{"single", shellResultMsg{options: []string{"default"}}},
		{"empty", shellResultMsg{}},
		{"error", shellResultMsg{err: errors.New("options unavailable")}},
	} {
		for _, action := range []string{"accept", "cancel", "preview"} {
			for _, input := range []string{"chosen", ""} {
				t.Run(result.name+"/"+action+"/"+input, func(t *testing.T) {
					m, index := setupTestModel()
					m.selected = index.Cheats[0]
					m.startVarResolution()
					m.varState.vars[0].prefill = "prefilled"
					m.varState.isPromptOnly = false
					m.textInput.SetValue(input)
					key := tea.KeyEsc
					if action == "preview" {
						m.enterPreview(m.selected)
					} else {
						m.subState = &substituteSearchState{
							picker:    NewPicker([]substituteOption{{Value: input}}, nil),
							prevInput: input,
						}
						m.phase = phaseSubstituteSearch
						m.textInput.SetValue("overlay query")
						if action == "accept" {
							key = tea.KeyEnter
						}
					}
					overlayInput := m.textInput.Value()
					msg := result.msg
					msg.generation = m.shellGeneration
					m.Update(msg)
					if m.textInput.Value() != overlayInput || m.varState.options != nil || m.varState.shellErr != nil {
						t.Fatal("shell result changed state while overlay was open")
					}
					_, cmd := m.Update(tea.KeyMsg{Type: key})
					deliverOverlayCommands(m, cmd)
					if !reflect.DeepEqual(m.varState.options, msg.options) || m.varState.shellErr != msg.err {
						t.Fatalf("result lost on overlay close: options=%v error=%v, want options=%v error=%v", m.varState.options, m.varState.shellErr, msg.options, msg.err)
					}
					if m.phase != phaseVarResolve || m.textInput.Value() != input {
						t.Fatalf("restored input = %q in phase %v, want %q", m.textInput.Value(), m.phase, input)
					}
					if m.varState.isPromptOnly != (len(msg.options) < 2) {
						t.Fatalf("wrong prompt mode after deferred result")
					}
					if len(msg.options) > 1 && input != "" && len(m.varState.picker.Filtered) != 1 {
						t.Fatalf("options were not filtered using restored input: %v", m.varState.picker.Filtered)
					}
				})
			}
		}
	}
}

func TestDeferredShellResultCannotReplaceNewPrompt(t *testing.T) {
	m, index := setupTestModel()
	m.selected = index.Cheats[0]
	m.startVarResolution()
	m.enterPreview(m.selected)
	m.Update(shellResultMsg{options: []string{"old", "older"}, generation: m.shellGeneration})
	_, deferred := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m.prepareCurrentVar()
	m.Update(shellResultMsg{options: []string{"current", "new"}, generation: m.shellGeneration})
	deliverOverlayCommands(m, deferred)
	if !reflect.DeepEqual(m.varState.options, []string{"current", "new"}) {
		t.Fatalf("old deferred result replaced new prompt options: %v", m.varState.options)
	}
}
