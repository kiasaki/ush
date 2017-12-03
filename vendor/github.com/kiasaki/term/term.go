package term

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

type Terminal struct {
	in           *os.File
	inBuffer     *bufio.Reader
	out          *os.File
	runes        chan rune
	events       chan *Event
	running      bool
	originalMode *termios
	sigwinch     chan os.Signal
	Width        int
	Height       int
}

func NewTerminal() *Terminal {
	return &Terminal{}
}

func (t *Terminal) Start() error {
	var err error

	if t.in == nil {
		if t.in, err = os.OpenFile("/dev/tty", os.O_RDONLY, 0); err != nil {
			return err
		}
		t.inBuffer = bufio.NewReader(t.in)
	}
	if t.out == nil {
		if t.out, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0); err != nil {
			return err
		}
	}
	t.running = true

	fd := t.out.Fd()
	t.originalMode, err = fetchMode(fd)
	if err != nil {
		return err
	}
	mode := cloneMode(t.originalMode)
	mode.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	mode.Oflag &^= syscall.OPOST
	mode.Cflag |= syscall.CS8
	mode.Cflag &^= syscall.CSIZE | syscall.PARENB
	mode.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	err = setMode(fd, mode)
	if err != nil {
		return err
	}

	t.Width, t.Height, err = windowSize(t.out.Fd())
	if err != nil {
		return err
	}

	t.sigwinch = make(chan os.Signal, 1)
	signal.Notify(t.sigwinch, syscall.SIGWINCH)

	t.runes = make(chan rune, 500)
	t.events = make(chan *Event, 500)
	go t.readRunes()
	go t.readEvents()

	return nil
}

func (t *Terminal) Stop() {
	t.running = false

	close(t.runes)
	close(t.events)

	signal.Reset(syscall.SIGWINCH)

	t.Puts(termShowCursor)

	setMode(t.out.Fd(), t.originalMode)
}

type EventType int

const (
	EventKey EventType = iota
	EventResize
)

type Event struct {
	Type EventType
	Key  Key
	Rune rune
}

func (t *Terminal) readRunes() {
	for {
		r, _, err := t.inBuffer.ReadRune()
		if !t.running {
			return
		}
		if err != nil {
			panic(err) // TODO don't panic
		}
		t.runes <- r
	}
}

func (t *Terminal) readEvents() {
	for {
		select {
		case _, ok := <-t.sigwinch:
			if !ok {
				return
			}
			var err error
			t.Width, t.Height, err = windowSize(t.out.Fd())
			if err != nil {
				// TODO dont panic, send error event
				t.Stop()
				panic(err)
			}
			t.events <- &Event{Type: EventResize}
		case r, ok := <-t.runes:
			if !ok {
				return
			}
			seq := []rune{r}

			// Gather up a few more characters for 50ms
			timeout := time.After(50 * time.Millisecond)
		wait:
			for {
				select {
				case r := <-t.runes:
					seq = append(seq, r)
					continue wait
				case <-timeout:
					break wait
				}
			}

			i := 0
			for i < len(seq) {
				r := seq[i]
				if r == KeyEsc {
					if len(seq[i+1:]) >= 2 {
						key := KeyUnknown
						switch string(seq[i+1 : i+3]) {
						case "[A":
							key = KeyUp
						case "[B":
							key = KeyDown
						case "[C":
							key = KeyRight
						case "[D":
							key = KeyLeft
						case "[H":
							key = KeyHome
						case "[F":
							key = KeyEnd
						case "[Z":
							key = KeyShiftTab
						case "OH":
							key = KeyHome
						case "OF":
							key = KeyEnd
						}
						if key != KeyUnknown {
							t.events <- &Event{Type: EventKey, Key: key}
							i = i + 4
							continue
						}
					}
					if len(seq[i+1:]) >= 3 {
						key := KeyUnknown
						switch string(seq[i+1 : i+4]) {
						case "[3~":
							key = KeyDel
						case "[5~":
							key = KeyPageUp
						case "[6~":
							key = KeyPageDown
						}
						if key != KeyUnknown {
							t.events <- &Event{Type: EventKey, Key: key}
							i = i + 5
							continue
						}
					}

					r = rune(KeyEsc)
					goto normal
				}
				if r < 27 || r == 127 {
					t.events <- &Event{Type: EventKey, Key: Key(r)}
					i++
					continue
				}
			normal:
				t.events <- &Event{Type: EventKey, Key: KeyRune, Rune: r}
				i++
			}
		}
	}
}

func (t *Terminal) Events() chan *Event {
	return t.events
}

func (t *Terminal) Puts(f string, args ...interface{}) {
	fmt.Fprintf(t.out, f, args...)
}

const (
	termClear           string = "\x1b[H\x1b[J"
	termClearLine              = "\x1b[0K"
	termSetColor               = "\x1b[%d;%dm"
	termSetFgBg                = "\x1b[3%d;4%dm"
	termAttrOff                = "\x1b[0;10m"
	termReset                  = "\x1b[0m"
	termEnterAcs               = "\x1b[11m"
	termExitAcs                = "\x1b[10m"
	termEnterCA                = "\x1b[?1049h"
	termExitCA                 = "\x1b[r\x1b[?1049l"
	termHideCursor             = "\x1b[?25l"
	termShowCursor             = "\x1b[?25h"
	termSetCursor              = "\x1b[%d;%dH"
	termSetCursorColumn        = "\x1b[%dG"
)

func (t *Terminal) Clear() {
	t.Puts(termClear)
}

func (t *Terminal) ShowCursor() {
	t.Puts(termShowCursor)
}

func (t *Terminal) HideCursor() {
	t.Puts(termHideCursor)
}

func (t *Terminal) ClearLine() {
	t.Puts(termClearLine)
}

func (t *Terminal) SetCursor(x, y int) {
	t.Puts(termSetCursor, y+1, x+1)
}

func (t *Terminal) SetCursorColumn(x int) {
	t.Puts(termSetCursorColumn, x+1)
}

func (t *Terminal) setColor(c Color, isBg bool) {
	var brightValue = 0
	if c > 8 && c <= 16 {
		c -= 8
		brightValue = 1
	}
	if isBg {
		c += 40
	} else {
		c += 30
	}
	t.Puts(termSetColor, brightValue, c)
}
func (t *Terminal) SetFg(c Color) {
	t.setColor(c, false)
}

func (t *Terminal) SetBg(c Color) {
	t.setColor(c, true)
}

func (t *Terminal) Reset() {
	t.Puts(termReset)
}

type Color int

const (
	ColorBlack Color = iota
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorPurple
	ColorCyan
	ColorLightGray
	ColorDefault
)

const (
	ColorGray Color = iota + 9
	ColorBrightRed
	ColorBrightGreen
	ColorBrightYellow
	ColorBrightBlue
	ColorBrightPurple
	ColorBrightCyan
	ColorWhite
)

type Key int

const (
	KeyNull      Key = 0
	KeyCtrlA         = 1
	KeyCtrlB         = 2
	KeyCtrlC         = 3
	KeyCtrlD         = 4
	KeyCtrlE         = 5
	KeyCtrlF         = 6
	KeyCtrlG         = 7
	KeyCtrlH         = 8
	KeyTab           = 9
	KeyNl            = 10
	KeyCtrlK         = 11
	KeyCtrlL         = 12
	KeyCr            = 13
	KeyCtrlN         = 14
	KeyCtrlO         = 15
	KeyCtrlP         = 16
	KeyCtrlQ         = 17
	KeyCtrlR         = 18
	KeyCtrlS         = 19
	KeyCtrlT         = 20
	KeyCtrlU         = 21
	KeyCtrlV         = 22
	KeyCtrlW         = 23
	KeyCtrlX         = 24
	KeyCtrlY         = 25
	KeyCtrlZ         = 26
	KeyEsc           = 27
	KeyFs            = 28
	KeyGs            = 29
	KeyRs            = 30
	KeyUs            = 31
	KeyBackspace     = 127
)

const (
	KeyRune Key = iota + 1000
	KeyUnknown
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyHome
	KeyEnd
	KeyInsert
	KeyDel
	KeyPageUp
	KeyPageDown
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyAltB
	KeyAltF
	KeyAltY
	KeyShiftTab
	KeyWordLeft
	KeyWordRight
)

type termios struct {
	syscall.Termios
}

func cloneMode(mode *termios) *termios {
	clone := &termios{}
	clone.Iflag = mode.Iflag
	clone.Oflag = mode.Oflag
	clone.Cflag = mode.Cflag
	clone.Lflag = mode.Lflag
	clone.Cc = mode.Cc
	clone.Ispeed = mode.Ispeed
	clone.Ospeed = mode.Ospeed
	return clone
}

func setMode(fd uintptr, mode *termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, termSetTermios, uintptr(unsafe.Pointer(mode)))
	if errno != 0 {
		return fmt.Errorf("Got error number %d setting terminal mode", errno)
	}
	return nil
}

func fetchMode(fd uintptr) (*termios, error) {
	var mode termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, termGetTermios, uintptr(unsafe.Pointer(&mode)))
	if errno != 0 {
		return nil, fmt.Errorf("Got error number %d fetching terminal mode", errno)
	}
	return &mode, nil
}

func windowSize(fd uintptr) (int, int, error) {
	dim := [4]uint16{}
	dimp := uintptr(unsafe.Pointer(&dim))
	ioc := uintptr(syscall.TIOCGWINSZ)
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL,
		fd, ioc, dimp, 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dim[1]), int(dim[0]), nil
}
