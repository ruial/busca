package core

import (
	"testing"

	"github.com/ruial/busca/test"
)

func TestTopTerms(t *testing.T) {
	query := []string{"some2", "some2", "test3", "test3", "test3", "query1"}
	expected := []string{"test3", "some2", "query1"}
	tf := NewTermFrequency(query)
	top := tf.Top(20)
	if !test.StringArrayEquals(top, expected, true) {
		t.Error("Invalid top terms:", top, expected)
	}
	top = tf.Top(2)
	if len(top) != 2 {
		t.Error("Invalid top terms length:", top)
	}
}
