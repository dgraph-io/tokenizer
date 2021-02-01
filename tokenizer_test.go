package tokenizer

import (
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/blevesearch/bleve/analysis"
	"github.com/blevesearch/bleve/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/analysis/token/lowercase"
	"github.com/blevesearch/bleve/analysis/token/unicodenorm"
	"github.com/blevesearch/bleve/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/registry"
	"github.com/go-nlp/bpe"
	"github.com/go-nlp/corpus"
	"github.com/stretchr/testify/require"
)

func testingTokenizer() (*Tokenizer, error) {
	f, err := os.Open("testdata/corpus_zh.txt")
	if err != nil {
		return nil, err
	}

	normalizer := NewNormalizer()
	enNorm := func(a string) string {
		a = normalizer.MustNorm(a)
		return strings.ToLower(a)
	}
	norm := func(a string) string {
		//return strings.Replace(enNorm(a), " ", "", -1)
		return a
	}

	c, err := corpus.FromTextCorpus(f, nil, norm)
	if err != nil {
		return nil, err
	}
	log.Printf("corpus %d", c.Size())
	g, err := os.Open("testdata/corpus_en.txt")
	if err != nil {
		return nil, err
	}

	c2, err := corpus.FromTextCorpus(g, nil, enNorm)
	if err != nil {
		return nil, err
	}
	c.Merge(c2)

	enc, err := bpe.Learn(c, 6000, 2, false)
	if err != nil {
		return nil, err
	}

	tok := NewTokenizer(enc)
	return tok, nil
}

func TestTokenizer(t *testing.T) {
	tok, err := testingTokenizer()
	require.NoError(t, err)
	input := "我现在在吃rabbit soup"
	tokens, err := tok.Tokenize(input)
	require.NoError(t, err)
	t.Logf("input %q. tokens %q", input, tokens)
}

// uniqueTerms takes a token stream and returns a string slice of unique terms.
func uniqueTerms(tokens analysis.TokenStream) []string {
	var terms []string
	for i := range tokens {
		terms = append(terms, string(tokens[i].Term))
	}
	terms = RemoveDuplicates(terms)
	return terms
}

// RemoveDuplicates sorts the slice of strings and removes duplicates. changes the input slice.
// This function should be called like: someSlice = RemoveDuplicates(someSlice)
func RemoveDuplicates(s []string) (out []string) {
	sort.Strings(s)
	out = s[:0]
	for i := range s {
		if i > 0 && s[i] == s[i-1] {
			continue
		}
		out = append(out, s[i])
	}
	return
}

func BenchmarkBleve(b *testing.B) {
	b.StopTimer()
	// set up
	unicodenormName := "unicodenorm_nfkc"
	bleveCache := registry.NewCache()
	_, err := bleveCache.DefineTokenFilter(unicodenormName,
		map[string]interface{}{
			"type": unicodenorm.Name,
			"form": unicodenorm.NFKC,
		})
	termAnalyzer, err := bleveCache.DefineAnalyzer("term",
		map[string]interface{}{
			"type":      custom.Name,
			"tokenizer": unicode.Name,
			"token_filters": []string{
				lowercase.Name,
				unicodenormName,
			},
		})
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.StartTimer()
	var tokens []string
	for i := 0; i < b.N; i++ {
		tks := termAnalyzer.Analyze([]byte("我现在在吃rabbit stew"))
		tokens = uniqueTerms(tks)
	}
	_ = tokens

}
func BenchmarkTokenizer(b *testing.B) {
	b.StopTimer()
	// set up
	f, err := os.Open("testdata/corpus_zh.txt")
	if err != nil {
		b.Fatal(err)
	}
	normalizer := func(a string) string {
		return strings.Replace(a, " ", "", -1)
	}

	c, err := corpus.FromTextCorpus(f, nil, normalizer)
	if err != nil {
		b.Fatal(err)
	}
	enc, err := bpe.Learn(c, 6000, 3, false)
	if err != nil {
		b.Fatal(err)
	}
	tokz := NewTokenizer(enc)

	b.ResetTimer()
	b.StartTimer()
	var tokens []string
	for i := 0; i < b.N; i++ {
		tokens, _ = tokz.Tokenize("我现在在吃rabbit stew")
	}
	_ = tokens

}
