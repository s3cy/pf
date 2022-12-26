package main

type ListView struct {
	Win         *Win
	List        List
	SelectAt    int
	ViewBeginAt int
}

func (v *ListView) Draw() {
	v.Win.Reset(v.List.Style)

	if v.SelectAt < 0 {
		v.SelectAt = 0
	}
	size := v.List.Size()
	if size == 0 {
		return
	}
	if v.SelectAt >= size {
		v.SelectAt = size - 1
	}

	if v.SelectAt < v.ViewBeginAt {
		v.ViewBeginAt = v.SelectAt
	}
	if v.SelectAt > v.ViewBeginAt+v.Win.H() {
		v.ViewBeginAt = v.SelectAt - v.Win.H()
	}

	idx := v.ViewBeginAt
	for row := 0; row <= v.Win.H(); row++ {
		if idx >= size {
			break
		}
		item := v.List.Get(idx, v.Win.W(), idx == v.SelectAt)
		if len(item) != 0 {
			v.Win.Render(0, row, item, v.List.Style, false)
		}
		idx++
	}
}

func (v *ListView) ScrollDown(n int) bool {
	v.ViewBeginAt += n
	if v.ViewBeginAt < 0 {
		v.ViewBeginAt = 0
	}
	size := v.List.Size()
	if v.ViewBeginAt >= size {
		v.ViewBeginAt = size - 1
	}

	var selectChanged bool
	if v.SelectAt < v.ViewBeginAt {
		v.SelectAt = v.ViewBeginAt
		selectChanged = true
	}
	rows := v.Win.H()
	if v.SelectAt > v.ViewBeginAt+rows {
		v.SelectAt = v.ViewBeginAt + rows
		selectChanged = true
	}
	return selectChanged
}

func (v *ListView) Select(y int) {
	v.SelectAt = v.ViewBeginAt + y - v.Win.Y1
}

func (v *ListView) Mark(unmark, toggle bool) {
	info := v.List.GetFileInfo(v.SelectAt)
	if info == nil {
		return
	}

	path := info.Path
	if toggle {
		_, ok := v.List.Marks[path]
		unmark = ok
	}

	if unmark {
		delete(v.List.Marks, path)
		return
	}

	v.List.Marks[path] = struct{}{}
}
