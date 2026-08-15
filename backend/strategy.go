package main

// This file re-implements, bar for bar, the logic from the user's Pine Script
// v6 indicator ("Deepanshu's Buy Sell Bot"): an ATR trailing-stop system
// (aka "UT Bot") with an optional Heikin Ashi source.

// Signal is the result of evaluating the strategy on the latest closed candle.
type Signal struct {
	Buy          bool
	Sell         bool
	Close        float64
	TrailingStop float64
}

// toHeikinAshi converts a normal OHLC series into Heikin Ashi candles,
// matching request.security(ticker.heikinashi(...), ..., close) behaviour.
func toHeikinAshi(candles []Candle) []float64 {
	haClose := make([]float64, len(candles))
	haOpen := make([]float64, len(candles))
	for i, c := range candles {
		haClose[i] = (c.Open + c.High + c.Low + c.Close) / 4
		if i == 0 {
			haOpen[i] = (c.Open + c.Close) / 2
		} else {
			haOpen[i] = (haOpen[i-1] + haClose[i-1]) / 2
		}
	}
	return haClose
}

// trueRange computes Wilder's true range for index i (i must be >= 1).
func trueRange(candles []Candle, i int) float64 {
	high, low, prevClose := candles[i].High, candles[i].Low, candles[i-1].Close
	tr := high - low
	if v := abs(high - prevClose); v > tr {
		tr = v
	}
	if v := abs(low - prevClose); v > tr {
		tr = v
	}
	return tr
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// atrRMA computes Wilder's RMA-smoothed ATR (what Pine's ta.atr() returns)
// for every bar, using the same recursive formula: atr[i] = (atr[i-1]*(n-1)+tr[i])/n.
func atrRMA(candles []Candle, period int) []float64 {
	n := len(candles)
	atr := make([]float64, n)
	if n == 0 {
		return atr
	}
	// Seed with a simple average of the first `period` true ranges.
	var sum float64
	seedEnd := period
	if seedEnd > n {
		seedEnd = n
	}
	for i := 1; i < seedEnd; i++ {
		sum += trueRange(candles, i)
	}
	if seedEnd > 1 {
		atr[seedEnd-1] = sum / float64(seedEnd-1)
	}
	for i := seedEnd; i < n; i++ {
		tr := trueRange(candles, i)
		atr[i] = (atr[i-1]*float64(period-1) + tr) / float64(period)
	}
	return atr
}

// Evaluate runs the full trailing-stop/crossover logic over candles and
// returns the signal for the LAST candle in the slice (i.e. the most
// recently closed one). Needs at least atrPeriod+2 candles to be meaningful.
func Evaluate(candles []Candle, keyValue float64, atrPeriod int, useHeikinAshi bool) Signal {
	n := len(candles)
	if n < atrPeriod+2 {
		return Signal{}
	}

	src := make([]float64, n)
	if useHeikinAshi {
		copy(src, toHeikinAshi(candles))
	} else {
		for i, c := range candles {
			src[i] = c.Close
		}
	}

	atr := atrRMA(candles, atrPeriod)

	stop := make([]float64, n)
	pos := make([]int, n)

	for i := 0; i < n; i++ {
		nLoss := keyValue * atr[i]
		if i == 0 {
			stop[i] = src[i] - nLoss
			continue
		}
		prevStop := stop[i-1]
		switch {
		case src[i] > prevStop && src[i-1] > prevStop:
			stop[i] = maxF(prevStop, src[i]-nLoss)
		case src[i] < prevStop && src[i-1] < prevStop:
			stop[i] = minF(prevStop, src[i]+nLoss)
		case src[i] > prevStop:
			stop[i] = src[i] - nLoss
		default:
			stop[i] = src[i] + nLoss
		}

		switch {
		case src[i-1] < stop[i-1] && src[i] > stop[i]:
			pos[i] = 1
		case src[i-1] > stop[i-1] && src[i] < stop[i]:
			pos[i] = -1
		default:
			pos[i] = pos[i-1]
		}
	}

	last := n - 1
	// ema1 with length 1 is just the source value itself, so the
	// crossover(ema1, stop) in the Pine script reduces to comparing src vs stop.
	above := src[last-1] <= stop[last-1] && src[last] > stop[last]
	below := src[last-1] >= stop[last-1] && src[last] < stop[last]

	buy := src[last] > stop[last] && above
	sell := src[last] < stop[last] && below

	return Signal{Buy: buy, Sell: sell, Close: candles[last].Close, TrailingStop: stop[last]}
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func minF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// SwingSupportResistance finds a simple support (recent swing low) and
// resistance (recent swing high) over the last `lookback` candles. Used as
// the automatic stop-loss level when the user leaves support_level /
// resistance_level at 0 in configuration.
func SwingSupportResistance(candles []Candle, lookback int) (support, resistance float64) {
	n := len(candles)
	if n == 0 {
		return 0, 0
	}
	start := n - lookback
	if start < 0 {
		start = 0
	}
	support = candles[start].Low
	resistance = candles[start].High
	for i := start; i < n; i++ {
		if candles[i].Low < support {
			support = candles[i].Low
		}
		if candles[i].High > resistance {
			resistance = candles[i].High
		}
	}
	return support, resistance
}
