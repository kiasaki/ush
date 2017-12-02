package main

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
	in           *bufio.Reader
	out          *os.File
	runes        chan rune
	events       chan *Event
	originalMode *termios
	sigwinch     chan os.Signal
}

func NewTerminal() *Terminal {
	return &Terminal{}
}

func (t *Terminal) Start() error {
	var err error
	if in, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0); err != nil {
		return err
	} else {
		t.in = bufio.NewReader(in)
	}
	if t.out, err = os.OpenFile("/dev/tty", os.O_WRONLY, 0); err != nil {
		return err
	}

	fd := t.out.Fd()
	t.originalMode, err = fetchMode(fd)
	if err != nil {
		return err
	}
	mode := cloneMode(t.originalMode)
	// No break, No CR to NL, No parity check, No strip char, No start/stop
	// output control
	mode.Iflag &^= icrnl | inpck | istrip | ixon
	// Disable post-processing
	mode.Oflag &^= opost
	// 8 bit chars
	mode.Cflag |= cs8
	// Canonical off, Echoing off, No Extendend functions, No signal chars (^C)
	mode.Lflag &^= syscall.ECHO | icanon | iexten | isig
	// Return each byte, or 1 for blocking
	mode.Cc[syscall.VMIN] = 1
	mode.Cc[syscall.VTIME] = 0
	err = setMode(fd, mode)
	if err != nil {
		return err
	}

	t.Puts(termEnterCA)
	t.Puts(termHideCursor)
	t.Puts(termEnterAcs)
	t.Puts(termClear)

	t.sigwinch = make(chan os.Signal, 1)
	signal.Notify(t.sigwinch, syscall.SIGWINCH)

	t.runes = make(chan rune, 500)
	t.events = make(chan *Event, 500)
	go t.readRunes()
	go t.readEvents()

	return nil
}

func (t *Terminal) Stop() {
	fd := t.out.Fd()
	setMode(fd, t.originalMode)

	t.Puts(termShowCursor)
	t.Puts(termClear)
	t.Puts(termExitAcs)
	t.Puts(termExitCA)
	// t.Puts(termClear)
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
		r, _, err := t.in.ReadRune()
		if err != nil {
			panic(err) // TODO don't panic
		}
		t.runes <- r

	}
}

func (t *Terminal) readEvents() {
	for {
		select {
		case <-t.sigwinch:
			t.events <- &Event{Type: EventResize}
		case r := <-t.runes:
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
	termClear      string = "\x1b[H\x1b[J"
	termSetColor          = "\x1b[%d;%dm"
	termSetFgBg           = "\x1b[3%d;4%dm"
	termAttrOff           = "\x1b[0;10m"
	termReset             = "\x1b[0m"
	termEnterAcs          = "\x1b[11m"
	termExitAcs           = "\x1b[10m"
	termEnterCA           = "\x1b[?1049h"
	termExitCA            = "\x1b[r\x1b[?1049l"
	termHideCursor        = "\x1b[?25l"
	termShowCursor        = "\x1b[?25h"
	termSetCursor         = "\x1b[%d;%dH"
)

func (t *Terminal) Clear() {
	fmt.Print(termClear)
}

func (t *Terminal) MoveCursor(x, y int) {
	t.Puts(termSetCursor, y+1, x+1)
}

func (t *Terminal) SetColor(c Color, bright bool) {
	var brightValue = 0
	if bright {
		brightValue = 1
	}
	t.Puts(termSetColor, brightValue, c)
}
func (t *Terminal) SetFg(c Color, bright bool) {
	t.SetColor(30+c, bright)
}

func (t *Terminal) SetBg(c Color, bright bool) {
	t.SetColor(40+c, bright)
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
	ColorWhite
	ColorDefault
)

type Key int

const (
	KeyNull  Key = 0
	KeyCtrlA     = 1
	KeyCtrlB     = 2
	KeyCtrlC     = 3
	KeyCtrlD     = 4
	KeyCtrlE     = 5
	KeyCtrlF     = 6
	KeyCtrlG     = 7
	KeyCtrlH     = 8
	KeyTab       = 9
	KeyLf        = 10
	KeyCtrlK     = 11
	KeyCtrlL     = 12
	KeyCr        = 13
	KeyCtrlN     = 14
	KeyCtrlO     = 15
	KeyCtrlP     = 16
	KeyCtrlQ     = 17
	KeyCtrlR     = 18
	KeyCtrlS     = 19
	KeyCtrlT     = 20
	KeyCtrlU     = 21
	KeyCtrlV     = 22
	KeyCtrlW     = 23
	KeyCtrlX     = 24
	KeyCtrlY     = 25
	KeyCtrlZ     = 26
	KeyEsc       = 27
	KeyBs        = 127
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
