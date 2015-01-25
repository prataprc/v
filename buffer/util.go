package buffer

import "unicode/utf8"

func getRuneStart(text []byte) (int64, error) {
	for i, b := range text {
		if utf8.RuneStart(b) {
			return int64(i), nil
		}
	}
	return 0, ErrorInvalidEncoding
}

func bytes2RunesN(bs []byte, rn int64, acc []rune) (int64, int64, error) {
	n, count := int64(0), int64(0)
	for n < int64(len(bs)) {
		r, sz := utf8.DecodeRune(bs[n:])
		if r == utf8.RuneError {
			return 0, 0, ErrorInvalidEncoding
		}
		acc = append(acc, r)
		n, count = n+int64(sz), count+1
		if count >= rn {
			return count, n, nil
		}
	}
	return count, n, nil
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
