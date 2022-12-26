package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

func sortBySize(files []*FileInfo) {
	sort.Slice(files, func(i, j int) bool { return files[i].Size() < files[j].Size() })
}

func sortByName(files []*FileInfo) {
	lessLower := func(sa, sb string) bool {
		for {
			rb, nb := utf8.DecodeRuneInString(sb)
			if nb == 0 {
				// The number of runes in sa is greater than or
				// equal to the number of runes in sb. It follows
				// that sa is not less than sb.
				return false
			}

			ra, na := utf8.DecodeRuneInString(sa)
			if na == 0 {
				// The number of runes in sa is less than the
				// number of runes in sb. It follows that sa
				// is less than sb.
				return true
			}

			rb = unicode.ToLower(rb)
			ra = unicode.ToLower(ra)

			if ra != rb {
				return ra < rb
			}

			// Trim rune from the beginning of each string.
			sa = sa[na:]
			sb = sb[nb:]
		}
	}
	sort.Slice(files, func(i, j int) bool { return lessLower(files[i].Name(), files[j].Name()) })
}

type DirEvent struct {
	Path string
	Rows []ListRow
	Err  error
}

type DirSet struct {
	dirs    []*Dir
	eventCh chan DirEvent
	styles  StyleMap
}

func NewDirSet(styles StyleMap) *DirSet {
	return &DirSet{
		eventCh: make(chan DirEvent, 1),
		styles:  styles,
	}
}

func (c *DirSet) Add(path string, cmds []string) {
	_, ok := c.find(path)
	if !ok {
		dir := NewDir(path, c.styles, c.eventCh)
		go dir.Run(cmds)
		c.dirs = append(c.dirs, dir)
	}
}

func (c *DirSet) Remove(path string) {
	i, ok := c.find(path)
	if ok {
		c.dirs[i].Fini()
		c.dirs = append(c.dirs[:i], c.dirs[i+1:]...)
	}
}

func (c *DirSet) find(path string) (int, bool) {
	for i, dir := range c.dirs {
		if dir.path == path {
			return i, true
		}
	}
	return 0, false
}

func (c *DirSet) Get(path string) *Dir {
	i, ok := c.find(path)
	if ok {
		return c.dirs[i]
	}
	return nil
}

func (c *DirSet) Event() <-chan DirEvent {
	return c.eventCh
}

type permColumnFormat int

const (
	permColumnIgnore permColumnFormat = iota
	permColumnPerm
)

type userColumnFormat int

const (
	userColumnIgnore userColumnFormat = iota
	userColumnUserName
	userColumnGroupName
	userColumnName // both names
)

type linkTargetColumnFormat int

const (
	linkTargetColumnIgnore linkTargetColumnFormat = iota
	linkTargetColumnLink
)

type linkCountColumnFormat int

const (
	linkCountColumnIgnore linkCountColumnFormat = iota
	linkCountColumnLink
)

type sizeColumnFormat int

const (
	sizeColumnIgnore sizeColumnFormat = iota
	sizeColumnHumanizeSize
	sizeColumnSize
)

type timeColumnFormat int

const (
	timeColumnIgnore timeColumnFormat = iota
	timeColumnAccessTime
	timeColumnChangeTime
	timeColumnModTime
)

type Dir struct {
	eventCh       chan<- DirEvent
	path          string
	sort          func([]*FileInfo)
	filteredFiles []*FileInfo
	files         []*FileInfo
	styles        StyleMap
	cmdCh         chan []string
	fini          int32 // atomic bool

	permColumn       permColumnFormat
	userColumn       userColumnFormat
	linkTargetColumn linkTargetColumnFormat
	linkCountColumn  linkCountColumnFormat
	sizeColumn       sizeColumnFormat
	timeColumn       timeColumnFormat
}

func NewDir(path string, styles StyleMap, eventCh chan<- DirEvent) *Dir {
	return &Dir{
		eventCh: eventCh,
		path:    path,
		styles:  styles,
		cmdCh:   make(chan []string, 1),
	}
}

func (d *Dir) Run(cmds []string) {
	if err := d.init(); err != nil {
		d.eventCh <- DirEvent{
			Path: d.path,
			Err:  err,
		}
		return
	}

	d.do(cmds)
	for cmds := range d.cmdCh {
		d.do(cmds)
	}
}

func (d *Dir) Fini() {
	if atomic.SwapInt32(&d.fini, 1) == 0 {
		close(d.cmdCh)
	}
}

func (d *Dir) Do(cmds []string) {
	d.cmdCh <- cmds
}

func (d *Dir) do(cmds []string) {
	var shouldSort bool
	for _, cmd := range cmds {
		pair := strings.SplitN(cmd, " ", 2)
		switch pair[0] {
		case "filter":
			if len(pair) < 2 {
				d.filter("")
			} else {
				d.filter(pair[1])
			}
		case "sort_by_size":
			d.sort = sortBySize
			shouldSort = true
		case "perm":
			d.permColumn = permColumnPerm
		case "user_name":
			d.userColumn = userColumnUserName
		case "group_name":
			d.userColumn = userColumnGroupName
		case "user_group_name":
			d.userColumn = userColumnName
		case "link_target":
			d.linkTargetColumn = linkTargetColumnLink
		case "link_count":
			d.linkCountColumn = linkCountColumnLink
		case "hsize":
			d.sizeColumn = sizeColumnHumanizeSize
		case "size":
			d.sizeColumn = sizeColumnSize
		case "atime":
			d.timeColumn = timeColumnAccessTime
		case "ctime":
			d.timeColumn = timeColumnChangeTime
		case "mtime":
			d.timeColumn = timeColumnModTime
		case "no_perm":
			d.permColumn = permColumnIgnore
		case "no_link_target":
			d.linkTargetColumn = linkTargetColumnIgnore
		case "no_link_count":
			d.linkCountColumn = linkCountColumnIgnore
		case "no_size":
			d.sizeColumn = sizeColumnIgnore
		case "no_user":
			d.userColumn = userColumnIgnore
		case "no_time":
			d.timeColumn = timeColumnIgnore
		case "reset_info":
			d.filteredFiles = nil
			d.permColumn = permColumnIgnore
			d.linkTargetColumn = linkTargetColumnIgnore
			d.linkCountColumn = linkCountColumnIgnore
			d.sizeColumn = sizeColumnIgnore
			d.userColumn = userColumnIgnore
			d.timeColumn = timeColumnIgnore
		default:
			log.Fatalf("unknown cmd: %s", cmd)
		}
	}

	if shouldSort {
		d.sort(d.files)
	}
	d.sendToC()
}

func (d *Dir) filter(s string) {
	if s == "" {
		d.filteredFiles = nil
		return
	}

	cmd := exec.Command("fzf", "-f", s)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("%+v", err)
	}

	go func() {
		buf := bufio.NewWriter(stdin)
		for _, file := range d.files {
			buf.WriteString(file.Name())
			buf.WriteByte('\n')
		}
		buf.Flush()
		stdin.Close()
	}()

	m := make(map[string]struct{})
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("%+v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("%+v", err)
	}

	filteredFiles := make([]*FileInfo, 0, len(m))
	for _, file := range d.files {
		if _, ok := m[file.Name()]; ok {
			filteredFiles = append(filteredFiles, file)
		}
	}
	d.filteredFiles = filteredFiles
}

func (d *Dir) init() error {
	if d.sort == nil {
		d.sort = sortByName
	}

	f, err := os.Open(d.path)
	if err != nil {
		return err
	}
	defer f.Close()

	fiList, err := f.Readdir(-1)
	if err != nil {
		return err
	}

	files := make([]*FileInfo, 0, len(fiList))
	for _, fi := range fiList {
		files = append(files, NewFileInfo(fi, d.path))
	}
	d.sort(files)

	d.files = files

	return nil
}

func (d *Dir) sendToC() {
	permColumn := d.permColumn
	userColumn := d.userColumn
	linkTargetColumn := d.linkTargetColumn
	linkCountColumn := d.linkCountColumn
	sizeColumn := d.sizeColumn
	timeColumn := d.timeColumn

	left := func(info *FileInfo) ListItem {
		item := ListItem{}
		item.WriteString(info.Name(), nil)
		if linkTargetColumn == linkTargetColumnLink {
			var linkTarget string
			if info.LinkTarget != "" {
				linkTarget = " -> " + info.LinkTarget
			}

			item.WriteString(linkTarget, nil)
		}
		return item
	}

	right := func(info *FileInfo, selected bool) ListItem {
		item := ListItem{}
		if permColumn == permColumnPerm {
			permSt := tcell.StyleDefault
			item = append(item, info.ANSIMode(permSt)...)
		}

		if linkCountColumn == linkCountColumnLink {
			linkSt := tcell.StyleDefault
			item.WriteContent(' ', &linkSt)
			item.WriteString(info.LinkCount(), &linkSt)
		}

		sizeSt := tcell.StyleDefault.Foreground(tcell.ColorGreen).Reverse(selected)
		if info.IsDir() {
			if sizeColumn != sizeColumnIgnore {
				sizeSt = sizeSt.Foreground(tcell.ColorGray)
				item.WriteString(fmt.Sprintf(" %4s", info.HumanizeSize()), &sizeSt)
			}
		} else {
			if sizeColumn == sizeColumnHumanizeSize {
				item.WriteString(fmt.Sprintf(" %4s", info.HumanizeSize()), &sizeSt)
			} else if sizeColumn == sizeColumnSize {
				item.WriteString(fmt.Sprintf(" %v", info.Size()), &sizeSt)
			}
		}

		userSt := tcell.StyleDefault.Foreground(tcell.ColorYellow).Reverse(selected)
		if userColumn == userColumnUserName {
			item.WriteString(" "+info.UserName(), &userSt)
		} else if userColumn == userColumnGroupName {
			item.WriteString(" "+info.GroupName(), &userSt)
		} else if userColumn == userColumnName {
			item.WriteString(" "+info.UserName(), &userSt)
			item.WriteString(" "+info.GroupName(), &userSt)
		}

		timeSt := tcell.StyleDefault.Foreground(tcell.ColorBlue).Reverse(selected)
		if timeColumn == timeColumnAccessTime {
			item.WriteString(" "+info.AccessTime().Format(time.ANSIC), &timeSt)
		} else if timeColumn == timeColumnChangeTime {
			item.WriteString(" "+info.ChangeTime().Format(time.ANSIC), &timeSt)
		} else if timeColumn == timeColumnModTime {
			item.WriteString(" "+info.ModTime().Format(time.ANSIC), &timeSt)
		}

		return item
	}

	files := d.filteredFiles
	if files == nil {
		files = d.files
	}
	rows := make([]ListRow, 0, len(files))
	for i := 0; i < len(files); i++ {
		file := files[i]
		style := d.styles.Get(file)

		row := ListRow{
			FileInfo: file,
			Left: func(selected bool) ListItem {
				return left(file)
			},
			Right: func(selected bool) ListItem {
				return right(file, selected)
			},
			Style: &style,
		}
		rows = append(rows, row)
	}

	d.eventCh <- DirEvent{
		Path: d.path,
		Rows: rows,
	}
}
