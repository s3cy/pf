package main

import (
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type PathView struct {
	Win      *Win
	User     string
	Host     string
	Home     string
	Path     string
	Style    tcell.Style
	StyleMap map[string]tcell.Style
}

func (v *PathView) Draw() {
	v.Win.Reset(v.Style)

	if v.Path == "" {
		return
	}

	name := strings.Join([]string{v.User, v.Host}, "@")
	nameStyle := v.StyleMap["ex"]
	dir, file := filepath.Split(v.Path)
	if strings.HasPrefix(dir, v.Home) {
		dir = "~" + strings.TrimPrefix(dir, v.Home)
	}
	dirStyle := v.StyleMap["di"]

	contents := make([]Content, 0, len(name)+1+len(dir)+1+len(file))
	add := func(r rune, style *tcell.Style) {
		contents = append(contents, Content{
			R:     r,
			Style: style,
		})
	}

	for _, r := range name {
		add(r, &nameStyle)
	}
	add(':', nil)
	for _, r := range dir {
		add(r, &dirStyle)
	}
	for _, r := range file {
		add(r, nil)
	}

	v.Win.Render(0, 0, contents, v.Style, true)
}

