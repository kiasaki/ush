package main

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

var ErrorPromptAborted = errors.New("Prompt aborted")

type Prompt struct {
	in       *bufio.Reader
	history  []string
	terminal *Terminal
}

func NewPrompt() *Prompt {
	return &Prompt{
		in:       bufio.NewReader(os.Stdin),
		history:  []string{},
		terminal: NewTerminal(),
	}
}

func (p *Prompt) Start() {
}

func (p *Prompt) Stop() {
}

func (p *Prompt) History() string {
	return strings.Join(p.history, "\n")
}

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

func (p *Prompt) AppendHistory(line string) {
	p.history = append(p.history, line)
}

func (p *Prompt) Prompt(prompt string) (string, error) {
	p.terminal.Start()
	defer p.terminal.Stop()
	//p.terminal.Clear()
	//p.terminal.SetCursor(0, 0)
	p.terminal.Puts(prompt)

	line := ""
	for {
		ev := <-p.terminal.Events()
		if ev.Type == EventKey {
			if ev.Key == KeyCtrlC || ev.Key == KeyCtrlD {
				return "", ErrorPromptAborted
			}
			if ev.Key == KeyCr {
				p.terminal.Puts("\n")
				p.terminal.SetCursorColumn(0)
				break
			}
			if ev.Key == KeyBackspace {
				if len(line) > 0 {
					line = line[:len(line)-1]
				}
			} else if ev.Key == KeyCtrlU {
				line = ""
			} else if ev.Key == KeyCtrlL {
				p.terminal.Clear()
			} else if ev.Key == KeyRune {
				line += string(ev.Rune)
			} else {
				// TODO remove debug
				p.terminal.Puts("\n")
				p.terminal.SetCursorColumn(0)
				p.terminal.Puts("unknown key: ", int(ev.Key), ev.Rune)
			}
		}
		p.terminal.SetCursorColumn(0)
		p.terminal.Puts(strings.Repeat(" ", p.terminal.Width))
		p.terminal.SetCursorColumn(0)
		p.terminal.Puts(prompt + line)
	}

	return string(line), nil
}
