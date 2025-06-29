package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

func Top10(text string) []string {
	if text == "" {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}
	wordCount := make(map[string]int)
	for _, word := range words {
		wordCount[word]++
	}

	type wordFreq struct {
		word  string
		count int
	}
	var freqs []wordFreq
	for word, count := range wordCount {
		freqs = append(freqs, wordFreq{word, count})
	}

	sort.Slice(freqs, func(i, j int) bool {
		if freqs[i].count == freqs[j].count {
			return freqs[i].word < freqs[j].word
		}
		return freqs[i].count > freqs[j].count
	})

	result := make([]string, 0, 10)
	for i := 0; i < len(freqs) && i < 10; i++ {
		result = append(result, freqs[i].word)
	}
	return result
}
