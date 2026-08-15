package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Bot owns the running state: current position, last processed candle, and
// the goroutine loop that polls Delta Exchange and reacts to signals.
type Bot struct {
	cfgStore *ConfigStore

	mu             sync.Mutex
	position       string // "flat" | "long" | "short"
	entrySize      int
	entryPrice     float64
	lastCandleTime int64
	logs           []string

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewBot(cfgStore *ConfigStore) *Bot {
	return &Bot{cfgStore: cfgStore, position: "flat"}
}

func (b *Bot) client() *DeltaClient {
	c := b.cfgStore.Get()
	return NewDeltaClient(c.DeltaAPIKey, c.DeltaAPISecret, c.BaseURL)
}

func (b *Bot) log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg)
	log.Println(line)
	b.mu.Lock()
	b.logs = append(b.logs, line)
	if len(b.logs) > 300 {
		b.logs = b.logs[len(b.logs)-300:]
	}
	b.mu.Unlock()
}

func (b *Bot) Logs() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.logs))
	copy(out, b.logs)
	return out
}

func (b *Bot) Status() (position string, entryPrice float64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.position, b.entryPrice
}

func (b *Bot) IsRunning() bool {
	return b.cfgStore.Get().BotRunning
}

func (b *Bot) Start() {
	if b.IsRunning() {
		return
	}
	b.cfgStore.SetRunning(true)
	b.stopCh = make(chan struct{})
	b.wg.Add(1)
	go b.loop(b.stopCh)
	b.log("Bot started")
}

func (b *Bot) Stop() {
	if !b.IsRunning() {
		return
	}
	b.cfgStore.SetRunning(false)
	close(b.stopCh)
	b.wg.Wait()
	b.log("Bot stopped")
}

func (b *Bot) loop(stop chan struct{}) {
	defer b.wg.Done()
	for {
		select {
		case <-stop:
			return
		default:
		}

		cfg := b.cfgStore.Get()
		if err := b.tick(cfg); err != nil {
			b.log("tick error: %v", err)
		}

		interval := cfg.PollIntervalSeconds
		if interval <= 0 {
			interval = 30
		}
		select {
		case <-stop:
			return
		case <-time.After(time.Duration(interval) * time.Second):
		}
	}
}

// tick fetches candles, evaluates the strategy on the latest CLOSED candle,
// and acts on a fresh buy/sell signal.
func (b *Bot) tick(cfg Config) error {
	dc := NewDeltaClient(cfg.DeltaAPIKey, cfg.DeltaAPISecret, cfg.BaseURL)

	candles, err := dc.GetCandles(cfg.ProductSymbol, cfg.Resolution, cfg.ATRPeriod*5+50)
	if err != nil {
		return fmt.Errorf("fetch candles: %w", err)
	}
	if len(candles) < cfg.ATRPeriod+2 {
		return fmt.Errorf("not enough candles yet (%d)", len(candles))
	}

	// Drop the still-forming candle so we only ever act on a CLOSED bar.
	closed := candles[:len(candles)-1]
	last := closed[len(closed)-1]

	b.mu.Lock()
	alreadyProcessed := last.Time == b.lastCandleTime
	b.mu.Unlock()
	if alreadyProcessed {
		return nil
	}

	sig := Evaluate(closed, cfg.KeyValue, cfg.ATRPeriod, cfg.UseHeikinAshi)

	b.mu.Lock()
	b.lastCandleTime = last.Time
	b.mu.Unlock()

	if !sig.Buy && !sig.Sell {
		return nil
	}

	expenses := cfg.FeeAmount + cfg.FundingAmount + cfg.BrokerageAmount
	profitOffset := expenses * cfg.Beta
	support, resistance := cfg.SupportLevel, cfg.ResistanceLevel
	if support == 0 || resistance == 0 {
		autoSupport, autoResistance := SwingSupportResistance(closed, cfg.SRLookback)
		if support == 0 {
			support = autoSupport
		}
		if resistance == 0 {
			resistance = autoResistance
		}
	}
	size := cfg.LotSize * cfg.NumLots
	if size <= 0 {
		size = 1
	}

	b.mu.Lock()
	pos := b.position
	b.mu.Unlock()

	if sig.Buy && pos != "long" {
		if pos == "short" {
			b.log("BUY signal — closing existing SHORT first")
			if _, err := dc.ClosePosition(cfg.ProductID, cfg.ProductSymbol, "buy", size); err != nil {
				return fmt.Errorf("close short: %w", err)
			}
		}
		tp := last.Close + profitOffset
		sl := support
		if sl <= 0 || sl >= last.Close {
			sl = last.Close - 3*expenses - 1 // safety fallback so SL is never above entry
		}
		b.log("BUY signal @ %.2f -> entering LONG size=%d TP=%.2f SL(support)=%.2f", last.Close, size, tp, sl)
		if _, err := dc.PlaceBracketOrder(cfg.ProductID, cfg.ProductSymbol, "buy", size, sl, tp, fmt.Sprintf("long_%d", last.Time)); err != nil {
			return fmt.Errorf("place long bracket: %w", err)
		}
		b.mu.Lock()
		b.position, b.entryPrice, b.entrySize = "long", last.Close, size
		b.mu.Unlock()
	}

	if sig.Sell && pos != "short" {
		if pos == "long" {
			b.log("SELL signal — closing existing LONG first")
			if _, err := dc.ClosePosition(cfg.ProductID, cfg.ProductSymbol, "sell", size); err != nil {
				return fmt.Errorf("close long: %w", err)
			}
		}
		tp := last.Close - profitOffset
		sl := resistance
		if sl <= 0 || sl <= last.Close {
			sl = last.Close + 3*expenses + 1 // safety fallback so SL is never below entry
		}
		b.log("SELL signal @ %.2f -> entering SHORT size=%d TP=%.2f SL(resistance)=%.2f", last.Close, size, tp, sl)
		if _, err := dc.PlaceBracketOrder(cfg.ProductID, cfg.ProductSymbol, "sell", size, sl, tp, fmt.Sprintf("short_%d", last.Time)); err != nil {
			return fmt.Errorf("place short bracket: %w", err)
		}
		b.mu.Lock()
		b.position, b.entryPrice, b.entrySize = "short", last.Close, size
		b.mu.Unlock()
	}

	return nil
}
