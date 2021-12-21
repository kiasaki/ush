# term

_A minimal terminal package in pure go._

### intro

`term` is small (< 1000 lines) terminal interaction library. It compromises a
lot on portability and compatibility with various terminals but in exchange it's
small enough for you to read in 15 minutes or embed in a bigger project and
modify yourself to fit your needs.

### features

- Settings the terminal in raw mode
- Reading basic keys and escape code from stdin
- Setting simple terminal attributes (fg, bg, bold, ...)
- Basic terminal actions (moving cursor, clearing the screen, ...)
- Reading and reacting to the terminal size
- Simple API for handling events and keypresses

### api

```
type Key int
type Color int
type EventType iny
type Event struct {
  Type EventType
  Key Key
  Rune rune
}
type Terminal struct {
  Width int
  Height int
}
func NewTerminal() *Terminal
func (t *Terminal) Start() error
func (t *Terminal) Stop()
func (t *Terminal) Events() chan *Event
func (t *Terminal) Puts(f string, args ...interface{})
func (t *Terminal) Clear()
func (t *Terminal) ClearLine()
func (t *Terminal) SetFg(c Color)
func (t *Terminal) SetBg(c Color)
func (t *Terminal) Reset()
func (t *Terminal) ShowCursor()
func (t *Terminal) HideCursor()
func (t *Terminal) SetCursor(x, y int)
func (t *Terminal) SetCursorColumn(x int)
```

### license

MIT. See `LICENSE` file.
