package models

import (
	"encoding/json"
	"time"
)

type NewMarket struct {
	ID                    string     `db:"id"`
	Symbol                string     `db:"symbol"`
	Interval              string     `db:"interval"`
	ConditionID           *string    `db:"condition_id"`
	TokenIDYes            *string    `db:"token_id_yes"`
	TokenIDNo             *string    `db:"token_id_no"`
	Question              *string    `db:"question"`
	Slug                  string     `db:"slug"`
	Outcomes              *string    `db:"outcomes"`
	StartDate             *time.Time `db:"start_date"`
	EndDate               *time.Time `db:"end_date"`
	GammaMarketID         *string    `db:"gamma_market_id"`
	EnableOrderBook       *bool      `db:"enable_order_book"`
	AcceptingOrders       *bool      `db:"accepting_orders"`
	Ready                 *bool      `db:"ready"`
	Funded                *bool      `db:"funded"`
	OrderMinSize          *float64   `db:"order_min_size"`
	OrderPriceMinTickSize *float64   `db:"order_price_min_tick_size"`
	BestBid               *float64   `db:"best_bid"`
	BestAsk               *float64   `db:"best_ask"`
	LastTradePrice        *float64   `db:"last_trade_price"`
	VolumeCLOB            *float64   `db:"volume_clob"`
	VolumeNum             *float64   `db:"volume_num"`
	Status                string     `db:"status"`
	WinningOutcome        *string    `db:"winning_outcome"`
	PriceToBeat           *float64   `db:"price_to_beat"`
	LastBookHash          *string    `db:"last_book_hash"`
	CreatedAt             time.Time  `db:"created_at"`
	UpdatedAt             time.Time  `db:"updated_at"`
}

// GammaMarketResponse is the JSON from GET /markets/slug/{slug}.
type GammaMarketResponse struct {
	ID                    string  `json:"id"`
	ConditionID           string  `json:"conditionId"`
	Slug                  string  `json:"slug"`
	Question              string  `json:"question"`
	Description           string  `json:"description"`
	Outcomes              string  `json:"outcomes"`
	Active                bool    `json:"active"`
	Closed                bool    `json:"closed"`
	ClosedTime            string  `json:"closedTime"`
	FiscalPeriod          string  `json:"fiscalPeriod"`
	AcceptingOrders       bool    `json:"acceptingOrders"`
	EnableOrderBook       bool    `json:"enableOrderBook"`
	NegRiskOther          bool    `json:"negRiskOther"`
	NegRisk               bool    `json:"negRisk"`
	NegRiskMarketID       string  `json:"negRiskMarketID"`
	Ready                 bool    `json:"ready"`
	Funded                bool    `json:"funded"`
	Announcing            bool    `json:"announcing"`
	Archived              bool    `json:"archived"`
	OvmL2L1StandardBridge *string `json:"ovm__L2L1StandardBridge"`
	OvmProposer           *string `json:"ovm__Proposer"`
	OvmProxyAdmin         *string `json:"ovm__ProxyAdmin"`
	OvmSequencer          *string `json:"ovm__Sequencer"`
	UsdcAddress           *string `json:"usdcAddress"`
	Deposited             bool    `json:"deposited"`
	Deploying             bool    `json:"deploying"`
	Disputing             bool    `json:"disputing"`
	Filling               bool    `json:"filling"`
	L1ResolutionTxHash    *string `json:"l1ResolutionTxHash"`
	L2TransactionHash     *string `json:"l2TransactionHash"`
	Live                  bool    `json:"live"`
	Locked                bool    `json:"locked"`
	LspAddress            *string `json:"lspAddress"`
	Pool                  *string `json:"pool"`
	Proposed              bool    `json:"proposed"`
	ReadyToPropose        bool    `json:"readyToPropose"`
	Resolved              bool    `json:"resolved"`
	StartDate             string  `json:"startDate"`
	EndDate               string  `json:"endDate"`
	Token0                string  `json:"token0"`
	Token1                string  `json:"token1"`
	UpdatedAt             string  `json:"updatedAt"`
	CreatedAt             string  `json:"createdAt"`
	GameStartTime         string  `json:"gameStartTime"`
	GameID                *string `json:"gameId"`
	GameSlug              *string `json:"gameSlug"`
	EventID               *string `json:"eventId"`
	EventSlug             *string `json:"eventSlug"`
	QuestionID            string  `json:"questionID"`
	RewardsMinSize        float64 `json:"rewardsMinSize"`
	RewardsMaxSpread      float64 `json:"rewardsMaxSpread"`
	Spread                float64 `json:"spread"`
	Volume24hr            float64 `json:"volume24hr"`
	Volume1wk             float64 `json:"volume1wk"`
	Volume                string  `json:"volume"`
	VolumeCLOB            float64 `json:"volumeClob"`
	VolumeNum             float64 `json:"volumeNum"`
	Liquidity             string  `json:"liquidity"`
	LiquidityNum          float64 `json:"liquidityNum"`
	LiquidityCLOB         float64 `json:"liquidityClob"`
	Fee                   *float64 `json:"fee"`
	MakerBaseFee          int     `json:"makerBaseFee"`
	TakerBaseFee          int     `json:"takerBaseFee"`
	MinIncentiveSize      float64 `json:"minIncentiveSize"`
	MaxIncentiveSpread    float64 `json:"maxIncentiveSpread"`
	BestBid               float64 `json:"bestBid"`
	BestAsk               float64 `json:"bestAsk"`
	LastTradePrice        float64 `json:"lastTradePrice"`
	OrderMinSize          float64 `json:"orderMinSize"`
	OrderPriceMinTickSize float64 `json:"orderPriceMinTickSize"`
	Price                 *float64 `json:"price"`
	Prices                *string `json:"prices"`
	Probability           *float64 `json:"probability"`
	Probabilities         *string `json:"probabilities"`
	OutcomesPrices        *string `json:"outcomePrices"`
	ClobTokenIDs          *string `json:"clobTokenIds"`
	LiquidityAMM          float64 `json:"liquidityAmm"`
	VolumeAMM             float64 `json:"volumeAmm"`
	MarketType            string  `json:"marketType"`
	MarketSlug            string  `json:"marketSlug"`
	IsMeta                *bool   `json:"isMeta"`
	Automated             bool    `json:"automated"`
	AutomaticallyActive   bool    `json:"automaticallyActive"`
	ManualActivation      bool    `json:"manualActivation"`
	OneDayPriceChange     float64 `json:"oneDayPriceChange"`
	OneHourPriceChange    float64 `json:"oneHourPriceChange"`
	OneWeekPriceChange    float64 `json:"oneWeekPriceChange"`
	OneMonthPriceChange   float64 `json:"oneMonthPriceChange"`
	OneYearPriceChange    float64 `json:"oneYearPriceChange"`
	Icon                  string  `json:"icon"`
	Image                 string  `json:"image"`
	OgImage               string  `json:"ogImage"`
	ClearingPrice         *float64 `json:"clearingPrice"`
	EnableNegRisk         bool    `json:"enableNegRisk"`
	GroupItemTitle        string  `json:"groupItemTitle"`
	CommentCount          int     `json:"commentCount"`
	Likes                 int     `json:"likes"`
	Dislikes              int     `json:"dislikes"`
	IsRecurring           bool    `json:"isRecurring"`
	RecurringDayOfWeek  int     `json:"recurringDayOfWeek"`
	UtcOffset             int     `json:"utcOffset"`
	IsYesNo               *bool   `json:"isYesNo"`
	Rewards               *string `json:"rewards"`
	Tags                  *string `json:"tags"`
	RelatedMarkets        *string `json:"relatedMarkets"`
	Events                json.RawMessage `json:"events,omitempty"`
	Category              *string `json:"category"`
	Categories            *string `json:"categories"`
}

// CLOBBookRequest is one entry for POST /books.
type CLOBBookRequest struct {
	TokenID string `json:"token_id"`
}

// CLOBBookResponse is one element from POST /books response.
type CLOBBookResponse struct {
	Market         string    `json:"market"`
	AssetID        string    `json:"asset_id"`
	Timestamp      string    `json:"timestamp"`
	Hash           string    `json:"hash"`
	Bids           []Level   `json:"bids"`
	Asks           []Level   `json:"asks"`
	MinOrderSize   string    `json:"min_order_size"`
	TickSize       string    `json:"tick_size"`
	NegRisk        bool      `json:"neg_risk"`
	LastTradePrice string    `json:"last_trade_price"`
}

type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// ActiveToken represents one token pair from a market.
type ActiveToken struct {
	TokenID  string `db:"token_id"`
	Side     string `db:"side"` // "yes" or "no"
	Symbol   string `db:"symbol"`
	Interval string `db:"interval"`
}

// BookSnapshot holds a full order book snapshot for one token.
type BookSnapshot struct {
	TokenID     string
	Side        string
	Symbol      string
	Interval    string
	LastTrade   *float64
	BookHash     *string
	TimestampAPI *int64     // raw timestamp returned by the CLOB API (epoch milliseconds)
	RawBids      []byte     // JSON array of bid levels
	RawAsks     []byte     // JSON array of ask levels
}
