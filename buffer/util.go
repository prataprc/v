package buffer

import "unicode/utf8"
import "fmt"

var _ = fmt.Sprintf("dummy")

func getRuneStart(text []byte, reverse bool) (int64, error) {
	if reverse { // search for rune start in backward direction
		for i := len(text) - 1; i >= 0; i-- {
			if utf8.RuneStart(text[i]) {
				return int64(i), nil
			}
		}

	} else { // search for rune start in forward direction
		for i, b := range text {
			if utf8.RuneStart(b) {
				return int64(i), nil
			}
		}
	}
	return 0, ErrorInvalidEncoding
}

func bytes2RunesN(bs []byte, rn int64, acc []rune) (int64, int64, error) {
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

func bytes2Runes(bs []byte) ([]rune, error) {
	runes := make([]rune, 0, len(bs)/8)
	for i := 0; i < len(bs); {
		r, sz := utf8.DecodeRune(bs[i:])
		if r == utf8.RuneError {
			return nil, ErrorInvalidEncoding
		}
		runes = append(runes, r)
		i = i + sz
	}
	return runes, nil
}
