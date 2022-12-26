package main

import (
	"github.com/gdamore/tcell/v2"
)

type Content struct {
	R     rune
	Style *tcell.Style
}

type ListItem []Content

func (i *ListItem) WriteString(s string, style *tcell.Style) {
	for _, r := range s {
		i.WriteContent(r, style)
	}
}

func (i *ListItem) WriteContent(r rune, style *tcell.Style) {
	*i = append(*i, Content{R: r, Style: style})
}

type ListRow struct {
	FileInfo *FileInfo
	Left     func(bool) ListItem
	Right    func(bool) ListItem
	Style    *tcell.Style
}

type List struct {
	rows  []ListRow
	Style tcell.Style
	Marks map[string]struct{}
}

func (d *List) GetFileInfo(idx int) *FileInfo {
	if idx < 0 || idx >= len(d.rows) {
		return nil
	}
	row := d.rows[idx]
	return row.FileInfo
}

func (d *List) Get(idx int, width int, selected bool) ListItem {
	if idx < 0 || idx >= len(d.rows) || width <= 2 {
		return nil
	}

	row := d.rows[idx]
	style := d.Style
	if row.Style != nil {
		style = *row.Style
	}
	if selected {
		style = style.Reverse(true)
	}

	contents := make([]Content, width)
	if _, ok := d.Marks[row.FileInfo.Path]; ok {
		contents[0].R = '>'
		style := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorRed)
		contents[0].Style = &style
	}
	contents[1].R = ' '
	if selected {
		contents[1].Style = &style
	}
	i := 2
	putList := func(x ListItem) {
		for _, c := range x {
			if i >= width {
				break
			}
			contents[i].R = c.R
			contents[i].Style = c.Style
			if contents[i].Style == nil {
				contents[i].Style = &style
			}
			i++
		}
	}
	putList(row.Left(selected))

	right := row.Right(selected)
	avail := width - i - 1
	if avail <= 0 {
		right = nil
	} else if len(right) > avail {
		right = right[:avail]
	}
	rBegin := width - len(right)

	if selected {
		for ; i < rBegin; i++ {
			contents[i].R = ' '
			contents[i].Style = &style
		}
	} else {
		i = rBegin
	}

	putList(right)

	return contents
}

func (d *List) Size() int {
	return len(d.rows)
}

func (d *List) UpdateRows(rows []ListRow) {
	d.rows = rows
}
