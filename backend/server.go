package main

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	cfgStore *ConfigStore
	bot      *Bot
}

func NewServer(cfgStore *ConfigStore, bot *Bot) *Server {
	return &Server{cfgStore: cfgStore, bot: bot}
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-App-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	writeJSON(w, map[string]string{"error": err.Error()})
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", cors(s.handleStatus))
	mux.HandleFunc("/api/dashboard", cors(s.handleDashboard))
	mux.HandleFunc("/api/config", cors(s.handleConfig))
	mux.HandleFunc("/api/bot/start", cors(s.handleStart))
	mux.HandleFunc("/api/bot/stop", cors(s.handleStop))
	mux.HandleFunc("/api/logs", cors(s.handleLogs))
	return mux
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	pos, entry := s.bot.Status()
	writeJSON(w, map[string]interface{}{
		"running":     s.bot.IsRunning(),
		"position":    pos,
		"entry_price": entry,
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfgStore.Get()
	dc := NewDeltaClient(cfg.DeltaAPIKey, cfg.DeltaAPISecret, cfg.BaseURL)

	resp := map[string]interface{}{}

	if bal, err := dc.GetBalances(); err == nil {
		resp["balances"] = bal
	} else {
		resp["balances_error"] = err.Error()
	}

	if pos, err := dc.GetPositions(); err == nil {
		resp["positions"] = pos
	} else {
		resp["positions_error"] = err.Error()
	}

	if cfg.ProductSymbol != "" {
		if ticker, err := dc.GetTicker(cfg.ProductSymbol); err == nil {
			resp["market"] = ticker
		} else {
			resp["market_error"] = err.Error()
		}
	}

	position, entryPrice := s.bot.Status()
	resp["bot_position"] = position
	resp["bot_entry_price"] = entryPrice
	resp["bot_running"] = s.bot.IsRunning()

	writeJSON(w, resp)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := s.cfgStore.Get()
		cfg.DeltaAPISecret = "" // never echo the secret back to the app
		writeJSON(w, cfg)
	case http.MethodPost:
		var incoming Config
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		existing := s.cfgStore.Get()
		// Keep the stored secret if the app submitted a blank one (i.e. the
		// user only changed other fields on the Configuration screen).
		if incoming.DeltaAPISecret == "" {
			incoming.DeltaAPISecret = existing.DeltaAPISecret
		}
		incoming.BotRunning = existing.BotRunning // start/stop only via dedicated endpoints
		if err := s.cfgStore.Save(incoming); err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, map[string]string{"status": "saved"})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.bot.Start()
	writeJSON(w, map[string]string{"status": "started"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.bot.Stop()
	writeJSON(w, map[string]string{"status": "stopped"})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"logs": s.bot.Logs()})
}
