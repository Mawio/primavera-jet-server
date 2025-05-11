package main

import "net/http"

// I don't understsand if this is the proper way of doing this
func corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                            // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")               // Allow POST and OPTIONS methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allow specific headers
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			h.ServeHTTP(w, r)
		}
	}
}
