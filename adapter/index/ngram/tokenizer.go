package ngram

import (
	"strings"
	"unicode"
)

type Tokenizer struct {
	n int
}

func NewTokenizer(n int) *Tokenizer {
	if n < 1 {
		n = 2
	}
	return &Tokenizer{n: n}
}

func (t *Tokenizer) Tokenize(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var latinBuf []rune
	var hanBuf []rune
	out := make([]string, 0)
	seen := make(map[string]struct{})

	flushLatin := func() {
		if len(latinBuf) == 0 {
			return
		}
		token := strings.ToLower(string(latinBuf))
		if token != "" {
			if _, ok := seen[token]; !ok {
				seen[token] = struct{}{}
				out = append(out, token)
			}
		}
		latinBuf = latinBuf[:0]
	}
	flushHan := func() {
		if len(hanBuf) == 0 {
			return
		}
		if len(hanBuf) < t.n {
			token := string(hanBuf)
			if _, ok := seen[token]; !ok {
				seen[token] = struct{}{}
				out = append(out, token)
			}
			hanBuf = hanBuf[:0]
			return
		}
		for i := 0; i+t.n <= len(hanBuf); i++ {
			token := string(hanBuf[i : i+t.n])
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			out = append(out, token)
		}
		hanBuf = hanBuf[:0]
	}

	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			flushLatin()
			hanBuf = append(hanBuf, r)
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			flushHan()
			latinBuf = append(latinBuf, unicode.ToLower(r))
		default:
			flushLatin()
			flushHan()
		}
	}
	flushLatin()
	flushHan()

	return out
}
