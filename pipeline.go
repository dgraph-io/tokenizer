package tokenizer

import (
	"github.com/pkg/errors"
	"strings"
)

/*
normalizer : s -> (s|e)
splitter : s -> ([]s | e)
tokenizer : (s|[]s) -> ([]s | e)
*/

type Input interface {
	String() string
	Strings() []string
}

type Result interface {
	Input
	Err() error
}

type S string

func (s S) String() string    { return string(s) }
func (s S) Strings() []string { return []string{string(s)} }
func (s S) Err() error        { return nil }

type SS []string

func (s SS) String() string    { panic("Cannot return string") }
func (s SS) Strings() []string { return []string(s) }
func (s SS) Err() error {
	if len(s) == 0 {
		return errors.Errorf("Empty string slice")
	}
	return nil
}

type Err struct{ error }

func (e Err) String() string    { panic("Err is not a string") }
func (e Err) Strings() []string { panic("Err is not a string slice") }
func (e Err) Err() error        { return e.error }

type Step interface {
	Do(Input) Result
}

func Pipeline(steps ...Step) func(a string) ([]string, error) {
	return func(a string) ([]string, error) {
		var in Input = S(a)
		var out Result
		for i, step := range steps {
			out = step.Do(in)
			if err := out.Err(); err != nil {
				return nil, errors.Wrapf(err, "in Step %d of pipeline", i)
			}
			in = out
		}
		if err := out.Err(); err != nil {
			return nil, errors.Wrap(err, "In pipelined calls")
		}
		return out.Strings(), nil
	}
}

// some steps in the pipeline

func (n *Normalizer) Do(a Input) Result {
	s := a.String()
	retVal, err := n.Norm(s)
	if err != nil {
		return Err{err}
	}
	return S(retVal)
}

func (t *Tokenizer) Do(a Input) Result {
	ss := a.Strings()
	var retVal SS
	for _, s := range ss {
		toks, err := t.Tokenize(s)
		if err != nil {
			return Err{errors.Wrap(err, "in tokenizer.Do()")}
		}
		retVal = append(retVal, toks...)
	}
	return retVal
}

// Split is a function that splits a string into strings
type Split func(a string) []string

func (s Split) Do(a Input) Result {
	return SS(s(a.String()))
}

// BySpace splits a string into strings by space
func BySpace(a string) []string {
	return strings.Split(a, " ")
}

// Transform is a function that transforms a string.
type Transform func(a string) string

func (t Transform) Do(a Input) Result {
	return S(t(a.String()))
}
