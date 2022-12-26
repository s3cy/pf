package main

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Win struct {
	X1, Y1, X2, Y2 int
	Screen         tcell.Screen
}

func (win *Win) Reset(st tcell.Style) {
	for row := win.Y1; row <= win.Y2; row++ {
		for col := win.X1; col <= win.X2; col++ {
			win.Screen.SetContent(col, row, ' ', nil, st)
		}
	}
}

func (win *Win) RenderANSI(x, y int, s string, st tcell.Style) {
	off := x
	var comb []rune
	for i := 0; i < len(s); i++ {
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == 27 && i+1 < len(s) && s[i+1] == '[' {
			j := strings.IndexAny(s[i:min(len(s), i+64)], "mK")
			if j == -1 {
				continue
			}
			if s[i+j] == 'm' {
				st = ApplyAnsiCodes(s[i+2:i+j], st)
			}

			i += j
			continue
		}

		for {
			rc, wc := utf8.DecodeRuneInString(s[i+w:])
			if !unicode.Is(unicode.Mn, rc) {
				break
			}
			comb = append(comb, rc)
			i += wc
		}

		if x < win.W() {
			win.Screen.SetContent(win.X1+x, win.Y1+y, r, comb, st)
			comb = nil
		}

		i += w - 1

		if r == '\t' {
			s := 4 - (x-off)%4
			for i := 0; i < s && x+i < win.W(); i++ {
				win.Screen.SetContent(win.X1+x+i, win.Y1+y, ' ', nil, st)
			}
			x += s
		} else {
			x += runewidth.RuneWidth(r)
		}
	}
}

func (win *Win) Render(x, y int, contents ListItem, st tcell.Style, wrap bool) {
	col, row := win.X1+x, win.Y1+y
	for _, c := range contents {
		if row > win.Y2 {
			break
		}
		if col > win.X2 {
			if !wrap {
				break
			}
			row++
			col = win.X1
		}

		style := st
		if c.Style != nil {
			style = *c.Style
		}
		win.Screen.SetContent(col, row, c.R, nil, style)

		col++
	}
}

func (win *Win) In(x, y int) bool {
	return win.X1 <= x && x <= win.X2 && win.Y1 <= y && y <= win.Y2
}

func (win *Win) H() int {
	return win.Y2 - win.Y1
}

func (win *Win) W() int {
	return win.X2 - win.X1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
