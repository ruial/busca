package core

import (
	"testing"
)

func TestTopTerms(t *testing.T) {
	query := []string{"some2", "some2", "test3", "test3", "test3", "query1"}
	expected := []FloatHeapItem{{"test3", 3}, {"some2", 2}, {"query1", 1}}
	tf := NewTermFrequency(query)
	top := tf.Top(20)
	for i := range top {
		if top[i] != expected[i] {
			t.Error("Invalid top terms:", top[i], expected[i])
		}
	}
	top = tf.Top(2)
	if len(top) != 2 {
		t.Error("Invalid top terms length:", top)
	}
}
