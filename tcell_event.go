package main

import (
	"runtime"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	doubleClickDuration = 500 * time.Millisecond
)

type TcellEventHandler struct {
	prevMouseButton tcell.ButtonMask
	prevDownTime    time.Time
	clickY          []int
}

func (h *TcellEventHandler) Handle(screen tcell.Screen) Event {
	ev := screen.PollEvent()
	switch ev := ev.(type) {
	case *tcell.EventResize:
		return Event{Resize, 0, nil}

	// process mouse events:
	case *tcell.EventMouse:
		// mouse down events have zeroed buttons, so we can't use them
		// mouse up event consists of two events, 1. (main) event with modifier and other metadata, 2. event with zeroed buttons
		// so mouse click is three consecutive events, but the first and last are indistinguishable from movement events (with released buttons)
		// dragging has same structure, it only repeats the middle (main) event appropriately
		x, y := ev.Position()
		mod := ev.Modifiers() != 0

		// since we dont have mouse down events (unlike LightRenderer), we need to track state in prevButton
		prevButton, button := h.prevMouseButton, ev.Buttons()
		h.prevMouseButton = button
		drag := prevButton == button

		switch {
		case button&tcell.WheelDown != 0:
			return Event{Mouse, 0, &MouseEvent{y, x, -1, false, false, false, mod}}
		case button&tcell.WheelUp != 0:
			return Event{Mouse, 0, &MouseEvent{y, x, +1, false, false, false, mod}}
		case button&tcell.Button1 != 0 && !drag:
			// all potential double click events put their 'line' coordinate in the clickY array
			// double click event has two conditions, temporal and spatial, the first is checked here
			now := time.Now()
			if now.Sub(h.prevDownTime) < doubleClickDuration {
				h.clickY = append(h.clickY, y)
			} else {
				h.clickY = []int{y}
			}
			h.prevDownTime = now

			// detect double clicks (also check for spatial condition)
			n := len(h.clickY)
			double := n > 1 && h.clickY[n-2] == h.clickY[n-1]
			if double {
				// make sure two consecutive double clicks require four clicks
				h.clickY = []int{}
			}

			// fire single or double click event
			return Event{Mouse, 0, &MouseEvent{y, x, 0, true, !double, double, mod}}
		case button&tcell.Button2 != 0 && !drag:
			return Event{Mouse, 0, &MouseEvent{y, x, 0, false, true, false, mod}}
		case runtime.GOOS != "windows":

			// double and single taps on Windows don't quite work due to
			// the console acting on the events and not allowing us
			// to consume them.

			left := button&tcell.Button1 != 0
			down := left || button&tcell.Button3 != 0
			double := false
			if down {
				now := time.Now()
				if !left {
					h.clickY = []int{}
				} else if now.Sub(h.prevDownTime) < doubleClickDuration {
					h.clickY = append(h.clickY, x)
				} else {
					h.clickY = []int{x}
					h.prevDownTime = now
				}
			} else {
				if len(h.clickY) > 1 && h.clickY[0] == h.clickY[1] &&
					time.Now().Sub(h.prevDownTime) < doubleClickDuration {
					double = true
				}
			}

			return Event{Mouse, 0, &MouseEvent{y, x, 0, left, down, double, mod}}
		}

		// process keyboard:
	case *tcell.EventKey:
		mods := ev.Modifiers()
		none := mods == tcell.ModNone
		alt := (mods & tcell.ModAlt) > 0
		ctrl := (mods & tcell.ModCtrl) > 0
		shift := (mods & tcell.ModShift) > 0
		ctrlAlt := ctrl && alt
		altShift := alt && shift

		keyfn := func(r rune) Event {
			if alt {
				return CtrlAltKey(r)
			}
			return EventType(CtrlA.Int() - 'a' + int(r)).AsEvent()
		}
		switch ev.Key() {
		// section 1: Ctrl+(Alt)+[a-z]
		case tcell.KeyCtrlA:
			return keyfn('a')
		case tcell.KeyCtrlB:
			return keyfn('b')
		case tcell.KeyCtrlC:
			return keyfn('c')
		case tcell.KeyCtrlD:
			return keyfn('d')
		case tcell.KeyCtrlE:
			return keyfn('e')
		case tcell.KeyCtrlF:
			return keyfn('f')
		case tcell.KeyCtrlG:
			return keyfn('g')
		case tcell.KeyCtrlH:
			switch ev.Rune() {
			case 0:
				if ctrl {
					return Event{BSpace, 0, nil}
				}
			case rune(tcell.KeyCtrlH):
				switch {
				case ctrl:
					return keyfn('h')
				case alt:
					return Event{AltBS, 0, nil}
				case none, shift:
					return Event{BSpace, 0, nil}
				}
			}
		case tcell.KeyCtrlI:
			return keyfn('i')
		case tcell.KeyCtrlJ:
			return keyfn('j')
		case tcell.KeyCtrlK:
			return keyfn('k')
		case tcell.KeyCtrlL:
			return keyfn('l')
		case tcell.KeyCtrlM:
			return keyfn('m')
		case tcell.KeyCtrlN:
			return keyfn('n')
		case tcell.KeyCtrlO:
			return keyfn('o')
		case tcell.KeyCtrlP:
			return keyfn('p')
		case tcell.KeyCtrlQ:
			return keyfn('q')
		case tcell.KeyCtrlR:
			return keyfn('r')
		case tcell.KeyCtrlS:
			return keyfn('s')
		case tcell.KeyCtrlT:
			return keyfn('t')
		case tcell.KeyCtrlU:
			return keyfn('u')
		case tcell.KeyCtrlV:
			return keyfn('v')
		case tcell.KeyCtrlW:
			return keyfn('w')
		case tcell.KeyCtrlX:
			return keyfn('x')
		case tcell.KeyCtrlY:
			return keyfn('y')
		case tcell.KeyCtrlZ:
			return keyfn('z')
		// section 2: Ctrl+[ \]_]
		case tcell.KeyCtrlSpace:
			return Event{CtrlSpace, 0, nil}
		case tcell.KeyCtrlBackslash:
			return Event{CtrlBackSlash, 0, nil}
		case tcell.KeyCtrlRightSq:
			return Event{CtrlRightBracket, 0, nil}
		case tcell.KeyCtrlCarat:
			return Event{CtrlCaret, 0, nil}
		case tcell.KeyCtrlUnderscore:
			return Event{CtrlSlash, 0, nil}
		// section 3: (Alt)+Backspace2
		case tcell.KeyBackspace2:
			if alt {
				return Event{AltBS, 0, nil}
			}
			return Event{BSpace, 0, nil}

		// section 4: (Alt+Shift)+Key(Up|Down|Left|Right)
		case tcell.KeyUp:
			if altShift {
				return Event{AltSUp, 0, nil}
			}
			if shift {
				return Event{SUp, 0, nil}
			}
			if alt {
				return Event{AltUp, 0, nil}
			}
			return Event{Up, 0, nil}
		case tcell.KeyDown:
			if altShift {
				return Event{AltSDown, 0, nil}
			}
			if shift {
				return Event{SDown, 0, nil}
			}
			if alt {
				return Event{AltDown, 0, nil}
			}
			return Event{Down, 0, nil}
		case tcell.KeyLeft:
			if altShift {
				return Event{AltSLeft, 0, nil}
			}
			if shift {
				return Event{SLeft, 0, nil}
			}
			if alt {
				return Event{AltLeft, 0, nil}
			}
			return Event{Left, 0, nil}
		case tcell.KeyRight:
			if altShift {
				return Event{AltSRight, 0, nil}
			}
			if shift {
				return Event{SRight, 0, nil}
			}
			if alt {
				return Event{AltRight, 0, nil}
			}
			return Event{Right, 0, nil}

		// section 5: (Insert|Home|Delete|End|PgUp|PgDn|BackTab|F1-F12)
		case tcell.KeyInsert:
			return Event{Insert, 0, nil}
		case tcell.KeyHome:
			return Event{Home, 0, nil}
		case tcell.KeyDelete:
			return Event{Del, 0, nil}
		case tcell.KeyEnd:
			return Event{End, 0, nil}
		case tcell.KeyPgUp:
			return Event{PgUp, 0, nil}
		case tcell.KeyPgDn:
			return Event{PgDn, 0, nil}
		case tcell.KeyBacktab:
			return Event{BTab, 0, nil}
		case tcell.KeyF1:
			return Event{F1, 0, nil}
		case tcell.KeyF2:
			return Event{F2, 0, nil}
		case tcell.KeyF3:
			return Event{F3, 0, nil}
		case tcell.KeyF4:
			return Event{F4, 0, nil}
		case tcell.KeyF5:
			return Event{F5, 0, nil}
		case tcell.KeyF6:
			return Event{F6, 0, nil}
		case tcell.KeyF7:
			return Event{F7, 0, nil}
		case tcell.KeyF8:
			return Event{F8, 0, nil}
		case tcell.KeyF9:
			return Event{F9, 0, nil}
		case tcell.KeyF10:
			return Event{F10, 0, nil}
		case tcell.KeyF11:
			return Event{F11, 0, nil}
		case tcell.KeyF12:
			return Event{F12, 0, nil}

		// section 6: (Ctrl+Alt)+'rune'
		case tcell.KeyRune:
			r := ev.Rune()

			switch {
			// translate native key events to ascii control characters
			case r == ' ' && ctrl:
				return Event{CtrlSpace, 0, nil}
			// handle AltGr characters
			case ctrlAlt:
				return Event{Rune, r, nil} // dropping modifiers
			// simple characters (possibly with modifier)
			case alt:
				return AltKey(r)
			default:
				return Event{Rune, r, nil}
			}

		// section 7: Esc
		case tcell.KeyEsc:
			return Event{ESC, 0, nil}
		}
	}

	// section 8: Invalid
	return Event{Invalid, 0, nil}
}
