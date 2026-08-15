package main

import (
	"encoding/json"
	"os"
	"sync"
)

// Config holds every setting the Android app exposes on the "Configuration" tab.
// It is persisted to config.json next to the binary so restarts don't lose it.
type Config struct {
	// Delta Exchange credentials — NEVER sent to the Android app, only used server-side.
	DeltaAPIKey    string `json:"delta_api_key"`
	DeltaAPISecret string `json:"delta_api_secret"`
	BaseURL        string `json:"base_url"` // https://api.india.delta.exchange (or https://cdn-ind.testnet.deltaex.org for testnet)

	// Market / instrument
	ProductID     int    `json:"product_id"`     // Delta numeric product id for the contract you trade
	ProductSymbol string `json:"product_symbol"` // e.g. "BTCUSD"
	Resolution    string `json:"resolution"`     // candle timeframe: "1m","5m","15m","1h" ...

	// Strategy (mirrors the Pine script inputs)
	KeyValue      float64 `json:"key_value"`       // "a" - ATR multiplier / sensitivity
	ATRPeriod     int     `json:"atr_period"`      // "c"
	UseHeikinAshi bool    `json:"use_heikin_ashi"` // "h"

	// Position sizing
	LotSize int `json:"lot_size"` // contracts per lot
	NumLots int `json:"num_lots"` // number of lots per trade

	// Costs — same currency as the instrument is quoted in, per trade
	FeeAmount       float64 `json:"fee_amount"`
	FundingAmount   float64 `json:"funding_amount"`
	BrokerageAmount float64 `json:"brokerage_amount"`
	Beta            float64 `json:"beta"` // how many multiples of total expenses to target as profit

	// Stop loss — support/resistance. Leave at 0 to let the bot auto-detect
	// swing high/low over SRLookback candles instead of a fixed manual level.
	SupportLevel    float64 `json:"support_level"`
	ResistanceLevel float64 `json:"resistance_level"`
	SRLookback      int     `json:"sr_lookback"`

	// Runtime
	BotRunning          bool `json:"bot_running"`
	PollIntervalSeconds int  `json:"poll_interval_seconds"`
}

func defaultConfig() Config {
	return Config{
		BaseURL:             "https://api.india.delta.exchange",
		ProductSymbol:       "BTCUSD",
		Resolution:          "5m",
		KeyValue:            1,
		ATRPeriod:           10,
		UseHeikinAshi:       false,
		LotSize:             1,
		NumLots:             1,
		FeeAmount:           0,
		FundingAmount:       0,
		BrokerageAmount:     0,
		Beta:                2,
		SupportLevel:        0,
		ResistanceLevel:     0,
		SRLookback:          50,
		BotRunning:          false,
		PollIntervalSeconds: 30,
	}
}

// ConfigStore is a mutex-guarded, disk-backed holder for Config so the HTTP
// handlers and the bot loop can safely read/write it concurrently.
type ConfigStore struct {
	mu   sync.RWMutex
	path string
	cfg  Config
}

func NewConfigStore(path string) *ConfigStore {
	cs := &ConfigStore{path: path, cfg: defaultConfig()}
	cs.load()
	return cs
}

func (cs *ConfigStore) load() {
	b, err := os.ReadFile(cs.path)
	if err != nil {
		return // no file yet -> defaults stay
	}
	var c Config
	if err := json.Unmarshal(b, &c); err == nil {
		cs.cfg = c
	}
}

func (cs *ConfigStore) Get() Config {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.cfg
}

func (cs *ConfigStore) Save(c Config) error {
	cs.mu.Lock()
	cs.cfg = c
	cs.mu.Unlock()
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, b, 0600)
}

func (cs *ConfigStore) SetRunning(running bool) {
	cs.mu.Lock()
	cs.cfg.BotRunning = running
	cs.mu.Unlock()
}
