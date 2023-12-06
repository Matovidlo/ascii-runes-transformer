package asciimap

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/transform"
)

type transformTest struct {
	desc    string
	szDst   int
	atEOF   bool
	repl    string
	in      string
	out     string // result string of first call to Transform
	outFull string // transform of entire input string
	err     error
	errSpan error
	nSpan   int

	t transform.SpanningTransformer
}

const large = 10240

func (tt *transformTest) check(t *testing.T, i int) {
	dst := make([]byte, tt.szDst)
	src := []byte(tt.in)
	nDst, nSrc, err := tt.t.Transform(dst, src, tt.atEOF)
	assert.Equal(t, tt.err, err)
	got := string(dst[:nDst])
	assert.Equal(t, tt.out, got)
	// Calls tt.t.Transform for the remainder of the input. We use this to test
	// the nSrc return value.
	out := make([]byte, large)
	n := copy(out, dst[:nDst])
	nDst, _, _ = tt.t.Transform(out[n:], src[nSrc:], true)
	got = string(out[:n+nDst])
	assert.Equal(t, tt.outFull, got)
	tt.t.Reset()
	p := 0
	for ; p < len(tt.in) && p < len(tt.outFull) && tt.in[p] == tt.outFull[p]; p++ {
	}
	if tt.nSpan != 0 {
		p = tt.nSpan
	}

	n, err = tt.t.Span([]byte(tt.in), tt.atEOF)
	assert.Equal(t, tt.errSpan, err)
	assert.Equal(t, p, n)
}

func idem(r rune) string { return string(r) }

func TestMap(t *testing.T) {
	runes := []rune{'a', 'ç', '中', '\U00012345', 'a'}
	// Default mapper used for this test.
	rotate := Map(func(r rune) string {
		for i, m := range runes {
			if m == r {
				return string(runes[i+1])
			}
		}
		return string(r)
	})

	for i, tt := range []transformTest{{
		desc:    "empty",
		szDst:   large,
		atEOF:   true,
		in:      "",
		out:     "",
		outFull: "",
		t:       rotate,
	}, {
		desc:    "no change",
		szDst:   1,
		atEOF:   true,
		in:      "b",
		out:     "b",
		outFull: "b",
		t:       rotate,
	}, {
		desc:    "short dst",
		szDst:   2,
		atEOF:   true,
		in:      "aaaa",
		out:     "ç",
		outFull: "çççç",
		err:     transform.ErrShortDst,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "short dst ascii, no change",
		szDst:   2,
		atEOF:   true,
		in:      "bbb",
		out:     "bb",
		outFull: "bbb",
		err:     transform.ErrShortDst,
		t:       rotate,
	}, {
		desc:    "short dst writing error",
		szDst:   2,
		atEOF:   false,
		in:      "a\x80",
		out:     "ç",
		outFull: "ç\ufffd",
		err:     transform.ErrShortDst,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "short dst writing incomplete rune",
		szDst:   2,
		atEOF:   true,
		in:      "a\xc0",
		out:     "ç",
		outFull: "ç\ufffd",
		err:     transform.ErrShortDst,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		// TODO: actually smaller buffer needed.
		desc:    "short dst, longer",
		szDst:   5,
		atEOF:   true,
		in:      "Hellø",
		out:     "Hello",
		outFull: "Hello",
		errSpan: transform.ErrEndOfSpan,
		t:       Map(Ascii256Toascii128),
	}, {
		desc:    "short dst, single",
		szDst:   1,
		atEOF:   false,
		in:      "ø",
		out:     "",
		outFull: "ø",
		err:     transform.ErrShortDst,
		t:       Map(idem),
	}, {
		desc:    "short dst, longer, writing error",
		szDst:   8,
		atEOF:   false,
		in:      "\x80Hello\x80",
		out:     "\ufffdHello",
		outFull: "\ufffdHello\ufffd",
		err:     transform.ErrShortDst,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "short src",
		szDst:   2,
		atEOF:   false,
		in:      "a\xc2",
		out:     "ç",
		outFull: "ç\ufffd",
		err:     transform.ErrShortSrc,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "invalid input, atEOF",
		szDst:   large,
		atEOF:   true,
		in:      "\x80",
		out:     "\ufffd",
		outFull: "\ufffd",
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "invalid input, !atEOF",
		szDst:   large,
		atEOF:   false,
		in:      "\x80",
		out:     "\ufffd",
		outFull: "\ufffd",
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "incomplete rune !atEOF",
		szDst:   large,
		atEOF:   false,
		in:      "\xc2",
		out:     "",
		outFull: "\ufffd",
		err:     transform.ErrShortSrc,
		errSpan: transform.ErrShortSrc,
		t:       rotate,
	}, {
		desc:    "invalid input, incomplete rune atEOF",
		szDst:   large,
		atEOF:   true,
		in:      "\xc2",
		out:     "\ufffd",
		outFull: "\ufffd",
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "misc correct",
		szDst:   large,
		atEOF:   true,
		in:      "a\U00012345 ç!",
		out:     "ça 中!",
		outFull: "ça 中!",
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "misc correct and invalid",
		szDst:   large,
		atEOF:   true,
		in:      "Hello\x80 w\x80orl\xc0d!\xc0",
		out:     "Hello\ufffd w\ufffdorl\ufffdd!\ufffd",
		outFull: "Hello\ufffd w\ufffdorl\ufffdd!\ufffd",
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "misc correct and invalid, short src",
		szDst:   large,
		atEOF:   false,
		in:      "Hello\x80 w\x80orl\xc0d!\xc2",
		out:     "Hello\ufffd w\ufffdorl\ufffdd!",
		outFull: "Hello\ufffd w\ufffdorl\ufffdd!\ufffd",
		err:     transform.ErrShortSrc,
		errSpan: transform.ErrEndOfSpan,
		t:       rotate,
	}, {
		desc:    "misc correct and invalid, short src, replacing RuneError",
		szDst:   large,
		atEOF:   false,
		in:      "Hel\ufffdlo\x80 w\x80orl\xc0d!\xc2",
		out:     "Hel?lo? w?orl?d!",
		outFull: "Hel?lo? w?orl?d!?",
		errSpan: transform.ErrEndOfSpan,
		err:     transform.ErrShortSrc,
		t: Map(func(r rune) string {
			if r == utf8.RuneError {
				return "?"
			}

			return string(r)
		}),
	}} {
		tt.check(t, i)
	}
}

func TestMapAlloc(t *testing.T) {
	if n := testing.AllocsPerRun(3, func() {
		Map(idem).Transform(nil, nil, false)
	}); n > 0 {
		t.Errorf("got %f; want 0", n)
	}
}
