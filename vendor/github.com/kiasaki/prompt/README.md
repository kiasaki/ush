# prompt

_A simplistic command line prompt library in pure Go._

### intro

`prompt` is small (< 250 lines) command line prompt library. It does not have
all the fancy support and keybindings and config files that `readline` has but
it's still very functional. If you where looking for something light or small
enough to embed in a bigger application and modify to your own needs this is it.

### features

- Showing a custom prompt message
- Basic line editing and keybindings (<kbd>Backspace</kbd>, <kbd>Ctrl-U</kbd>, <kbd>Ctrl-L</kbd>, ...)
- History loading and exporting support + <kbd>Up</kbd>/<kbd>Down</kbd> keybindings
- Simple completion callback bound to <kbd>Tab</kbd>

### keys

- <kbd>Any Character</kbd> Gets added to the prompt line
- <kbd>Ctrl-C</kbd> Returns `ErrorPromptAborted`
- <kbd>Ctrl-D</kbd> Returns `ErrorPromptEnded`
- <kbd>Enter</kbd> Returns the entered prompt line
- <kbd>Backspace</kbd> Erases the last character in the line
- <kbd>Tab</kbd> Calls the `completionFn` if configured and auto-completes
- <kbd>Up</kbd> Selects previous history entry
- <kbd>Down</kbd> Shows next history entry
- <kbd>Ctrl-U</kbd> Erases the whole line
- <kbd>Ctrl-L</kbd> Clears the terminal screen

### api

```
var ErrorPromptAborted = errors.New("Prompt aborted")
var ErrorPromptEnded = errors.New("Prompt ended")
func NewPrompt() *Prompt
func (p *Prompt) SetCompletionFn(func(string) []string)
func (p *Prompt) History() string
func (p *Prompt) LoadHistory(history string)
func (p *Prompt) AppendHistory(line string)
func (p *Prompt) Prompt(prompt string) (string, error)
```

### license

MIT. See `LICENSE` file.
