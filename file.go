package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"time"

	"github.com/djherbis/times"
	"github.com/gdamore/tcell/v2"
)

type LinkState int

const (
	LinkStateNone LinkState = iota
	LinkStateWorking
	LinkStateBroken
)

type FileInfo struct {
	fs.FileInfo
	Path       string
	LinkState  LinkState
	LinkTarget string
	Ext        string

	accessTime time.Time
	changeTime time.Time
}

func NewFileInfo(info fs.FileInfo, dir string) *FileInfo {
	ts := times.Get(info)
	at := ts.AccessTime()
	var ct time.Time
	// from times docs: ChangeTime() panics unless HasChangeTime() is true
	if ts.HasChangeTime() {
		ct = ts.ChangeTime()
	} else {
		// fall back to ModTime if ChangeTime cannot be determined
		ct = info.ModTime()
	}

	fi := &FileInfo{
		FileInfo:   info,
		Path:       filepath.Join(dir, info.Name()),
		Ext:        filepath.Ext(info.Name()),
		accessTime: at,
		changeTime: ct,
	}

	if info.Mode()&fs.ModeSymlink != 0 {
		link, err := os.Readlink(fi.Path)
		if err != nil {
			fi.LinkState = LinkStateBroken
		} else {
			fi.LinkState = LinkStateWorking
			fi.LinkTarget = link
		}
	}
	return fi
}

func (i *FileInfo) Info() ListItem {
	var linkTarget string
	if i.LinkTarget != "" {
		linkTarget = " -> " + i.LinkTarget
	}

	s := fmt.Sprintf("%v %v%v%v%4s %v%s",
		i.Mode(),
		i.LinkCount(),
		i.UserName(),
		i.GroupName(),
		i.HumanizeSize(),
		i.ModTime().Format(time.ANSIC),
		linkTarget)

	contents := make(ListItem, 0, len(s))
	contents.WriteString(s, nil)
	return contents
}

func (i *FileInfo) ANSIMode(style tcell.Style) ListItem {
	mode := fmt.Sprintf("%v", i.Mode())
	contents := make(ListItem, 0, len(mode))
	for _, r := range mode {
		st := style
		switch r {
		case 'd':
			st = st.Foreground(tcell.ColorBlue)
		case 'r':
			st = st.Foreground(tcell.ColorYellow)
		case 'w':
			st = st.Foreground(tcell.ColorRed)
		case 'x':
			st = st.Foreground(tcell.ColorGreen)
		default:
			st = st.Foreground(tcell.ColorGray)
		}
		contents.WriteContent(r, &st)
	}
	return contents
}

func (i *FileInfo) UserName() string {
	if stat, ok := i.Sys().(*syscall.Stat_t); ok {
		if u, err := user.LookupId(fmt.Sprint(stat.Uid)); err == nil {
			return fmt.Sprintf("%v ", u.Username)
		}
	}
	return ""
}

func (i *FileInfo) GroupName() string {
	if stat, ok := i.Sys().(*syscall.Stat_t); ok {
		if g, err := user.LookupGroupId(fmt.Sprint(stat.Gid)); err == nil {
			return fmt.Sprintf("%v ", g.Name)
		}
	}
	return ""
}

func (i *FileInfo) LinkCount() string {
	if stat, ok := i.Sys().(*syscall.Stat_t); ok {
		return fmt.Sprintf("%v ", stat.Nlink)
	}
	return ""
}

func (i *FileInfo) HumanizeSize() string {
	if i.IsDir() {
		return "-"
	}
	return humanize(i.Size())
}

// This function converts a size in bytes to a human readable form using metric
// suffixes (e.g. 1K = 1000). For values less than 10 the first significant
// digit is shown, otherwise it is hidden. Numbers are always rounded down.
// This should be fine for most human beings.
func humanize(size int64) string {
	if size < 1000 {
		return fmt.Sprintf("%dB", size)
	}

	suffix := []string{
		"K", // kilo
		"M", // mega
		"G", // giga
		"T", // tera
		"P", // peta
		"E", // exa
		"Z", // zeta
		"Y", // yotta
	}

	curr := float64(size) / 1000
	for _, s := range suffix {
		if curr < 10 {
			return fmt.Sprintf("%.1f%s", curr-0.0499, s)
		} else if curr < 1000 {
			return fmt.Sprintf("%d%s", int(curr), s)
		}
		curr /= 1000
	}

	return ""
}

func (i *FileInfo) AccessTime() time.Time {
	return i.accessTime
}

func (i *FileInfo) ChangeTime() time.Time {
	return i.changeTime
}
