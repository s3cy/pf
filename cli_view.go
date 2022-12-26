package main

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

type CLIView struct {
	Win      *Win
	Info     *FileInfo
	Style    tcell.Style
	hideInfo bool
}

func (v *CLIView) ShowInfo() {
	if v.hideInfo {
		return
	}

	v.Win.Reset(v.Style)

	if v.Info == nil {
		return
	}

	v.Win.Render(0, 0, v.Info.Info(), v.Style, false)
}

func (v *CLIView) Warn(format string, a ...any) {
	v.hideInfo = true
	v.Win.Reset(v.Style)
	v.Win.RenderANSI(0, 0, fmt.Sprintf(format, a...), v.Style.Background(tcell.ColorRed))
	time.AfterFunc(3 * time.Second, func() { v.hideInfo = false })
}

func (v *CLIView) CMD(s string, cursor int) {
	v.hideInfo = s != ""
	v.Win.Reset(v.Style)
	if s == "" {
		return
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(s) {
		cursor = len(s)
	}
	contents := make(ListItem, 0, len(s)+1)
	st := v.Style.Reverse(true)
	for i, r := range s {
		if i == cursor {
			contents.WriteContent(r, &st)
		} else {
			contents.WriteContent(r, nil)
		}
	}
	if cursor == len(s) {
		contents.WriteContent(' ', &st)
	}

	v.Win.Render(0, 0, contents, v.Style, true)
}
