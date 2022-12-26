package main

import (
	"log"
	"strings"
)

const (
	escapedColon = 0
	escapedComma = 1
	escapedPlus  = 2
)

func parseKeyChord(key string) Event {
	if len(key) == 0 {
		log.Fatal("empty key")
	}

	lkey := strings.ToLower(key)
	switch lkey {
	case "up":
		return Up.AsEvent()
	case "down":
		return Down.AsEvent()
	case "left":
		return Left.AsEvent()
	case "right":
		return Right.AsEvent()
	case "enter", "return":
		return CtrlM.AsEvent()
	case "space":
		return Key(' ')
	case "bspace", "bs":
		return BSpace.AsEvent()
	case "ctrl-space":
		return CtrlSpace.AsEvent()
	case "ctrl-^", "ctrl-6":
		return CtrlCaret.AsEvent()
	case "ctrl-/", "ctrl-_":
		return CtrlSlash.AsEvent()
	case "ctrl-\\":
		return CtrlBackSlash.AsEvent()
	case "ctrl-]":
		return CtrlRightBracket.AsEvent()
	case "change":
		return Change.AsEvent()
	case "backward-eof":
		return BackwardEOF.AsEvent()
	case "start":
		return Start.AsEvent()
	case "alt-enter", "alt-return":
		return CtrlAltKey('m')
	case "alt-space":
		return AltKey(' ')
	case "alt-bs", "alt-bspace":
		return AltBS.AsEvent()
	case "alt-up":
		return AltUp.AsEvent()
	case "alt-down":
		return AltDown.AsEvent()
	case "alt-left":
		return AltLeft.AsEvent()
	case "alt-right":
		return AltRight.AsEvent()
	case "tab":
		return Tab.AsEvent()
	case "btab", "shift-tab":
		return BTab.AsEvent()
	case "esc":
		return ESC.AsEvent()
	case "del":
		return Del.AsEvent()
	case "home":
		return Home.AsEvent()
	case "end":
		return End.AsEvent()
	case "insert":
		return Insert.AsEvent()
	case "pgup", "page-up":
		return PgUp.AsEvent()
	case "pgdn", "page-down":
		return PgDn.AsEvent()
	case "alt-shift-up", "shift-alt-up":
		return AltSUp.AsEvent()
	case "alt-shift-down", "shift-alt-down":
		return AltSDown.AsEvent()
	case "alt-shift-left", "shift-alt-left":
		return AltSLeft.AsEvent()
	case "alt-shift-right", "shift-alt-right":
		return AltSRight.AsEvent()
	case "shift-up":
		return SUp.AsEvent()
	case "shift-down":
		return SDown.AsEvent()
	case "shift-left":
		return SLeft.AsEvent()
	case "shift-right":
		return SRight.AsEvent()
	case "left-click":
		return LeftClick.AsEvent()
	case "right-click":
		return RightClick.AsEvent()
	case "double-click":
		return DoubleClick.AsEvent()
	case "f10":
		return F10.AsEvent()
	case "f11":
		return F11.AsEvent()
	case "f12":
		return F12.AsEvent()
	default:
		runes := []rune(key)
		if len(key) == 10 && strings.HasPrefix(lkey, "ctrl-alt-") && isAlphabet(lkey[9]) {
			return CtrlAltKey(rune(key[9]))
		}
		if len(key) == 6 && strings.HasPrefix(lkey, "ctrl-") && isAlphabet(lkey[5]) {
			return EventType(CtrlA.Int() + int(lkey[5]) - 'a').AsEvent()
		}
		if len(runes) == 5 && strings.HasPrefix(lkey, "alt-") {
			r := runes[4]
			switch r {
			case escapedColon:
				r = ':'
			case escapedComma:
				r = ','
			case escapedPlus:
				r = '+'
			}
			return AltKey(r)
		}
		if len(key) == 2 && strings.HasPrefix(lkey, "f") && key[1] >= '1' && key[1] <= '9' {
			return EventType(F1.Int() + int(key[1]) - '1').AsEvent()
		}
		if len(runes) == 1 {
			return Key(runes[0])
		}
	}
	log.Fatalf("unsupported key: " + key)

	// Unreachable
	return Invalid.AsEvent()
}

func ParseKeymap(keymap map[Event][]Action, cmds map[string]CMD, str string) {
	masked := strings.Replace(str, "::", string([]rune{escapedColon, ':'}), -1)
	masked = strings.Replace(masked, ",:", string([]rune{escapedComma, ':'}), -1)
	masked = strings.Replace(masked, "+:", string([]rune{escapedPlus, ':'}), -1)

	for _, pairStr := range strings.Split(masked, ",") {
		pair := strings.SplitN(pairStr, ":", 2)

		var key Event
		if len(pair[0]) == 1 && pair[0][0] == escapedColon {
			key = Key(':')
		} else if len(pair[0]) == 1 && pair[0][0] == escapedComma {
			key = Key(',')
		} else if len(pair[0]) == 1 && pair[0][0] == escapedPlus {
			key = Key('+')
		} else {
			key = parseKeyChord(pair[0])
		}

		specs := strings.Split(pair[1], "+")
		actions := make([]Action, 0, len(specs))
		for _, spec := range specs {
			tokens := strings.Split(spec, " ")
			cmd, ok := cmds[tokens[0]]
			if !ok {
				log.Fatalf("unknown cmd %s", tokens[0])
			}

			action := ToAction(cmd, tokens[1:])
			actions = append(actions, action)
		}

		keymap[key] = actions
	}
}

func isAlphabet(char uint8) bool {
	return char >= 'a' && char <= 'z'
}
