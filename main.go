package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func initLog() {
	f, err := os.Create("log")
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	log.SetOutput(f)
}

func main() {
	initLog()

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	marks := make(map[string]struct{})
	printMarks := func() {
		for mark := range marks {
			fmt.Println(mark)
		}
	}

	quit := func() {
		maybePanic := recover()
		s.Fini()
		if maybePanic != nil {
			if code, ok := maybePanic.(int); ok && code == 0 {
				printMarks()
				return
			}
			panic(maybePanic)
		}
	}
	defer quit()

	dirs := NewDirSet(StyleM)
	c := NewController(dirs, marks, s)
	cli := NewCLI(c)
	cmds := initBuiltinCMDTable(c, cli)
	cli.SetCMDs(cmds)
	keybindings := initBuiltinKeybindings()
	keymap := initBuiltinKeymap(cmds)
	ParseKeymap(keymap, cmds, keybindings)

	eventCh := make(chan Event, 1)
	go func() {
		hdr := TcellEventHandler{}
		for {
			eventCh <- hdr.Handle(s)
		}
	}()
	for {
		c.Show()

		select {
		case ev := <-eventCh:
			HandleAction(ev, &keymap)
		case dirEvent := <-dirs.Event():
			c.HandleDirEvent(dirEvent)
		}
	}
}

func initBuiltinKeymap(cmds map[string]CMD) map[Event][]Action {
	keymap := make(map[Event][]Action)
	add := func(et EventType, cmdName string) {
		cmd, ok := cmds[cmdName]
		if !ok {
			log.Fatalf("unknown cmd name: %s", cmdName)
		}
		keymap[et.AsEvent()] = []Action{ToAction(cmd, nil)}
	}

	add(Resize, "resize")
	add(Mouse, "mouse")
	return keymap
}

func initBuiltinCMDTable(c *Controller, cli *CLI) map[string]CMD {
	scroll := func(up bool) func(args []string) {
		return func(args []string) {
			i := ParseArgOrDefault(args, 0, 1)
			left := ParseArgOrDefault(args, 1, false)
			c.ScrollDown(i.(int), left.(bool))
		}
	}

	mouse := func(action func(x, y int)) func(Event, *map[Event][]Action, []string) {
		return func(ev Event, _ *map[Event][]Action, _ []string) {
			me := ev.MouseEvent
			action(me.X, me.Y)
		}
	}

	return map[string]CMD{
		"resize":          c.Resize,
		"next":            c.Next,
		"prev":            c.Prev,
		"top":             c.Top,
		"bottom":          c.Bottom,
		"scroll_down":     scroll(false),
		"scroll_up":       scroll(true),
		"half_page_down":  c.HalfPageDown,
		"half_page_up":    c.HalfPageUp,
		"page_down":       c.PageDown,
		"page_up":         c.PageUp,
		"clear_marks":     c.ClearMarks,
		"mark":            c.Mark,
		"unmark":          c.Unmark,
		"toggle_mark":     c.ToggleMark,
		"select":          mouse(c.Select),
		"goto":            mouse(c.Goto),
		"in":              c.In,
		"out":             c.Out,
		"quit":            c.Quit,
		"mouse":           c.HandleMouseEvent,
		"toggle_dir_info": c.ToggleDirInfo,
		"dir":             c.DirDo,
		"command":         cli.StartCMD,
		"filter":          cli.StartFilter,
	}
}

func initBuiltinKeybindings() string {
	binds := []string{
		"ctrl-l:resize",
		"j:next",
		"k:prev",
		"g:top",
		"G:bottom",
		"ctrl-e:scroll_down 1",
		"ctrl-y:scroll_up 1",
		"ctrl-d:half_page_down",
		"ctrl-u:half_page_up",
		"space:toggle_mark",
		"tab:toggle_mark+next",
		"shift-tab:toggle_mark+prev",
		"enter:clear_marks+mark+quit",
		"l:in",
		"h:out",
		"ctrl-c:quit",
		"q:quit",
		"left-click:select",
		"double-click:goto",
		"right-click:out",
		"i:toggle_dir_info perm hsize mtime link_target",
		"s:dir sort_by_size",
		"::command",
		"/:filter",
	}
	return strings.Join(binds, ",")
}
