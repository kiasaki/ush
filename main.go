package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kiasaki/prompt"
	"github.com/kiasaki/ush/parser"
)

var ushVersion = "devel"

var varRegexp = regexp.MustCompile(`\$[a-zA-Z_]+`)

var currentCmd *exec.Cmd
var currentState *State

type State struct {
	Cwd             string
	IsInteractive   bool
	Aliases         map[string]string
	prompt          *prompt.Prompt
	configFileName  string
	historyFileName string
	history         string
}

func NewState() *State {
	s := &State{
		Cwd:             "/",
		IsInteractive:   false,
		Aliases:         map[string]string{},
		prompt:          prompt.NewPrompt(),
		configFileName:  "",
		historyFileName: "",
		history:         "",
	}
	s.prompt.SetCompletionFn(s.defaultAutocomplete)

	if cwd, err := os.Getwd(); err != nil {
		s.ReportError("error reading current directory")
	} else {
		s.Cwd = cwd
	}

	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		s.configFileName = filepath.Join(homeDir, ".ushrc")
		s.historyFileName = filepath.Join(homeDir, ".ush_history")
	} else {
		s.historyFileName = filepath.Join(os.TempDir(), ".ush_history")
	}

	if _, err := os.Stat(s.historyFileName); err == nil {
		if contents, err := ioutil.ReadFile(s.historyFileName); err != nil {
			s.ReportError("error reading history file")
		} else {
			s.prompt.LoadHistory(string(contents))
		}
	}

	return s
}

func (s *State) Quit(statusCode int) {
	// Save history to disk
	history := []byte(s.prompt.History())
	if err := ioutil.WriteFile(s.historyFileName, history, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "ush: error writing history file")
	}

	os.Exit(statusCode)
}

func (s *State) ReportError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ush: "+format+"\n", args...)
	if !s.IsInteractive {
		s.Quit(1)
	}
}

func (s *State) defaultAutocomplete(line string) []string {
	// Parse current line
	pipes := s.ParseLine(line)
	parts := []string{}
	for i := range pipes {
		if i > 0 {
			parts = append(parts, "|")
		}
		parts = append(parts, pipes[i]...)
	}

	// If we didn't write anything yet well have problem indexing later, do the simple case
	if len(parts) == 0 {
		if suggestions, err := filepath.Glob("*"); err != nil {
			return []string{}
		} else {
			return suggestions
		}
	}

	// Autocomplete line's last part
	if suggestions, err := filepath.Glob(parts[len(parts)-1] + "**"); err != nil {
		return []string{}
	} else {
		for i, s := range suggestions {
			suggestions[i] = strings.Join(parts[:len(parts)-1], " ") + " " + s
			suggestions[i] = strings.TrimLeft(suggestions[i], " ")
			if isDir(s) {
				suggestions[i] += string(os.PathSeparator)
			} else {
				suggestions[i] += " "
			}
		}
		return suggestions
	}
}

func (s *State) ParseLine(line string) [][]string {
	words, err := parser.Parse(line)
	if err != nil {
		s.ReportError("error parsing line [%s] %v", line, err)
	}

	commands := make([][]string, 0)
	cmd := make([]string, 0)

	for _, word := range words {
		trimed := strings.TrimSpace(word)

		// Handle comments
		if len(trimed) > 0 && trimed[0] == '#' {
			break
		}

		// Handle ~ as $HOME
		if len(trimed) > 0 && trimed[0] == '~' {
			trimed = filepath.Join(os.Getenv("HOME"), trimed[1:])
		}

		// Expand/Replace variables
		trimed = varRegexp.ReplaceAllStringFunc(trimed, func(match string) string {
			return os.Getenv(match[1:])
		})

		if len(trimed) > 0 && trimed[0] == '*' {
			// Handle basic glob
			expandeds, err := filepath.Glob(trimed)
			if err != nil {
				log.Fatal(err)
			}
			cmd = append(cmd, expandeds...)
		} else if trimed == "|" {
			commands = append(commands, cmd)
			cmd = make([]string, 0)
		} else {
			cmd = append(cmd, trimed)
		}
	}

	if len(cmd) > 0 {
		commands = append(commands, cmd)
	}

	return commands
}

// {{{ Execute
func commandErrorExitCode(err error) (int, bool) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), true
		}
	}
	return 0, false
}

func (s *State) createSubprocess(command []string, in *io.PipeReader, out *io.PipeWriter, ch chan<- bool) {
	go func() {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if in != nil {
			cmd.Stdin = in
		}
		if out != nil {
			cmd.Stdout = out
		}

		err := cmd.Start()
		if err != nil {
			s.ReportError("error running [%s] %v", parser.Format(command...), err)
		} else {
			currentCmd = cmd // Set global for signal fowarding
			err = cmd.Wait()
			if err != nil {
				if statusCode, ok := commandErrorExitCode(err); ok {
					os.Setenv("exit", strconv.Itoa(statusCode))
				} else {
					s.ReportError("error running [%s] %v", parser.Format(command...), err)
				}
			}
		}
		if in != nil {
			in.Close()
		}
		if out != nil {
			out.Close()
		}

		ch <- true
	}()
}

type DataPipes struct {
	in  *io.PipeReader
	out *io.PipeWriter
}

func makeSubprocessPipes(processes int) []*DataPipes {
	pipes := make([]*DataPipes, 0)
	for i := 0; i < processes; i++ {
		in, out := io.Pipe()
		data := &DataPipes{in, out}
		pipes = append(pipes, data)
	}
	return pipes
}

func waitSubprocess(processes int, ch <-chan bool) {
	for i := 0; i < processes; i++ {
		<-ch
	}
}

func (s *State) Execute(commands [][]string) {
	if len(commands) == 0 || len(commands[0]) == 0 {
		return // Skip empty commands
	}

	arg0 := commands[0][0]

	// Handle builtins
	if arg0 == "exit" {
		s.BuiltinExit(commands[0])
		return
	} else if arg0 == "help" {
		s.BuiltinHelp(commands[0])
		return
	} else if arg0 == "exec" {
		s.BuiltinExec(commands[0])
		return
	} else if arg0 == "cd" {
		s.BuiltinCd(commands[0])
		return
	} else if arg0 == "set" {
		s.BuiltinSet(commands[0])
		return
	} else if arg0 == "unset" {
		s.BuiltinUnset(commands[0])
		return
	} else if arg0 == "alias" {
		s.BuiltinAlias(commands[0])
		return
	} else if arg0 == "source" {
		s.BuiltinSource(commands[0])
		return
	}

	// Replace aliases with aliased commands
	if value, ok := s.Aliases[arg0]; ok {
		commandRest := commands[0][1:]
		commandsRest := commands[1:]
		commands = s.ParseLine(value)
		commands[len(commands)-1] = append(commands[len(commands)-1], commandRest...)
		commands = append(commands, commandsRest...)
	}

	processCount := len(commands)
	pipes := makeSubprocessPipes(processCount)
	ch := make(chan bool, len(commands))
	for i, command := range commands {
		var in *io.PipeReader
		var out *io.PipeWriter

		if i != 0 {
			in = pipes[i-1].in
		}
		if i != len(commands)-1 {
			out = pipes[i].out
		}

		s.createSubprocess(command, in, out, ch)
	}

	waitSubprocess(processCount, ch)
}

// }}}

func (s *State) ExecuteLine(line string) {
	s.Execute(s.ParseLine(line))
}

func (s *State) ExecuteFile(fileName string) {
	if contents, err := ioutil.ReadFile(fileName); err != nil {
		s.ReportError("errror reading file: %v", fileName)
	} else {
		for _, line := range strings.Split(string(contents), "\n") {
			s.ExecuteLine(line)
		}
	}
}

func (s *State) BuiltinExit(args []string) {
	s.Quit(0)
}

func (s *State) BuiltinHelp(args []string) {
	fmt.Fprintf(os.Stderr, `ush: a shell with a microscopic feature set

Args

  -v --version  Show ush's version
  -h --help     Show this message
  -c            Run the following command and exit

Builtins

  help    Show this message
  exit    Exit the shell
  exec    Replaces shell with new process
  cd      Change the current directory
  set     Set an environment variable's value
  unset   Delete an environment variable
  alias   Register an alias for a command
  source  Load and execute a file

`)
}

func (s *State) BuiltinExec(args []string) {
	if len(args) <= 1 {
		s.ReportError("exec needs at least 1 argument")
		return
	}
	err := syscall.Exec(args[1], args[1:], os.Environ())
	if err != nil {
		s.ReportError("error calling exec: %v: %v", args, err.Error())
	}
}

func (s *State) BuiltinCd(args []string) {
	var err error
	if len(args) > 1 {
		err = os.Chdir(args[1])
	} else {
		err = os.Chdir(os.Getenv("HOME"))
	}
	if err != nil {
		s.ReportError("error changing directory %v", err)
	}

	if cwd, err := os.Getwd(); err != nil {
		s.ReportError("error getting current directory %v", err)
	} else {
		s.Cwd = cwd
	}
}

func (s *State) BuiltinSet(args []string) {
	if len(args) != 3 {
		s.ReportError("set needs 2 arguments, got [%s]", parser.Format(args...))
		return
	}
	os.Setenv(args[1], args[2])
}

func (s *State) BuiltinUnset(args []string) {
	if len(args) != 2 {
		s.ReportError("unset needs 1 argument, got [%s]", parser.Format(args...))
		return
	}
	os.Unsetenv(args[1])
}

func (s *State) BuiltinAlias(args []string) {
	if len(args) != 3 {
		s.ReportError("alias needs 2 arguments, got [%s]", parser.Format(args...))
		return
	}
	s.Aliases[args[1]] = args[2]
}

func (s *State) BuiltinSource(args []string) {
	if len(args) != 2 {
		s.ReportError("source needs 1 argument, got [%s]", parser.Format(args...))
		return
	}
	s.ExecuteFile(args[1])
}

// Returns if a file is a directory, returning false in case of any error
func isDir(fileName string) bool {
	if fileInfo, err := os.Stat(fileName); err == nil {
		return fileInfo.IsDir()
	}
	return false
}

func init() {
	// Set $SHELL
	ex, err := os.Executable()
	if err == nil {
		os.Setenv("SHELL", ex)
	}

	// Forward signals to running commands
	sigc := make(chan os.Signal, 5)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for {
			select {
			case s := <-sigc:
				if currentCmd != nil {
					currentCmd.Process.Signal(s)
				} else if s == syscall.SIGTERM {
					currentState.ReportError("got SIGTERM, exiting")
					currentState.Quit(1)
				} else if s == syscall.SIGQUIT {
					currentState.ReportError("got SIGQUIT, exiting")
					currentState.Quit(1)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

func main() {
	currentState = NewState()
	s := currentState

	// Execute ~/.ushrc
	if s.configFileName != "" {
		if _, err := os.Stat(s.configFileName); err == nil {
			s.ExecuteFile(s.configFileName)
		}
	}

	// Handle args
	ranFile := false
	for i, arg := range os.Args[1:] {
		if arg == "-v" || arg == "-V" || arg == "--version" || arg == "version" {
			fmt.Fprintf(os.Stderr, "ush version %s\n", ushVersion)
			s.Quit(0)
		}
		if arg == "-h" || arg == "-H" || arg == "--help" || arg == "help" {
			s.BuiltinHelp([]string{})
			s.Quit(0)
		}
		if arg == "-c" {
			command := os.Args[i+2:]
			if len(command) == 0 {
				s.ReportError("called with '-c' but missing a command")
				return
			}
			s.Execute([][]string{command})
			s.Quit(0)
			return
		}
		if arg[0] == '-' {
			// ignore unknown args starting with - or --
			continue
		}
		if _, err := os.Stat(arg); err == nil {
			ranFile = true
			s.ExecuteFile(arg)
		} else {
			s.ReportError("\"%s\" is not a file", arg)
			return
		}
	}
	if ranFile {
		s.Quit(0) // Exit before starting interactive more if we ran a file
	}

	s.IsInteractive = true

	// Main interactive loop
	for {
		promptLine := filepath.Base(s.Cwd) + "$ "
		if line, err := s.prompt.Prompt(promptLine); err == nil {
			s.prompt.AppendHistory(line)
			s.ExecuteLine(line)
		} else if err == prompt.ErrorPromptAborted || err == prompt.ErrorPromptEnded {
			fmt.Println()
			continue
		} else {
			s.ReportError("error reading line: %v", err)
			s.Quit(1)
		}
	}
}
