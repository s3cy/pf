package main

import (
	"fmt"
	"strings"
)

type CLI struct {
	c         *Controller
	oldKeymap map[Event][]Action
	keymap    map[Event][]Action
	cmds      map[string]CMD

	prevCMD string
	cmd     string
	cursor  int
}

func NewCLI(c *Controller) *CLI {
	cli := &CLI{
		c:      c,
		keymap: make(map[Event][]Action),
	}
	add := func(ev Event, cmd CMD) {
		cli.keymap[ev] = []Action{ToAction(cmd, nil)}
	}

	for i := 0; i < 26; i++ {
		add(Key(rune(i+'a')), cli.Add)
		add(Key(rune(i+'A')), cli.Add)
	}
	add(Key(rune(' ')), cli.Add)
	add(BSpace.AsEvent(), cli.Delete)
	add(CtrlU.AsEvent(), cli.Clear)

	add(ESC.AsEvent(), cli.Cancel)
	add(CtrlC.AsEvent(), cli.Cancel)
	// enter, return key
	add(CtrlM.AsEvent(), cli.Enter)

	add(CtrlB.AsEvent(), cli.CursorBack)
	add(CtrlF.AsEvent(), cli.CursorForward)
	add(CtrlA.AsEvent(), cli.CursorBegin)
	add(CtrlE.AsEvent(), cli.CursorEnd)

	return cli
}

func (c *CLI) SetCMDs(cmds map[string]CMD) {
	c.cmds = cmds
}

func (c *CLI) StartCMD(ev Event, keymap *map[Event][]Action, _ []string) {
	c.oldKeymap = *keymap
	*keymap = c.keymap

	c.Add(Key(':'), nil, nil)
}

func (c *CLI) StartFilter(ev Event, keymap *map[Event][]Action, _ []string) {
	c.oldKeymap = *keymap
	*keymap = c.keymap

	c.Add(Key('/'), nil, nil)
}

func (c *CLI) Add(ev Event, _ *map[Event][]Action, _ []string) {
	c.cmd = c.cmd[:c.cursor] + string(ev.Char) + c.cmd[c.cursor:]
	c.cursor++
	c.draw()
}

func (c *CLI) Delete(ev Event, keymap *map[Event][]Action, args []string) {
	if c.cursor > 1 {
		c.cmd = c.cmd[:c.cursor-1] + c.cmd[c.cursor:]
		c.cursor--
	} else {
		c.Cancel(ev, keymap, args)
	}
	c.draw()
}

func (c *CLI) Clear() {
	c.cmd = string(c.prevCMD[0])
	c.cursor = 1
	c.draw()
}

func (c *CLI) CursorBack() {
	if c.cursor > 1 {
		c.cursor--
	}
	c.draw()
}

func (c *CLI) CursorForward() {
	if c.cursor < len(c.cmd) {
		c.cursor++
	}
	c.draw()
}

func (c *CLI) CursorBegin() {
	c.cursor = 1
	c.draw()
}

func (c *CLI) CursorEnd() {
	c.cursor = len(c.cmd)
	c.draw()
}

func (c *CLI) Cancel(ev Event, keymap *map[Event][]Action, args []string) {
	c.cmd = ""
	c.cursor = 0
	*keymap = c.oldKeymap
	c.draw()
}

func (c *CLI) Enter(ev Event, keymap *map[Event][]Action, args []string) {
	mode := c.cmd[0]
	spec := c.cmd[1:]

	c.prevCMD = ""
	c.cmd = ""
	c.cursor = 0
	*keymap = c.oldKeymap
	c.c.CMD(c.cmd, c.cursor)

	if mode == ':' {
		tokens := strings.Split(spec, " ")
		cmd, ok := c.cmds[tokens[0]]
		if ok {
			action := ToAction(cmd, tokens[1:])
			action(ev, keymap)
		} else {
			c.c.Warn("Command '%s' not found", tokens[0])
		}
	}
}

func (c *CLI) draw() {
	if c.prevCMD == c.cmd {
		return
	}
	prevCMD := c.prevCMD
	c.prevCMD = c.cmd
	if c.cmd == "" {
		if prevCMD[0] == '/' {
			c.c.DirDo([]string{"filter"})
		}
	} else if c.cmd[0] == '/' {
		c.c.DirDo([]string{fmt.Sprintf("filter %s", c.cmd[1:])})
	}
	c.c.CMD(c.cmd, c.cursor)
}
