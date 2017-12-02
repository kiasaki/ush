package main

import (
	"bufio"
	"errors"
	"io"
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
	p.terminal.Start()
}

func (p *Prompt) Stop() {
	p.terminal.Stop()
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
	os.Stdout.Write([]byte(prompt))
	line, _, err := p.in.ReadLine()
	if err == io.EOF {
		return "", ErrorPromptAborted
	}
	for r := range line {
		if r == KeyCtrlC {
			return "", ErrorPromptAborted
		}
		if r == KeyCtrlD {
			return "", ErrorPromptAborted
		}
	}
	return string(line), err
}
