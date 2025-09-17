package main

import (
	"strings"
	"unicode"
)

func cleanProfaneWords(input string) string {
	profaneWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}
	return hideProfanity(input, profaneWords)
}

func hideProfanity(input string, profaneWords []string) string {
	words := strings.Fields(input)
	for i, w := range words {
		base := strings.TrimRightFunc(w, func(r rune) bool {
			return unicode.IsPunct(r)
		})
		trailing := w[len(base):]

		for _, bad := range profaneWords {
			if strings.Contains(strings.ToLower(base), strings.ToLower(bad)) {
				words[i] = "****" + trailing
				break
			}
		}
	}
	return strings.Join(words, " ")
}