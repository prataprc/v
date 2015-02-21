package buffer

import "fmt"

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

// LinesIterator will iterate on blocks of lines
// within edit-buffer.
type LinesIterator func() Lines

// LineIterator will iterate on each consecutive line
// within a block.
type LineIterator func() (start, end int64)

// return an iterator on block of consecutive-lines.
func (lines Lines) blocks() LinesIterator {
    lines.checkSanity(false)
    i := 0
    return LinesIterator(func() Lines {
        block := make(Lines, 0, 16)
        for i < len(lines) {
            start, end := lines[i], lines[i+1]
            block = append(block, start, end)
            i += 2
            if i < len(lines)  &&  end != lines[i] {
                return block
            }
        }
        if len(block) > 0 {
            return block
        }
        return nil
    })
}

// return an iterator on block of consecutive-lines,
// first block will contain bCur.
func (lines Lines) blocksFrom(bCur int64) LinesIterator {
    lines.checkSanity(false)
    iterBlock := lines.blocks()
    return LinesIterator(func() Lines {
        for {
            block := iterBlock()
            if bCur < 0 {
                return block
            } else if block != nil  {
                if block.containsCursor(bCur) || block.afterCursor(bCur) {
                    bCur = -1 // let us iterate on each block from now on.
                    return block
                }
            }
            // NOTE: continue to find the next block containing bCur.
        }
    })
}

// Lines will return an iterator for consecutive lines.
func (lines Lines) Lines(bCur int64) LineIterator {
    lines.checkSanity(false)
    from := 0
    if bCur > 0 {
        for ; from < len(lines); from += 2 {
            if (bCur >= lines[from]  &&  bCur < lines[from+1]) ||
               (bCur < lines[from]) {
                break
            }
        }
    }
    return LineIterator(func() (start, end int64) {
        prev := from
        if from > 0  &&  from < len(lines) {
            from += 2
            start, end := lines[prev], lines[prev+1]
            if from >= len(lines)  ||  lines[prev+1] != lines[from] {
                from = -1
            }
            return start, end
        }
        return -1, -1
    })
}

// merge a consecutive block of line, into lines.
func (lines Lines) mergeBlock(block Lines) Lines {
    block.checkSanity(true); lines.checkSanity(false)
    if len(block) == 0 {
        return lines
    }
    i, floorIdx, remaining := 0, 0, []int64{}
    a := block[0]
    for ; i < len(lines); i += 2 {
        if lines[i+1] >= a {
            z := block[len(block)-1]
            for i = i+2; i < len(lines); i += 2 {
                if lines[i] >= z {
                    remaining = lines[i:]
                    break
                }
            }
            lines = lines[:floorIdx+2]
            break
        }
        floorIdx = i
    }
    lines = append(lines, block...)
    lines = append(lines, remaining...)
    return lines
}

// return line index containing bCur, if the line that
// ought to contain bCur is missing, `ok` is false, return
// the index of first available line after bCur.
func (lines Lines) indexof(bCur int64) (i int64, ok bool) {
    for i := 0; i < len(lines); i += 2 {
        x, y := lines[i], lines[i+1]
        if x < 0 {
            panic(fmt.Errorf("impossible situation\n"))
        } else if y < 0 {
            panic(fmt.Errorf("impossible situation\n"))
        } else if x > y {
            panic(fmt.Errorf("impossible situation\n"))
        } else if bCur >= x && bCur < y {
            return int64(i), true
        } else if bCur < x {
            return int64(i), false
        }
    }
    return -1, false
}

// assuming that lines are contiguous, check wether bCur falls
// within the lines covered.
func (lines Lines) containsCursor(bCur int64) bool {
    lines.checkSanity(true)
    if bCur < 0 {
        return true
    } else if bCur >= lines[0]  &&  bCur < lines[len(lines)-1] {
        return true
    }
    return false
}

// return whether the lines start after the cursor.
func (lines Lines) afterCursor(bCur int64) bool {
    if len(lines) > 0 {
        return bCur < lines[0]
    }
    return false
}

// [ 40, 30, 30, 20, 20, 10 ] -> [ 10, 20, 20, 30, 30, 40 ]
func (lines Lines) reverse() {
    n := len(lines)
    for i, j := 0, n-1; i < (n / 2); i, j = i+1, j-1 {
        lines[j], lines[i] = lines[i], lines[j]
    }
}

// [ 30, 40, 20, 30, 10, 20 ] -> [ 10, 20, 20, 30, 30, 40 ]
func (lines Lines) flipReverse() {
    n := len(lines)
    for i, j := 0, n-1; i < (n / 2); i, j = i+2, j-2 {
        lines[j-1], lines[i] = lines[i], lines[j-1]
        lines[j], lines[i+1] = lines[i+1], lines[j]
    }
}

// check for various sanity of lines.
func (lines Lines) checkSanity(block bool) {
    if (len(lines) % 2) != 0 {
        panic(fmt.Errorf("impossible situation\n"))
    }
    if block && len(lines) > 0 {
        end := lines[1]
        for i := 2; i < len(lines); i+=2 {
            if end != lines[i] {
                panic(fmt.Errorf("lines are not contiguous\n"))
            }
            end = lines[i+1]
        }
    }
}
