package ngram

import "testing"

func TestTokenizerTokenize(t *testing.T) {
	tokenizer := NewTokenizer(2)
	tokens := tokenizer.Tokenize("静态博客 search API")
	if len(tokens) == 0 {
		t.Fatalf("expected tokens")
	}

	foundChinese := false
	foundLatin := false
	for _, token := range tokens {
		if token == "静态" {
			foundChinese = true
		}
		if token == "search" {
			foundLatin = true
		}
	}

	if !foundChinese {
		t.Fatalf("expected chinese ngram token")
	}
	if !foundLatin {
		t.Fatalf("expected latin token")
	}
}
