package textrank_test

import (
	"testing"

	"github.com/armckinney/gyrus/internal/textrank"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies BM25 and tokenization at their edges.
// [Execution Surface]: Pure function
// [Assertions]: an empty corpus, empty documents (average length 0), empty query terms and text of only stopwords or
// punctuation all score 0 or tokenize to nothing, without panicking.
// -----------------------------------------------------------------------------
func TestEdges(t *testing.T) {
	if got := textrank.NewCorpus(nil); got == nil {
		t.Fatal("NewCorpus(nil) must return a corpus")
	}
	empty := textrank.NewCorpus([][]string{{}, {}})
	if got := empty.Score(0, []string{"queue"}); got != 0 {
		t.Errorf("all-empty corpus: score %v", got)
	}
	c := textrank.NewCorpus([][]string{{"queue", "restart"}, {"deploy"}})
	if got := c.Score(0, nil); got != 0 {
		t.Errorf("empty terms: score %v", got)
	}
	for _, text := range []string{"", "the and of", "!!! --- ___", "   "} {
		if got := textrank.Tokenize(text); len(got) != 0 {
			t.Errorf("Tokenize(%q) = %v, want nothing", text, got)
		}
	}
}
