package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode"
)

func handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}

	profaneWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp := params.Body
	goodChirp := hideProfanity(chirp, profaneWords)


	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: goodChirp,
	})
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