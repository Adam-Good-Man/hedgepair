package util

type Precision struct {
	Symbol             string `json:"symbol"`
	PricePrecision     int    `json:"pricePrecision"`
	QuantityPrecision  int    `json:"quantityPrecision"`
	BaseAssetPrecision int    `json:"baseAssetPrecision"`
	QuotePrecision     int    `json:"quotePrecision"`
}

// TickerResponse 定义了 API 响应的结构
type TickerResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type ExchangeInfo struct {
	Symbols []Precision `json:"symbols"`
}

type LeverageInfo struct {
	// 能够设置的 最高 保证金
	Leverage int `json:"initialLeverage"`
	// 维持保证金率
	MaintMarginRatio float64 `json:"maintMarginRatio"`
	// ...其他字段...
}

// LeverageBracket 结构体用于解析API响应
type LeverageBracket struct {
	Symbol   string         `json:"symbol"`
	Brackets []LeverageInfo `json:"brackets"`
}

// trade 交易列表
type TradeInfo struct {
	Symbol string  `json:"symbol"`
	Doller float64 `json:"doller"`
	Times  int     `json:"times"`
}

type TradeConfig struct {
	Symbol             string
	Times              int
	Decimal            int
	Doller             float64
	MarginType         string
	PriceOrder         float64
	LastPrice          float64
	LastProcessedPrice float64 // 上次处理的价格
	Amount             float64
	CountUp            int
	CountDown          int
	TradeUp            bool
	TradeDown          bool
	Liquidation        float64
	TargetProfit       float64
}

// Position 仓位信息结构体
type Position struct {
	Symbol           string  `json:"symbol"`
	PositionAmt      float64 `json:"positionAmt,string"`
	EntryPrice       float64 `json:"entryPrice,string"`
	MarkPrice        float64 `json:"markPrice,string"`
	UnRealizedPnl    float64 `json:"unRealizedPnl,string"`
	LiquidationPrice float64 `json:"liquidationPrice,string"`
	Leverage         int     `json:"leverage,string"`
	MaxNotionalValue float64 `json:"maxNotionalValue,string"`
	MarginType       string  `json:"marginType"`
	IsolatedMargin   float64 `json:"isolatedMargin,string"`
	IsAutoAddMargin  bool    `json:"isAutoAddMargin,string"`
	PositionSide     string  `json:"positionSide"`
	Notional         float64 `json:"notional,string"`
	IsolatedWallet   float64 `json:"isolatedWallet,string"`
	UpdateTime       int64   `json:"updateTime"`
}

// PositionResponse API响应结构体
type PositionResponse struct {
	Positions []Position `json:"positions"`
}

// 获取 委托列表
type OpenOrder struct {
	Symbol       string  `json:"symbol"`
	OrderId      int64   `json:"orderId"`
	Price        float64 `json:"price,string"`
	OrigQty      float64 `json:"origQty,string"`
	ExecutedQty  float64 `json:"executedQty,string"`
	Status       string  `json:"status"`
	Side         string  `json:"side"`
	PositionSide string  `json:"positionSide"`
	Type         string  `json:"type"`
}
