package buffer

import "unicode/utf8"
import "fmt"

var _ = fmt.Sprintf("dummy")

// figure out unicode boundary within `text`,
// if `reverse` is true, figure out the unicode
// boundary from backward.
func getRuneStart(text []byte, reverse bool) (int64, error) {
	// search for rune start in backward direction
	if reverse {
		for i := len(text) - 1; i >= 0; i-- {
			if utf8.RuneStart(text[i]) {
				return int64(i), nil
			}
		}

		// search for rune start in forward direction
	} else {
		for i, b := range text {
			if utf8.RuneStart(b) {
				return int64(i), nil
			}
		}
	}
	return 0, ErrorInvalidEncoding
}

// decode array of utf8 encoded bytes to unicode runes,
// stop decoding once `rn` number of runes are decoded.
// return the number of runes decoded, bytes consumed.
func bytes2NRunes(bs []byte, rn int64, acc []rune) (int64, int64, error) {
	size, count := int64(0), int64(0)
	for size < int64(len(bs)) {
		r, sz := utf8.DecodeRune(bs[size:])
		if r == utf8.RuneError {
			return 0, 0, ErrorInvalidEncoding
		}
		acc[count] = r
		size, count = size+int64(sz), count+1
		if count >= rn {
			return count, size, nil
		}
	}
	return count, size, nil
}

// decode utf8 encoded bytes to unicode runes.
func bytes2Runes(bs []byte) []rune {
	runes := make([]rune, len(bs))
	i, n := 0, 0
	for ; i < len(bs); n++ {
		r, sz := utf8.DecodeRune(bs[i:])
		if r == utf8.RuneError {
			panic(ErrorInvalidEncoding)
		}
		runes[n] = r
		i += sz
	}
	return runes[:n]
}

// encode unicode runes to utf8 encoded bytes.
func runes2Bytes(rs []rune) []byte {
	bytes := make([]byte, len(rs)*4)
	off := 0
	for _, r := range rs {
		sz := utf8.EncodeRune(bytes[off:], r)
		off += sz
	}
	return bytes[:off]
}

func runePositions(bs []byte) []int64 {
	offs := make([]int64, len(bs))
	i, n := 0, 0
	for ; i < len(bs); n++ {
		r, sz := utf8.DecodeRune(bs[i:])
		if r == utf8.RuneError {
			panic(utf8.RuneError)
		}
		offs[n] = int64(i)
		i += sz
	}
	return offs[:n]
}

func reverseRunes(runes []rune) []rune {
	reversed := make([]rune, len(runes))
	for i, j := 0, len(runes)-1; i < len(runes); i, j = i+1, j-1 {
		reversed[j] = runes[i]
	}
	return reversed
}
