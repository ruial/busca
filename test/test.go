// Package test contains test utils
package test

import "sort"

func StringArrayEquals(a, b []string, keepOrder bool) bool {
	if len(a) != len(b) {
		return false
	}
	if !keepOrder {
		sort.Strings(a)
		sort.Strings(b)
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
