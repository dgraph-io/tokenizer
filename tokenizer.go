package tokenizer

import (
	"github.com/go-nlp/bpe"
	"github.com/pkg/errors"
)

type Tokenizer struct {
	ranks   map[bpe.Pair]int
	enc     bpe.Encoder
	lut     map[rune]bpe.Pair
	pairbuf []bpe.Pair
}

func NewTokenizer(enc bpe.Encoder) *Tokenizer {
	ranks := make(map[bpe.Pair]int)
	for i, p := range enc.Pairs {
		ranks[p] = i
	}
	lut := make(map[rune]bpe.Pair)
	for k, v := range enc.Replacements {
		lut[v] = k
	}
	return &Tokenizer{
		ranks:   ranks,
		enc:     enc,
		lut:     lut,
		pairbuf: make([]bpe.Pair, 0, 256),
	}
}

type tmp struct {
	P bpe.Pair
}

func (t *Tokenizer) Tokenize(a string) ([]string, error) {
	pairs := bpe.Pairs(a, bpe.WithReuse(t.pairbuf))
	if len(pairs) == 0 {
		return []string{a}, nil
	}

	// var b bytes.Buffer
	// for i := range t.enc.Pairs {
	// 	fmt.Fprintf(&b, "%v:%d\n", t.enc.Pairs[i], i)
	// }
	// fmt.Fprintf(&b, "\n\nDICT\n")
	// for i := 0; i < t.enc.Corpus.Size(); i++ {
	// 	word, _ := t.enc.Corpus.Word(i)
	// 	freq := t.enc.Corpus.IDFreq(i)
	// 	word = t.Untokenize([]string{word})
	// 	fmt.Fprintf(&b, "%d - %q:%d\n", i, word, freq)

	// }
	// ioutil.WriteFile("tmp", b.Bytes(), 0644)

	w := []rune(a)
	newWord := make([]rune, 0, len(w)) // newWord is a buffer for working. It would at most be the same length as `w`
	for {
		bigram, ok := t.minRank(pairs)
		if !ok {
			break
		}
		fst := bigram.Fst()
		snd := bigram.Snd()

		for i := 0; i < len(w); {
			j, err := index(w, i, fst)
			if err != nil {
				newWord = append(newWord, w[i:]...)
				break
			} else {
				newWord = append(newWord, w[i:j]...)
				i = j
			}

			if w[i] == fst && i < len(w)-1 && w[i+1] == snd {
				replacement, ok := t.enc.Replacements[bigram]
				if !ok {
					return nil, errors.Errorf("Cannot find replacement for the bigram %v", bigram)
				}
				newWord = append(newWord, replacement)
				i += 2
			} else {
				newWord = append(newWord, w[i])
				i++
			}
		}
		copy(w, newWord)
		w = w[:len(newWord)]
		newWord = newWord[:0]
		if len(w) == 1 {
			break
		}
		pairs = bpe.PairsRunes(w, bpe.WithReuse(t.pairbuf))
	}

	buf := make([]rune, 0, 8) // 8 words is more than enough for a token
	var tokens []string
	for _, r := range w {
		buf = t.untokenize(r, buf)
		if len(buf) == 0 {
			continue
		}
		s := string(buf)
		if s == " " {
			continue
		}
		tokens = append(tokens, s)
		buf = buf[:0]
	}
	return tokens, nil
}

func (t *Tokenizer) Untokenize(a []string) string {
	var retVal []rune
	for _, w := range a {
		asRunes := []rune(w)
		var buf []rune
		for _, r := range asRunes {
			buf = t.untokenize(r, buf)
			retVal = append(retVal, buf...)
			buf = buf[:0]
		}
	}
	return string(retVal)
}

func (t *Tokenizer) untokenize(a rune, buf []rune) []rune {
	pair, ok := t.lut[a]
	if ok {
		buf = t.untokenize(pair.Fst(), buf)
		buf = t.untokenize(pair.Snd(), buf)
	} else {
		buf = append(buf, a)
	}
	return buf
}

func (t *Tokenizer) minRank(ps []bpe.Pair) (min bpe.Pair, ok bool) {
	rank := len(t.ranks) + 1
	for _, p := range ps {
		r, k := t.ranks[p]
		if k && r < rank {
			rank = r
			min = p
			ok = true
		}
	}
	return
}

// UTIL

func index(rs []rune, start int, of rune) (int, error) {
	if start >= len(rs) {
		return -1, errors.Errorf("start:%d < len(rs):%d", start, len(rs))
	}
	for i, r := range rs[start:] {
		if r == of {
			return i + start, nil
		}
	}
	return -1, errors.Errorf("Not Found")
}
