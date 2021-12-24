package querysan

import (
	"strings"

	"github.com/ikawaha/kagome/tokenizer"
)

func words(s string) []string {
	t := tokenizer.New()
	// tokens := t.Tokenize(s)
	tokens := t.Analyze(s, tokenizer.Search)
	var a []string
	for _, token := range tokens {
		if token.Class == tokenizer.DUMMY {
			continue
		}
		features := token.Features()
		x := features[0]
		if x == "記号" {
			continue
		}
		a = append(a, strings.ToLower(token.Surface))
	}
	return a
}
