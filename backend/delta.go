package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DeltaClient talks to Delta Exchange's v2 REST API.
// Signing scheme (per docs.delta.exchange): signature = HMAC_SHA256(secret,
// method + timestamp + requestPath + queryString + body)
type DeltaClient struct {
	APIKey    string
	APISecret string
	BaseURL   string
	HTTP      *http.Client
}

func NewDeltaClient(apiKey, apiSecret, baseURL string) *DeltaClient {
	return &DeltaClient{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   baseURL,
		HTTP:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (d *DeltaClient) sign(method, path, query, body string) (sig, ts string) {
	ts = strconv.FormatInt(time.Now().Unix(), 10)
	payload := method + ts + path + query + body
	mac := hmac.New(sha256.New, []byte(d.APISecret))
	mac.Write([]byte(payload))
	sig = hex.EncodeToString(mac.Sum(nil))
	return sig, ts
}

// request performs a signed call. queryVals may be nil. body (if non-nil) is
// marshalled to JSON. Set signed=false for public endpoints (candles).
func (d *DeltaClient) request(method, path string, queryVals url.Values, body interface{}, signed bool) ([]byte, error) {
	q := ""
	if queryVals != nil && len(queryVals) > 0 {
		q = "?" + queryVals.Encode()
	}
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, d.BaseURL+path+q, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "dpsharma-tradingbot/1.0")

	if signed {
		sig, ts := d.sign(method, path, q, string(bodyBytes))
		req.Header.Set("api-key", d.APIKey)
		req.Header.Set("timestamp", ts)
		req.Header.Set("signature", sig)
	}

	resp, err := d.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return out, fmt.Errorf("delta api %s %s -> %d: %s", method, path, resp.StatusCode, string(out))
	}
	return out, nil
}

// ---------- Market data ----------

type Candle struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

type candleResp struct {
	Success bool `json:"success"`
	Result  []struct {
		Time   int64   `json:"time"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume float64 `json:"volume"`
	} `json:"result"`
}

// GetCandles fetches the last `count` closed candles for symbol/resolution.
func (d *DeltaClient) GetCandles(symbol, resolution string, count int) ([]Candle, error) {
	end := time.Now().Unix()
	start := end - int64(count)*resolutionSeconds(resolution)
	q := url.Values{}
	q.Set("resolution", resolution)
	q.Set("symbol", symbol)
	q.Set("start", strconv.FormatInt(start, 10))
	q.Set("end", strconv.FormatInt(end, 10))

	b, err := d.request(http.MethodGet, "/v2/history/candles", q, nil, false)
	if err != nil {
		return nil, err
	}
	var cr candleResp
	if err := json.Unmarshal(b, &cr); err != nil {
		return nil, err
	}
	candles := make([]Candle, 0, len(cr.Result))
	for _, r := range cr.Result {
		candles = append(candles, Candle{Time: r.Time, Open: r.Open, High: r.High, Low: r.Low, Close: r.Close, Volume: r.Volume})
	}
	// Delta returns newest-first; normalise to oldest-first for the strategy code.
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}
	return candles, nil
}

func (d *DeltaClient) GetTicker(symbol string) (map[string]interface{}, error) {
	b, err := d.request(http.MethodGet, "/v2/tickers/"+symbol, nil, nil, false)
	if err != nil {
		return nil, err
	}
	var out struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

// ---------- Account ----------

func (d *DeltaClient) GetBalances() (json.RawMessage, error) {
	b, err := d.request(http.MethodGet, "/v2/wallet/balances", nil, nil, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func (d *DeltaClient) GetPositions() (json.RawMessage, error) {
	b, err := d.request(http.MethodGet, "/v2/positions/margined", nil, nil, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func (d *DeltaClient) GetOrders() (json.RawMessage, error) {
	q := url.Values{}
	q.Set("states", "open,pending")
	b, err := d.request(http.MethodGet, "/v2/orders", q, nil, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ---------- Orders ----------

// BracketOrderReq places a market entry order with a take-profit and
// stop-loss bracket attached, per docs.delta.exchange (place order,
// bracket_* fields).
type BracketOrderReq struct {
	ProductID              int    `json:"product_id"`
	ProductSymbol          string `json:"product_symbol"`
	Size                   int    `json:"size"`
	Side                   string `json:"side"` // "buy" or "sell"
	OrderType              string `json:"order_type"`
	ReduceOnly             bool   `json:"reduce_only"`
	ClientOrderID          string `json:"client_order_id,omitempty"`
	BracketStopLossPrice   string `json:"bracket_stop_loss_price,omitempty"`
	BracketTakeProfitPrice string `json:"bracket_take_profit_price,omitempty"`
	BracketStopTrigger     string `json:"bracket_stop_trigger_method,omitempty"`
}

func (d *DeltaClient) PlaceBracketOrder(productID int, symbol, side string, size int, stopLoss, takeProfit float64, clientOrderID string) (json.RawMessage, error) {
	req := BracketOrderReq{
		ProductID:              productID,
		ProductSymbol:          symbol,
		Size:                   size,
		Side:                   side,
		OrderType:              "market_order",
		ReduceOnly:             false,
		ClientOrderID:          clientOrderID,
		BracketStopLossPrice:   fmt.Sprintf("%.2f", stopLoss),
		BracketTakeProfitPrice: fmt.Sprintf("%.2f", takeProfit),
		BracketStopTrigger:     "last_traded_price",
	}
	b, err := d.request(http.MethodPost, "/v2/orders", nil, req, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

// ClosePosition sends a reduce-only market order in the opposite direction
// to flatten the current position.
func (d *DeltaClient) ClosePosition(productID int, symbol, closingSide string, size int) (json.RawMessage, error) {
	req := map[string]interface{}{
		"product_id":     productID,
		"product_symbol": symbol,
		"size":           size,
		"side":           closingSide,
		"order_type":     "market_order",
		"reduce_only":    true,
	}
	b, err := d.request(http.MethodPost, "/v2/orders", nil, req, true)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}

func resolutionSeconds(res string) int64 {
	switch res {
	case "1m":
		return 60
	case "3m":
		return 180
	case "5m":
		return 300
	case "15m":
		return 900
	case "30m":
		return 1800
	case "1h":
		return 3600
	case "4h":
		return 14400
	case "1d":
		return 86400
	default:
		return 300
	}
}
