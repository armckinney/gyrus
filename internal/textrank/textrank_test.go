package textrank_test

import (
	"reflect"
	"testing"

	"github.com/armckinney/gyrus/internal/textrank"
)

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies tokenization: lowercasing, stopwords dropped, no stemming, and hyphenated words.
// [Execution Surface]: Pure function
// [Assertions]: "How do I send Tasks on the agent-ops bus?" yields send, tasks, agent-ops, agent, ops, bus.
// -----------------------------------------------------------------------------
func TestTokenize(t *testing.T) {
	got := textrank.Tokenize("How do I send Tasks on the agent-ops bus?")
	want := []string{"send", "tasks", "agent-ops", "agent", "ops", "bus"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tokenize = %v, want %v", got, want)
	}
	if got := textrank.Unique([]string{"a", "b", "a"}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("Unique = %v", got)
	}
}

// -----------------------------------------------------------------------------
// [Test Level]: Unit Test
// [Purpose]: Verifies BM25: rare terms outweigh common ones, and a matching short document beats a padded one.
// [Execution Surface]: Pure function
// [Assertions]: The document holding both query terms ranks first; one holding only a term common to every
// document scores below one holding the rare term; a document with no term scores zero.
// -----------------------------------------------------------------------------
func TestCorpusScore(t *testing.T) {
	docs := [][]string{
		textrank.Tokenize("send a task to a session and close the thread"),
		textrank.Tokenize("session status session session"),
		textrank.Tokenize("thread etiquette for a session"),
		textrank.Tokenize("deploy the game server"),
	}
	c := textrank.NewCorpus(docs)
	q := textrank.Unique(textrank.Tokenize("close a thread"))
	if !(c.Score(0, q) > c.Score(2, q) && c.Score(2, q) > c.Score(1, q)) {
		t.Errorf("unexpected order: %v %v %v", c.Score(0, q), c.Score(2, q), c.Score(1, q))
	}
	if c.Score(3, q) != 0 {
		t.Errorf("non-matching document must score 0, got %v", c.Score(3, q))
	}
	common := textrank.Unique(textrank.Tokenize("session"))
	rare := textrank.Unique(textrank.Tokenize("deploy"))
	if c.Score(1, common) >= c.Score(3, rare) {
		t.Errorf("a term in most documents must weigh less than a rare one: %v >= %v", c.Score(1, common), c.Score(3, rare))
	}
}
