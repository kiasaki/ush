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

	"github.com/kballard/go-shellquote"
	"github.com/peterh/liner"
)

var varRegexp = regexp.MustCompile(`\$[a-zA-Z_]+`)
var globRegexp = regexp.MustCompile(`\*`)

func parseLine(line string) [][]string {
	words, err := shellquote.Split(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ush: error parsing line [%s] %v\n", line, err)
		quit(1)
	}

	commands := make([][]string, 0)
	cmd := make([]string, 0)

	for _, word := range words {
		trimed := strings.TrimSpace(word)

		// Handle comments
		if len(trimed) > 0 && trimed[0] == '#' {
			break
		}

		if len(trimed) > 0 && trimed[0] == '~' {
			trimed = filepath.Join(os.Getenv("HOME"), trimed[1:])
		}

		// Expand/Replace variables
		trimed = varRegexp.ReplaceAllStringFunc(trimed, func(match string) string {
			return os.Getenv(match[1:])
		})

		if len(trimed) > 0 && trimed[0] == '*' {
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

func createSubprocess(command []string, in *io.PipeReader, out *io.PipeWriter, ch chan<- bool) {
	go func() {
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Stderr = os.Stderr

		if in == nil {
			cmd.Stdin = os.Stdout
		} else {
			cmd.Stdin = in
		}

		if out == nil {
			cmd.Stdout = os.Stdout
		} else {
			cmd.Stdout = out
		}

		err := cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ush: error running [%v] %v\n", strings.Join(command, " "), err)
		} else {
			currentCmd = cmd
			err = cmd.Wait()
			if err != nil {
				if exiterr, ok := err.(*exec.ExitError); ok {
					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
						os.Setenv("exit", strconv.Itoa(status.ExitStatus()))
					} else {
						fmt.Fprintf(os.Stderr, "ush: error running [%v] %v\n", strings.Join(command, " "), err)
					}
				} else {
					fmt.Fprintf(os.Stderr, "ush: error running [%v] %v\n", strings.Join(command, " "), err)
				}
			}

			if in != nil {
				in.Close()
			}

			if out != nil {
				out.Close()
			}
		}

		ch <- true
	}()
}

func waitSubprocess(processes int, ch <-chan bool) {
	for i := 0; i < processes; i++ {
		<-ch
	}
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

func executeCommand(commands [][]string) {
	if len(commands) == 0 || len(commands[0]) == 0 {
		return
	}

	if commands[0][0] == "exit" {
		quit(0)
	} else if commands[0][0] == "cd" {
		var err error
		if len(commands[0]) > 1 {
			err = os.Chdir(commands[0][1])
		} else {
			err = os.Chdir(os.Getenv("HOME"))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "ush: error changing directory %v\n", err)
			if !isInteractive {
				quit(1)
			}
		}
		cwd, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ush: error getting current directory %v\n", err)
			if !isInteractive {
				quit(1)
			}
		}
		return
	} else if commands[0][0] == "set" {
		if len(commands[0]) != 3 {
			fmt.Fprintf(os.Stderr, "ush: set needs 2 arguments, got [%s]\n", shellquote.Join(commands[0]...))
			return
		}
		os.Setenv(commands[0][1], commands[0][2])
		return
	} else if commands[0][0] == "unset" {
		if len(commands[0]) != 2 {
			fmt.Fprintf(os.Stderr, "ush: unset needs 1 argument\n")
			return
		}
		os.Unsetenv(commands[0][1])
		return
	} else if commands[0][0] == "alias" {
		if len(commands[0]) != 3 {
			fmt.Fprintf(os.Stderr, "ush: alias needs 2 arguments, got [%s]\n", shellquote.Join(commands[0]...))
			return
		}
		aliases[commands[0][1]] = commands[0][2]
		return
	} else if commands[0][0] == "source" {
		executeCommandFile(commands[0][1])
		return
	}

	// Handle aliases
	if value, ok := aliases[commands[0][0]]; ok {
		commandRest := commands[0][1:]
		commandsRest := commands[1:]
		commands = parseLine(value)
		commands[len(commands)-1] = append(commands[len(commands)-1], commandRest...)
		commands = append(commands, commandsRest...)
	}

	processes := len(commands)
	pipes := makeSubprocessPipes(processes)

	// Stop line to reset terminal input settings
	if line != nil {
		line.Close()
	}

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

		createSubprocess(command, in, out, ch)
	}

	waitSubprocess(processes, ch)

	// Restart liner
	initLiner()
}

func executeCommandText(text string) {
	commandLine := parseLine(text)
	executeCommand(commandLine)
}

func executeCommandFile(fileName string) {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ush: error reading file: %v\n%v\n", fileName, err)
		if !isInteractive {
			quit(1)
		}
	}

	for _, line := range strings.Split(string(contents), "\n") {
		if strings.TrimSpace(line) != "" {
			executeCommandText(line)
		}
	}
}

var line *liner.State
var isInteractive = true
var cwd string
var currentCmd *exec.Cmd
var configFileName string
var historyFileName string
var historyFile *os.File
var aliases = map[string]string{}

func init() {
	// Set $SHELL
	ex, err := os.Executable()
	if err == nil {
		os.Setenv("SHELL", ex)
	}

	// Forward signals to running commands
	sigc := make(chan os.Signal, 5)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for {
			select {
			case s := <-sigc:
				if currentCmd != nil {
					currentCmd.Process.Signal(s)
				} else if s == syscall.SIGTERM {
					fmt.Fprintf(os.Stderr, "ush: got SIGTERM, exiting\n")
					quit(0)
				} else if s == syscall.SIGQUIT {
					fmt.Fprintf(os.Stderr, "ush: got SIGQUIT, exiting\n")
					quit(0)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	cwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ush: can't get working directory\n")
		quit(1)
	}

	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		configFileName = filepath.Join(homeDir, ".ushrc")
		historyFileName = filepath.Join(homeDir, ".ush_history")
	} else {
		historyFileName = filepath.Join(os.TempDir(), ".ush_history")
	}
}

func autocomplete(line string) []string {
	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) == 0 {
		suggestions, err := filepath.Glob("*")
		if err != nil {
			return []string{}
		}
		return suggestions
	}
	suggestions, err := filepath.Glob(parts[len(parts)-1] + "**")
	if err != nil {
		return []string{}
	}
	for i, s := range suggestions {
		suggestions[i] = line[:len(line)-len(parts[len(parts)-1])] + s
	}
	return suggestions
}

func initLiner() *liner.State {
	line = liner.NewLiner()
	line.SetCompleter(autocomplete)

	if f, err := os.Open(historyFileName); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	if f, err := os.Create(historyFileName); err != nil {
		fmt.Fprintf(os.Stderr, "ush: error creating history file: %v\n", err)
		quit(1)
	} else {
		historyFile = f
	}

	return line
}

func historyAppend(line *liner.State, input string) {
	line.AppendHistory(input)
	line.WriteHistory(historyFile)
}

func quit(code int) {
	historyFile.Close()
	os.Exit(code)
}

func main() {
	if configFileName != "" {
		executeCommandFile(configFileName)
	}

	if len(os.Args) > 1 {
		isInteractive = false
		executeCommandFile(os.Args[1])
		quit(0)
	}

	initLiner()
	defer func() {
		line.Close()
	}()

	for {
		prompt := filepath.Base(cwd) + "$ "
		if input, err := line.Prompt(prompt); err == nil {
			historyAppend(line, input)
			executeCommandText(input)
		} else if err == liner.ErrPromptAborted {
			continue
		} else {
			fmt.Printf("ush: error reading line: %v\n", err)
		}
	}
}
