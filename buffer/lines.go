package buffer

import "fmt"

// BuildBlock around specified cursor position,
// if `bCur` is -1 use current cursor position.
// Return Lines of specified width*2 + 1.
func (ebuf *EditBuffer) BuildBlock(bCur int64, width int64) Lines {
	if bCur < 0 {
		bCur = ebuf.dot
	}
	lnNl := int64(len(ebuf.newline))
	lines := make(Lines, 0, (width*2+1)*2+2)
	// gather lines above cursor.
	iter := Find(ebuf.reNlR, ebuf.buffer.BackStreamFrom(bCur))
	for i, end := int64(0), int64(-1); i < width+1; i++ {
		loc := iter()
		if loc != nil {
			lines = append(lines, end, int64(loc[0])+lnNl)
			end = int64(loc[1])
		} else {
			lines = append(lines, end, 0)
			break
		}
	}
	lines.reverse()
	// gather lines below cursor.
	iter = Find(ebuf.reNlR, ebuf.buffer.StreamFrom(bCur))
	for i, start := int64(0), int64(-1); i < width; i++ {
		loc := iter()
		if loc != nil {
			lines = append(lines, start, int64(loc[1]))
			start = int64(loc[1])
		} else {
			l, err := ebuf.buffer.Length()
			if err != nil {
				err = fmt.Errorf("impossible situation: %v\n", err)
				panic(err)
			}
			lines = append(lines, start, l)
			break
		}
	}
	// consolidate.
	for i, j := 1, 1; (i + 1) < len(lines); i += 2 {
		if lines[i] != lines[i+1] {
			panic("impossible situation")
		} else if lines[i] == -1 && lines[i+1] == -1 {
			continue
		}
		lines[j], lines[j+1] = lines[i], lines[i+1]
		j += 2
	}
	return lines
}

// Line returns the line starting from `start` and ending
// before `end`.
func (ebuf *EditBuffer) Line(bCur int64) (start int64, end int64) {
	i := ebuf.lines.indexof(bCur)
	if i == -1 {
		return -1, -1
	}
	start, end = ebuf.lines[i], ebuf.lines[i+1]
	return start, end
}

// Lines returns a block of Lines around the line containing
// bCur.
func (ebuf *EditBuffer) Lines(bCur int64, width int64) (ls Lines) {
	idx := ebuf.lines.indexof(bCur)
	if idx == -1 {
		ebuf.lines = ebuf.lines.merge(ebuf.BuildBlock(bCur, width))
		idx = ebuf.lines.indexof(bCur)
	}
	if idx == -1 {
		panic(fmt.Errorf("impossible situation\n"))
	}
	start := idx - width*2
	if start < 0 {
		start = 0
	}
	n := (width*2 + 1) * 2
	lines := make(Lines, 0, n)
	for i, j := start, int64(0); j < n; i, j = i+2, j+2 {
		lines[j], lines[j+1] = ebuf.lines[j], ebuf.lines[i+1]
	}
	return lines
}

// Lines index pairs within the input buffer,
// eg lines[2*n:2*n+1] identifies the indexes of the
// nth line starting from lines[2*n] and ending
// before lines[2*n+1].
//
// Empty-buffer:
//      [0, 0]
// Empty-line:
//      [..., x, x, ...]
// First-line:
//      [0, x, ...]
// Last-line:
//      [..., x, n+1]
type Lines []int64

func (lines Lines) indexof(bCur int64) int64 {
	if len(lines)%2 != 0 {
		panic(fmt.Errorf("impossible situation\n"))
	}
	for i := 0; i <= len(lines); i += 2 {
		x, y := lines[i], lines[i+1]
		if x < 0 {
			panic(fmt.Errorf("impossible situation\n"))
		} else if y < 0 {
			panic(fmt.Errorf("impossible situation\n"))
		} else if x > y {
			panic(fmt.Errorf("impossible situation\n"))
		} else if bCur >= x && bCur < y {
			return int64(i)
		} else if bCur < x {
			return -1
		}
	}
	return -1
}

func (lines Lines) reverse() {
	n := len(lines)
	for i, j := 0, n-1; i < (n / 2); i, j = i+1, j-1 {
		lines[j], lines[i] = lines[i], lines[j]
	}
}

func (lines Lines) merge(ls Lines) Lines {
	if len(ls)%2 != 0 {
		panic(fmt.Errorf("impossible situation\n"))
	} else if len(ls) == 0 {
		return lines
	}
	i, floorIdx, remaining := 0, 0, []int64{}
	for ; i < len(lines); i += 2 {
		if ls[0] < lines[i+1] {
			z := ls[len(ls)-1]
			for ; i < len(lines); i += 2 {
				if z <= lines[i] {
					remaining = lines[i:]
					break
				}
			}
			lines = lines[:floorIdx]
			break
		}
		floorIdx = i
	}
	lines = append(lines, ls...)
	lines = append(lines, remaining...)
	return lines
}
