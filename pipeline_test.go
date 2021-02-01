package tokenizer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipeline(t *testing.T) {
	n := NewNormalizer()
	tok, err := testingTokenizer()
	require.NoError(t, err)
	f := Pipeline(n, Transform(strings.ToLower), Split(BySpace), tok)

	tokens, err := f("你好吗？ 我现在在Wonderland to meet Alice ")
	require.NoError(t, err)

	t.Logf("Tokens %q", tokens)
}
