package utils

import "strings"

// KMPMatch implements the Knuth-Morris-Pratt string matching algorithm (case-insensitive)
func KMPMatch(text string, pattern string) bool {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	n := len(text)
	m := len(pattern)
	if m == 0 {
		return true
	}

	lps := make([]int, m)
	lenLps := 0
	i := 1

	for i < m {
		if pattern[i] == pattern[lenLps] {
			lenLps++
			lps[i] = lenLps
			i++
		} else {
			if lenLps != 0 {
				lenLps = lps[lenLps-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}

	i = 0
	j := 0
	for i < n {
		if pattern[j] == text[i] {
			j++
			i++
		}
		if j == m {
			return true
		} else if i < n && pattern[j] != text[i] {
			if j != 0 {
				j = lps[j-1]
			} else {
				i++
			}
		}
	}
	return false
}
