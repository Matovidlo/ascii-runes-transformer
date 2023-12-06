package asciimap

import (
	"unicode/utf8"

	"golang.org/x/text/transform"
)

const runeErrorString = string(utf8.RuneError)

// Transformer implements the transform.Transformer interface.
type Transformer struct {
	t transform.SpanningTransformer
}

func (t Transformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	return t.t.Transform(dst, src, atEOF)
}

func (t Transformer) Span(b []byte, atEOF bool) (n int, err error) {
	return t.t.Span(b, atEOF)
}

func (t Transformer) Reset() { t.t.Reset() }

// Map returns a Transformer that maps the runes in the input using the given
// mapping. Illegal bytes in the input are converted to utf8.RuneError before
// being passed to the mapping func.
func Map(mapping func(rune) string) Transformer {
	return Transformer{mapper(mapping)}
}

type mapper func(rune) string

func (mapper) Reset() {}

func (t mapper) Span(src []byte, atEOF bool) (n int, err error) {
	for r, size := rune(0), 0; n < len(src); n += size {
		if r = rune(src[n]); r < utf8.RuneSelf {
			size = 1
		} else if r, size = utf8.DecodeRune(src[n:]); size == 1 {
			// Invalid rune.
			if !atEOF && !utf8.FullRune(src[n:]) {
				err = transform.ErrShortSrc
			} else {
				err = transform.ErrEndOfSpan
			}
			break
		}

		for _, transformedRune := range t(r) {
			if transformedRune != r {
				err = transform.ErrEndOfSpan
			}
		}

		if err != nil {
			break
		}
	}
	return n, err
}

func (t mapper) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var replacement string
	var b [utf8.UTFMax]byte

	for r, size := rune(0), 0; nSrc < len(src); {
		if r = rune(src[nSrc]); r < utf8.RuneSelf {
			replacement = t(r)
			for _, repl := range replacement {
				size = utf8.RuneLen(r)
				if nDst == len(dst) {
					err = transform.ErrShortDst
					break
				}

				nDst += utf8.EncodeRune(dst[nDst:], repl)
				nSrc += size
			}
			if err != nil {
				break
			}

			size = 1
			continue
		} else if r, size = utf8.DecodeRune(src[nSrc:]); size == 1 {
			// Invalid rune.
			if !atEOF && !utf8.FullRune(src[nSrc:]) {
				err = transform.ErrShortSrc
				break
			}

			if replacement = t(utf8.RuneError); len(replacement) == 3 {
				if nDst+3 > len(dst) {
					err = transform.ErrShortDst
					break
				}
				dst[nDst+0] = runeErrorString[0]
				dst[nDst+1] = runeErrorString[1]
				dst[nDst+2] = runeErrorString[2]
				nDst += 3
				nSrc++
				continue
			}
		} else if replacement = t(r); len(replacement) > 1 {
			if nDst+size > len(dst) {
				err = transform.ErrShortDst
				break
			}

			for _, repl := range replacement {
				if nDst+size == len(dst) {
					err = transform.ErrShortDst
					break
				}

				nDst += utf8.EncodeRune(dst[nDst:], repl)
			}

			nSrc += size
			if err != nil {
				break
			}

			continue
		}

		for _, repl := range replacement {
			n := utf8.EncodeRune(b[:], repl)

			if nDst+n > len(dst) {
				err = transform.ErrShortDst
				break
			}
			for i := 0; i < n; i++ {
				dst[nDst] = b[i]
				nDst++
			}
			nSrc += size
		}
	}
	return
}

// Ascii256Toascii128 maps given rune `r` that is in both ascii 256
// and unicode.Ll or unicode.Lu categories. This conversion does not support
// stripping of unicode.Mn as it can be done using norm.NFD normalization.
func Ascii256Toascii128(r rune) string {
	switch r {
	case 'Œ':
		return "OE"

	case 'œ':
		return "oe"

	case 'µ':
		return "m"

	case 'Æ':
		return "AE"

	case 'Ð':
		return "D"

	case 'Ø':
		return "O"

	case 'ß':
		return "ss"

	case 'æ':
		return "ae"

	case 'ð':
		return "d"

	case 'ø':
		return "o"
	}

	return string(r)
}
