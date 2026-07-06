package utils

import (
	"strings"
	"supplierhub-backend/models"
)

// KMPMatch mengimplementasikan Knuth-Morris-Pratt untuk pencarian string
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

// QuickSortPrice mengimplementasikan Quick Sort untuk mengurutkan harga
func QuickSortPrice(items []models.Product, low, high int, asc bool) {
	if low < high {
		pi := partitionPrice(items, low, high, asc)
		QuickSortPrice(items, low, pi-1, asc)
		QuickSortPrice(items, pi+1, high, asc)
	}
}

func partitionPrice(items []models.Product, low, high int, asc bool) int {
	pivot := items[high].Price
	i := low - 1
	for j := low; j < high; j++ {
		if asc {
			if items[j].Price <= pivot {
				i++
				items[i], items[j] = items[j], items[i]
			}
		} else {
			if items[j].Price >= pivot {
				i++
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	items[i+1], items[high] = items[high], items[i+1]
	return i + 1
}

// IsItemValid mengimplementasikan Binary Search untuk validasi ketersediaan produk (berdasarkan ID)
func IsItemValid(sortedIDs []string, targetID string) bool {
	low, high := 0, len(sortedIDs)-1
	for low <= high {
		mid := low + (high-low)/2
		if sortedIDs[mid] == targetID {
			return true // Item Valid
		} else if sortedIDs[mid] < targetID {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return false // Item Tidak Ditemukan / Invalid
}
