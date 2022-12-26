package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type listViewState struct {
	selected string
	offset   int // distance to top
}

type Controller struct {
	dirs            *DirSet
	path            *PathView
	cli             *CLIView
	left            *ListView
	main            *ListView
	screen          tcell.Screen
	cwd             string
	cwdInited       bool
	parentCwd       string
	parentCwdInited bool
	marks           map[string]struct{}
	listViewStates  map[string]listViewState
	dirInfoCMD      []string
}

func NewController(dirs *DirSet, marks map[string]struct{}, screen tcell.Screen) *Controller {
	defStyle := tcell.StyleDefault

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	dirs.Add(cwd, nil)

	parentCwd := ""
	if cwd != "/" {
		parentCwd = filepath.Dir(cwd)
		dirs.Add(parentCwd, nil)
	}

	path := &PathView{
		User:     UserName,
		Host:     HostName,
		Home:     UserHomeDir,
		Style:    defStyle,
		StyleMap: StyleM,
	}

	left := &ListView{
		List: List{
			Style: defStyle,
			Marks: marks,
		},
	}

	main := &ListView{
		List: List{
			Style: defStyle,
			Marks: marks,
		},
	}

	c := &Controller{
		dirs:           dirs,
		path:           path,
		cli:            &CLIView{},
		left:           left,
		main:           main,
		screen:         screen,
		cwd:            cwd,
		parentCwd:      parentCwd,
		marks:          marks,
		listViewStates: initListViewStates(cwd),
	}
	c.resize()

	return c
}

func initListViewStates(cwd string) map[string]listViewState {
	if cwd == "" || cwd == "/" || !strings.HasPrefix(cwd, "/") {
		return make(map[string]listViewState)
	}
	if strings.HasSuffix(cwd, "/") {
		cwd = cwd[:len(cwd)-1]
	}

	tokens := strings.Split(cwd, string(filepath.Separator))
	cache := make(map[string]listViewState, len(tokens)-1)
	path := "/"
	for _, token := range tokens[1:] {
		cache[path] = listViewState{
			selected: token,
		}
		path = filepath.Join(path, token)
	}
	return cache
}

func (c *Controller) Show() {
	if info := c.main.List.GetFileInfo(c.main.SelectAt); info != nil {
		if info.Path != c.path.Path {
			c.path.Path = info.Path
			c.path.Draw()
		}

		c.cli.Info = info
		c.cli.ShowInfo()
	}

	c.screen.Show()
}

func (c *Controller) resize() {
	width, height := c.screen.Size()

	// Top
	c.path.Win = &Win{
		X1:     0,
		X2:     width,
		Y1:     0,
		Y2:     0,
		Screen: c.screen,
	}
	c.path.Draw()

	// Bottom
	c.cli.Win = &Win{
		X1:     0,
		X2:     width,
		Y1:     height - 1,
		Y2:     height - 1,
		Screen: c.screen,
	}
	c.cli.ShowInfo()

	// Left 1/3
	c.left.Win = &Win{
		X1:     0,
		X2:     width / 3,
		Y1:     1,
		Y2:     height - 2,
		Screen: c.screen,
	}
	c.left.Draw()

	// Right 2/3
	c.main.Win = &Win{
		X1:     width/3 + 1,
		X2:     width,
		Y1:     1,
		Y2:     height - 2,
		Screen: c.screen,
	}
	c.main.Draw()
}

func (c *Controller) Resize() {
	c.resize()
	c.screen.Sync()
}

func (c *Controller) Next() {
	c.main.SelectAt++
	c.main.Draw()
}

func (c *Controller) Prev() {
	c.main.SelectAt--
	c.main.Draw()
}

func (c *Controller) Top() {
	c.main.SelectAt = 0
	c.main.Draw()
}

func (c *Controller) Bottom() {
	c.main.SelectAt = c.main.List.Size() - 1
	c.main.Draw()
}

func (c *Controller) ScrollDown(n int, left bool) {
	if left {
		selectChanged := c.left.ScrollDown(n)
		if selectChanged {
			c.Out()
			c.In()
			return
		}
		c.left.Draw()
		return
	}
	c.main.ScrollDown(n)
	c.main.Draw()
}

func (c *Controller) LeftScrollDown(n int) {
}

func (c *Controller) pageDown(factor float64) {
	off := int(float64(c.main.Win.H()) * factor)
	c.main.SelectAt += off
	c.main.Draw()
}

func (c *Controller) HalfPageDown() {
	c.pageDown(0.5)
}

func (c *Controller) HalfPageUp() {
	c.pageDown(-0.5)
}

func (c *Controller) PageDown() {
	c.pageDown(1)
}

func (c *Controller) PageUp() {
	c.pageDown(-1)
}

func (c *Controller) ClearMarks() {
	for mark := range c.marks {
		delete(c.marks, mark)
	}
}

func (c *Controller) Mark() {
	c.main.Mark(false, false)
	c.main.Draw()
}

func (c *Controller) Unmark() {
	c.main.Mark(true, false)
	c.main.Draw()
}

func (c *Controller) ToggleMark() {
	c.main.Mark(false, true)
	c.main.Draw()
}

func (c *Controller) inLeft(x, y int) bool {
	return c.left.Win.In(x, y)
}

func (c *Controller) inMain(x, y int) bool {
	return c.main.Win.In(x, y)
}

func (c *Controller) Select(x, y int) {
	if c.inLeft(x, y) {
		c.left.Select(y)
		c.Out()
		c.In()
		return
	}
	if c.inMain(x, y) {
		c.main.Select(y)
		c.main.Draw()
		return
	}
}

func (c *Controller) Goto(x, y int) {
	if c.inLeft(x, y) {
		c.left.Select(y)
		c.Out()
		return
	}
	if c.inMain(x, y) {
		c.main.Select(y)
		c.In()
	}
}

func (c *Controller) saveListViewState() {
	if info := c.main.List.GetFileInfo(c.main.SelectAt); info != nil {
		c.listViewStates[c.cwd] = listViewState{
			selected: info.Name(),
			offset:   c.main.SelectAt - c.main.ViewBeginAt,
		}
	}
}

func (c *Controller) In() {
	info := c.main.List.GetFileInfo(c.main.SelectAt)
	if info == nil || !(info.IsDir() || info.LinkState == LinkStateWorking) {
		return
	}

	c.saveListViewState()

	newCwd := info.Path
	if err := os.Chdir(newCwd); err != nil {
		c.cli.Warn("%+v", err)
		return
	}
	c.dirs.Add(newCwd, c.dirInfoCMD)
	c.dirs.Remove(c.parentCwd)
	c.parentCwd = c.cwd
	c.cwd = newCwd
	c.parentCwdInited = c.cwdInited
	c.cwdInited = false
	c.left.List = c.main.List
	c.left.ViewBeginAt = c.main.ViewBeginAt
	c.left.SelectAt = c.main.SelectAt
	c.left.Draw()

	// Clear info in parent dir
	if c.dirInfoCMD != nil {
		if dir := c.dirs.Get(c.parentCwd); dir != nil {
			dir.Do([]string{"reset_info"})
		}
	}
}

func (c *Controller) Out() {
	if c.cwd == "/" {
		return
	}

	c.saveListViewState()

	if err := os.Chdir(c.parentCwd); err != nil {
		c.cli.Warn("%+v", err)
		return
	}

	c.dirs.Remove(c.cwd)
	c.cwd = c.parentCwd
	c.cwdInited = c.parentCwdInited
	c.parentCwdInited = false
	c.main.List = c.left.List
	c.main.ViewBeginAt = c.left.ViewBeginAt
	c.main.SelectAt = c.left.SelectAt
	c.main.Draw()

	if c.parentCwd == "/" {
		c.parentCwd = ""
		c.left.List.UpdateRows(nil)
		c.left.Draw()
	} else {
		newParentCwd := filepath.Dir(c.parentCwd)
		c.dirs.Add(newParentCwd, nil)
		c.parentCwd = newParentCwd
	}

	// Show info in current dir
	if c.dirInfoCMD != nil {
		if dir := c.dirs.Get(c.cwd); dir != nil {
			dir.Do(c.dirInfoCMD)
		}
	}
}

func (c *Controller) Quit() {
	panic(0)
}

func (c *Controller) HandleDirEvent(event DirEvent) {
	if event.Err != nil {
		c.cli.Warn("%+v", event.Err)
	}

	if event.Path == c.cwd {
		c.main.List.UpdateRows(event.Rows)
		if !c.cwdInited {
			c.cwdInited = true
			if state, ok := c.listViewStates[c.cwd]; ok {
				c.main.SelectAt = findInRow(event, state.selected)
				c.main.ViewBeginAt = c.main.SelectAt - state.offset
			} else {
				c.main.SelectAt = 0
			}
		}
		c.main.Draw()
	} else if event.Path == c.parentCwd {
		c.left.List.UpdateRows(event.Rows)
		c.parentCwdInited = true
		cwdBase := filepath.Base(c.cwd)
		c.left.SelectAt = findInRow(event, cwdBase)
		c.left.Draw()
	}
}

func findInRow(event DirEvent, s string) int {
	for i, row := range event.Rows {
		if row.FileInfo.Name() == s {
			return i
		}
	}
	return 0
}

func (c *Controller) HandleMouseEvent(ev Event, keymap *map[Event][]Action, args []string) {
	me := ev.MouseEvent
	if me.S != 0 {
		// Scroll
		c.ScrollDown(-me.S*3, c.inLeft(me.X, me.Y))
		return
	}
	if me.Down {
		var e Event
		if me.Left {
			e = LeftClick.AsEvent()
		} else {
			e = RightClick.AsEvent()
		}

		e.MouseEvent = me
		HandleAction(e, keymap)
		return
	}
	if me.Double {
		e := DoubleClick.AsEvent()
		e.MouseEvent = me
		HandleAction(e, keymap)
		return
	}
}

func (c *Controller) ToggleDirInfo(cmds []string) {
	if c.dirInfoCMD != nil {
		c.dirInfoCMD = nil
	} else {
		c.dirInfoCMD = cmds
	}

	if dir := c.dirs.Get(c.cwd); dir != nil {
		dir.Do(append([]string{"reset_info"}, c.dirInfoCMD...))
	}
}

func (c *Controller) DirDo(cmds []string) {
	if dir := c.dirs.Get(c.cwd); dir != nil {
		dir.Do(cmds)
	}
}

func (c *Controller) Warn(format string, a ...any) {
	c.cli.Warn(format, a...)
}

func (c *Controller) CMD(s string, cursor int) {
	c.cli.CMD(s, cursor)
}
