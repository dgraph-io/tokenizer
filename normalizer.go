package tokenizer

import (
	"bytes"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var mn = runes.In(unicode.Mn)

// Normalizer is just a byte buffer
type Normalizer struct {
	src   bytes.Buffer
	dst   bytes.Buffer
	trans transform.Transformer
}

// NewNormalizer creates a new Normalizer
func NewNormalizer() *Normalizer {
	return &Normalizer{
		trans: transform.Chain(norm.NFD, runes.Remove(mn), norm.NFC),
	}
}

func (n *Normalizer) Norm(a string) (string, error) {
	n.src.Write([]byte(a))
	n.dst.Write([]byte(a)) // allocate at least this much
	dst, _, err := n.trans.Transform(n.dst.Bytes(), n.src.Bytes(), true)
	retVal := string(n.dst.Bytes()[:dst])
	n.src.Reset()
	n.dst.Reset()
	return retVal, err
}

func (n Normalizer) MustNorm(a string) string {
	normed, err := n.Norm(a)
	if err != nil {
		panic(err)
	}
	return normed
}
