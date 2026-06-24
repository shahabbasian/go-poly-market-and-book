package models

// Coin maps symbols to their metadata.
type Coin struct {
	Symbol   string
	FullName string
}

var Coins = []Coin{
	{Symbol: "btc", FullName: "bitcoin"},
	{Symbol: "eth", FullName: "ethereum"},
	{Symbol: "sol", FullName: "solana"},
	{Symbol: "xrp", FullName: "xrp"},
}

// Interval represents one of the active market timeframes.
type Interval struct {
	Name     string
	Duration int64 // seconds between markets
	SlugKey  string
}

var Intervals = []Interval{
	{Name: "5m", Duration: 300, SlugKey: "5m"},
	{Name: "15m", Duration: 900, SlugKey: "15m"},
}

var CoinSymbolMap = make(map[string]string)
var CoinFullNameMap = make(map[string]string)

func init() {
	for _, c := range Coins {
		CoinSymbolMap[c.Symbol] = c.FullName
		CoinFullNameMap[c.FullName] = c.Symbol
	}
}
