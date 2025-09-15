package main

import (
	"log"
	"net/http"
)


func main() {
	const port = "8080"

	mux := http.NewServeMux()

	// Serve files under /app/
	appHandler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", appHandler)

	// Serve files from the "assets" directory under /assets/
	mux.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})

	// Health (readiness/liveness) endpoint
	mux.HandleFunc("/healthz", readinessHandler)


	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}