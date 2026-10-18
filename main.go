package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"example.com/m/cache"
	"example.com/m/util"
)

func initializeTrading(config *util.TradeConfig) bool {
	ok := util.SetMarginType(config.Symbol, config.MarginType)
	util.SetLeverage(config.Symbol, config.Times)
	return ok
}

// shouldProcessPrice 判断是否应该处理这个价格变化
func shouldProcessPrice(config *util.TradeConfig, currentPrice float64) bool {
	// 如果是第一次处理价格，直接处理
	if config.LastProcessedPrice == 0 {
		config.LastProcessedPrice = currentPrice
		fmt.Printf("[%s] 首次价格处理: %.4f\n", config.Symbol, currentPrice)
		return true
	}

	// 计算价格变化的精度
	priceDiff := math.Abs(currentPrice - config.LastProcessedPrice)

	// 根据Decimal计算最小变化单位
	minChange := math.Pow(10, -float64(config.Decimal))

	// 只有当价格变化达到或超过最小变化单位时才处理
	if priceDiff >= minChange {
		config.LastProcessedPrice = currentPrice
		fmt.Printf("[%s] 价格变化达到阈值: %.4f -> %.4f (变化: %.6f, 阈值: %.6f)\n",
			config.Symbol, config.LastProcessedPrice, currentPrice, priceDiff, minChange)
		return true
	}

	return false
}

func subscribeAndHandleHedgePrice(config *util.TradeConfig) {
	fmt.Println("subscribeAndHandleHedgePrice", config.Symbol, config.Amount)
	priceChan := make(chan float64)
	go util.StreamPrice(config.Symbol, priceChan)

	for price := range priceChan {
		// time.Sleep(time.Second * 1)
		handlePriceChange(config, price)
	}
}

func calcDecimal(amount float64, decimal int) float64 {
	decimalPow := math.Pow(10, float64(decimal))
	return math.Floor(amount*decimalPow) / decimalPow
}

func startTrading(config *util.TradeConfig, price float64) {
	config.PriceOrder, config.LastPrice = price, price
	config.Amount = calcDecimal(config.Doller*float64(config.Times)/price, config.Decimal)
	fmt.Println(" * Trade start * ", price, config.Amount)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		util.PlaceOrder(config.Amount, config.Symbol, "BUY", "LONG", "MARKET", "")
	}()

	go func() {
		defer wg.Done()
		util.PlaceOrder(config.Amount, config.Symbol, "SELL", "SHORT", "MARKET", "")
	}()

	// 阻塞到两条 goroutine 都结束
	wg.Wait()
}

func handlePriceChange(config *util.TradeConfig, price float64) {
	// 检查价格变化是否达到处理阈值
	if !shouldProcessPrice(config, price) {
		return // 价格变化不够大，跳过处理
	}

	// 如果价格订单为0，则开始交易
	if config.PriceOrder <= 0 {
		startTrading(config, price)
	}

	checkAndResetTrades(config)
}

// 设置一个 委托挂单的函数
func PlaceLimitOrder(config *util.TradeConfig, price float64) {

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		util.PlaceOrder(config.Amount, config.Symbol, "BUY", "LONG", "MARKET", "")
	}()

	go func() {
		defer wg.Done()
		util.PlaceOrder(config.Amount, config.Symbol, "SELL", "SHORT", "MARKET", "")
	}()

	// 阻塞到两条 goroutine 都结束
	wg.Wait()
}

func checkAndResetTrades(config *util.TradeConfig) {

	positions, err := util.GetPositions()
	if err != nil {
		fmt.Println("获取仓位信息失败: ", err)
		return
	}

	// 过滤所有 强平价格不为0 的仓位
	filteredPositions := []util.Position{}
	for _, position := range positions {
		if position.LiquidationPrice != 0 {
			filteredPositions = append(filteredPositions, position)
		}
	}

	// 先获取委托列表
	openOrders, err := util.GetOpenOrders()
	if err != nil {
		fmt.Println("获取委托列表失败: ", err)
		return
	}

	fmt.Println("委托列表: ", openOrders)
	fmt.Println("仓位列表: ", filteredPositions)

	if len(openOrders) == 0 {
		// 那么开始按照强平价格 设置 委托
		for _, position := range filteredPositions {
			if position.PositionSide == "LONG" {
				util.PlaceOrder(config.Amount, config.Symbol, "SELL", "SHORT", "LIMIT", "0.21")
			} else {
				util.PlaceOrder(config.Amount, config.Symbol, "BUY", "LONG", "LIMIT", "0.175")
			}
		}
	}

}

func main() {
	cache.InitGlobalCache()

	fmt.Println("=== 对冲交易系统启动 ===")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 对冲交易配置
	hedgeConfigs := util.TradeConfig{
		Symbol:       "DOGEUSDT",
		Times:        5,
		Decimal:      4,
		Liquidation:  0.001,
		Doller:       2,
		MarginType:   "ISOLATED",
		TargetProfit: 0.01, // 1%目标盈利
	}

	func(cfg util.TradeConfig) {
		if initializeTrading(&cfg) {
			fmt.Println("初始化对冲交易完成=====>")
			subscribeAndHandleHedgePrice(&cfg)
		}
	}(hedgeConfigs)

	<-ctx.Done()
	fmt.Println("=== 对冲交易系统结束 ===")
	stop()
}
