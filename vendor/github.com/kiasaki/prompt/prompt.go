package prompt

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/kiasaki/term"
)

// ErrorPromptAborted is returned by `Prompt()` when CtrlC is pressed
var ErrorPromptAborted = errors.New("Prompt aborted")

// ErrorPromptEnded is returned by `Prompt()` when CtrlD is pressed
var ErrorPromptEnded = errors.New("Prompt ended")

// Prompt represents a current instance of a prompt and it's associated
// history and completion callbacks
type Prompt struct {
	in           *bufio.Reader
	history      []string
	terminal     *term.Terminal
	completionFn func(string) []string
}

// NewPrompt creates a new instance of a prompt
func NewPrompt() *Prompt {
	return &Prompt{
		in:           bufio.NewReader(os.Stdin),
		history:      []string{},
		terminal:     term.NewTerminal(),
		completionFn: nil,
	}
}

// SetCompletionFn sets the auto completion callback function
func (p *Prompt) SetCompletionFn(fn func(string) []string) {
	p.completionFn = fn
}

// History returns the current history as a `\n` separated string
func (p *Prompt) History() string {
	return strings.Join(p.history, "\n")
}

// LoadHistory replaces the current history with the contents of the
// provided `\n` separated string
func (p *Prompt) LoadHistory(history string) {
	buf := ""
	for r := range history {
		if r == '\n' {
			p.history = append(p.history, buf)
			buf = ""
		}
		buf += string(r)
	}
}

// AppendHistory adds a line to the prompt's history
func (p *Prompt) AppendHistory(line string) {
	p.history = append(p.history, line)
}

// Prompt puts the terminal in line editing mode and waits for the user to
// enter some text followed by the `Enter` key.
func (p *Prompt) Prompt(prompt string) (string, error) {
	p.terminal.Start()
	defer p.terminal.Stop()
	p.terminal.Puts(prompt)

	state := promptState{
		p:                 p,
		prompt:            prompt,
		line:              "",
		historyIndex:      -1,
		completionOptions: nil,
	}
	for {
		ev := <-p.terminal.Events()
		if ev.Type == term.EventKey {
			if ev.Key == term.KeyCtrlC {
				return "", ErrorPromptAborted
			}
			if ev.Key == term.KeyCtrlD {
				return "", ErrorPromptEnded
			}
			if ev.Key == term.KeyCr {
				state.MaybeUseHistory()
				p.terminal.Puts("\n")
				p.terminal.SetCursorColumn(0)
				break
			}
			if ev.Key == term.KeyBackspace {
				state.MaybeUseHistory()
				if len(state.line) > 0 {
					state.Edit(func(line string) string {
						return line[:len(line)-1]
					})
				}
			} else if ev.Key == term.KeyTab {
				state.CompleteSuggest()
			} else if ev.Key == term.KeyUp {
				state.HistoryPrevious()
			} else if ev.Key == term.KeyDown {
				state.HistoryNext()
			} else if ev.Key == term.KeyCtrlU {
				state.Edit(func(line string) string {
					return ""
				})
			} else if ev.Key == term.KeyCtrlL {
				p.terminal.Clear()
				state.Render()
			} else if ev.Key == term.KeyRune {
				state.Edit(func(line string) string {
					return line + string(ev.Rune)
				})
			} else {
				// TODO remove debug
				// p.terminal.Puts("\n")
				// p.terminal.SetCursorColumn(0)
				// p.terminal.Puts("unknown key: ", int(ev.Key), ev.Rune)
			}
		}
	}

	return string(state.line), nil
}

type promptState struct {
	p                 *Prompt
	prompt            string
	line              string
	historyIndex      int
	completionOptions []string
}

func (s *promptState) Render() {
	s.p.terminal.SetCursorColumn(0)
	s.p.terminal.Puts(strings.Repeat(" ", s.p.terminal.Width))
	s.p.terminal.SetCursorColumn(0)
	if s.historyIndex == -1 {
		s.p.terminal.Puts(s.prompt + s.line)
	} else {
		s.p.terminal.Puts(s.prompt + s.p.history[s.historyIndex])
	}
}

func (s *promptState) RenderCompletionOptions() {
	t := s.p.terminal
	t.Puts("\n")
	t.SetCursorColumn(0)
	for _, option := range s.completionOptions {
		t.Puts(option + "\n")
		t.SetCursorColumn(0)
	}
}

func (s *promptState) MaybeUseHistory() {
	if s.historyIndex != -1 {
		s.line = s.p.history[s.historyIndex]
		s.historyIndex = -1
	}
}

func (s *promptState) Edit(cb func(string) string) {
	s.MaybeUseHistory()

	// Reset completion options so that they get recomputed
	s.completionOptions = nil

	s.line = cb(s.line)

	s.Render()
}

func (s *promptState) HistoryPrevious() {
	if s.historyIndex == -1 {
		if len(s.p.history) > 0 {
			s.historyIndex = len(s.p.history) - 1
		}
	} else if s.historyIndex > 0 {
		s.historyIndex--
	}
	s.Render()
}

func (s *promptState) HistoryNext() {
	if s.historyIndex == len(s.p.history)-1 || s.historyIndex == -1 {
		s.historyIndex = -1
	} else {
		s.historyIndex++
	}
	s.Render()
}

func (s *promptState) CompleteSuggest() {
	if s.p.completionFn == nil {
		return
	}

	// We just completed, list all completion options
	if s.completionOptions != nil {
		s.RenderCompletionOptions()
		s.Render()
		return
	}

	// Try getting auto completion options
	s.completionOptions = s.p.completionFn(s.line)
	if s.completionOptions == nil || len(s.completionOptions) == 0 {
		// Abort if no options were returned
		s.completionOptions = nil
		return
	}

	s.MaybeUseHistory()

	// Find the common root ([]string{"blue", "black"} => "bl")
	bestOption := s.completionOptions[0]
	for _, option := range s.completionOptions[1:] {
		for i, c := range option {
			if i >= len(bestOption) {
				break
			}
			if c != []rune(bestOption)[i] {
				bestOption = option[:i]
				break
			}
		}
	}

	// If we autocompleted but there was more than 1 option,
	// show completion options right away
	if len(s.completionOptions) > 1 {
		s.RenderCompletionOptions()
	}

	// Don't keep the suggestions around if we just changed the promt line
	// to enable chaining `Tab` with different suggestions
	if s.line != bestOption {
		s.completionOptions = nil
	}

	// Suggest it
	s.line = bestOption
	s.Render()
}
