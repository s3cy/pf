package main

import "log"

type Action func(ev Event, keymap *map[Event][]Action)

func HandleAction(ev Event, keymap *map[Event][]Action) {
	actions, ok := (*keymap)[ev.Comparable()]
	if ok {
		for _, action := range actions {
			action(ev, keymap)
		}
	} else {
		log.Printf("unknown event: %+v\n", ev)
	}
}
