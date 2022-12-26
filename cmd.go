package main

import (
	"log"
	"strconv"
)

type CMD interface{}

func ToAction(cmd CMD, args []string) Action {
	switch f := cmd.(type) {
	case func():
		return func(ev Event, keymap *map[Event][]Action) {
			f()
		}
	case func([]string):
		return func(ev Event, keymap *map[Event][]Action) {
			f(args)
		}
	case func(ev Event, keymap *map[Event][]Action, args []string):
		return func(ev Event, keymap *map[Event][]Action) {
			f(ev, keymap, args)
		}
	}
	log.Fatalf("unsupported cmd type %T", cmd)

	// Unreachable
	return nil
}

func ParseArgOrDefault(args []string, pos int, default_ any) any {
	if pos < 0 || pos >= len(args) {
		return default_
	}
	arg := args[pos]
	switch default_.(type) {
	case string:
		return arg
	case int:
		i, err := strconv.Atoi(arg)
		if err != nil {
			return default_
		}
		return i
	default:
		log.Fatalf("not support type %T", default_)
	}

	// unreachable
	return nil
}
