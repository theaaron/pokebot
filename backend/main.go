// backend/main.go
package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("✗  GEMINI_API_KEY is not set")
		fmt.Println("   export GEMINI_API_KEY=\"your-key\"")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := NewApp(apiKey)

	mux := http.NewServeMux()
	mux.HandleFunc("GET    /api/conversations", app.listConversations)
	mux.HandleFunc("POST   /api/conversations", app.createConversation)
	mux.HandleFunc("GET    /api/conversations/{id}", app.getConversation)
	mux.HandleFunc("DELETE /api/conversations/{id}", app.deleteConversation)
	mux.HandleFunc("POST   /api/conversations/{id}/chat", app.chatStream)
	mux.HandleFunc("GET    /health", app.health)

	fmt.Printf("PokéBot API → http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, corsMiddleware(mux)); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

// ── CORS ──────────────────────────────────────────────────────────────

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
