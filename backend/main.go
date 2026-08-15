package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	configPath := envOr("CONFIG_PATH", "config.json")
	port := envOr("PORT", "8080")
	appToken := os.Getenv("APP_TOKEN") // set this and put the same value in the Android app settings

	cfgStore := NewConfigStore(configPath)
	bot := NewBot(cfgStore)

	// Resume the bot automatically if it was left running before a restart.
	if cfgStore.Get().BotRunning {
		cfgStore.SetRunning(false) // reset so Start() doesn't no-op
		bot.Start()
	}

	srv := NewServer(cfgStore, bot)
	mux := srv.Routes()

	handler := authMiddleware(appToken, mux)

	log.Printf("D.P.Sharma trading bot backend listening on :%s (config: %s)", port, configPath)
	if appToken == "" {
		log.Println("WARNING: APP_TOKEN is not set — the API is unauthenticated. Set APP_TOKEN before exposing this on a public VPS.")
	}
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// authMiddleware requires header "X-App-Token: <token>" to match APP_TOKEN
// when APP_TOKEN is set. This is a lightweight guard so the bot's
// start/stop/config endpoints aren't wide open on a public VPS IP.
func authMiddleware(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("X-App-Token") != token {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
