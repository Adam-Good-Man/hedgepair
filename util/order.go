package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// 设置杠杆倍率
func SetLeverage(symbol string, leverage int) bool {

	config := GetConfig()

	apiUrl := FAPI + API_leverage

	// 创建请求数据
	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("leverage", strconv.Itoa(leverage))
	data.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	// 生成签名
	signature := GenerateSignature(data.Encode(), config.SECRET_KEY)
	data.Set("signature", signature)

	// 创建请求
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MBX-APIKEY", config.APIKEY)

	// Send request
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}
	defer response.Body.Close()

	// Handle response
	if response.StatusCode == http.StatusOK {
		return true
	} else {
		fmt.Printf("Failed to set leverage: %s\n", response.Status)
		return false
	}
}

// 下单
func PlaceOrder(amount float64, symbol, side, positionSide, orderType, price string) {

	config := GetConfig()
	precisionInfo, err := GetSymbolPrecision(symbol)

	if err != nil {
		fmt.Println("get precision error ", err)
		return
	}

	apiUrl := FAPI + API_order

	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("side", side) // BUY 或 SELL
	data.Set("type", orderType)

	if price != "" {
		data.Set("price", price)
		data.Set("timeInForce", "GTC")
	}

	data.Set("positionSide", positionSide)
	data.Set("quantity", strconv.FormatFloat(amount, 'f', precisionInfo.QuantityPrecision, 64))
	data.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	// 生成签名
	signature := GenerateSignature(data.Encode(), config.SECRET_KEY)
	data.Set("signature", signature)

	// 创建并发送请求
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MBX-APIKEY", config.APIKEY)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {

	} else {

		// 读取响应体
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return
		}

		fmt.Printf("Failed to place order:: %s\n", string(body)) // 打印响应体中的错误信息
	}
}

// 设置 全仓 - 逐仓
// setMarginType 设置保证金模式
// 全仓（ CROSSED ）或逐仓（ ISOLATED ）
func SetMarginType(symbol string, marginType string) bool {
	config := GetConfig()

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("marginType", marginType)
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	signature := GenerateSignature(params.Encode(), config.SECRET_KEY)
	params.Set("signature", signature)

	req, err := http.NewRequest("POST", FAPI+API_marginType, strings.NewReader(params.Encode()))
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	req.Header.Set("X-MBX-APIKEY", config.APIKEY)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return true
	} else {
		return false
	}
}

// GetPositions 获取当前所有仓位信息
func GetPositions() ([]Position, error) {
	config := GetConfig()

	// 创建请求参数
	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	queryString := fmt.Sprintf("timestamp=%s", timestamp)

	// 生成签名
	signature := GenerateSignature(queryString, config.SECRET_KEY)
	fullQueryString := fmt.Sprintf("%s&signature=%s", queryString, signature)

	// 构建完整的API URL
	apiUrl := fmt.Sprintf("%s%s?%s", FAPI, API_position, fullQueryString)

	// 创建GET请求
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("X-MBX-APIKEY", config.APIKEY)

	// 发送请求
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer response.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", response.StatusCode, string(body))
	}

	// 解析JSON响应
	var positions []Position
	err = json.Unmarshal(body, &positions)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return positions, nil
}

// GetPositionBySymbol 获取指定交易对的仓位信息
func GetPositionBySymbol(symbol string) (*Position, error) {
	positions, err := GetPositions()
	if err != nil {
		return nil, err
	}

	for _, position := range positions {
		if position.Symbol == symbol {
			return &position, nil
		}
	}

	return nil, fmt.Errorf("未找到交易对 %s 的仓位信息", symbol)
}

// PrintPositions 打印所有仓位信息
func PrintPositions() {
	positions, err := GetPositions()
	if err != nil {
		fmt.Printf("获取仓位信息失败: %v\n", err)
		return
	}

	if len(positions) == 0 {
		fmt.Println("当前没有持仓")
		return
	}

	fmt.Println("=== 当前仓位信息 ===")
	for _, pos := range positions {
		// 只显示有持仓的交易对
		if pos.PositionAmt != 0 {
			fmt.Printf("交易对: %s\n", pos.Symbol)
			fmt.Printf("  持仓数量: %.4f\n", pos.PositionAmt)
			fmt.Printf("  开仓价格: %.4f\n", pos.EntryPrice)
			fmt.Printf("  标记价格: %.4f\n", pos.MarkPrice)
			fmt.Printf("  未实现盈亏: %.4f\n", pos.UnRealizedPnl)
			fmt.Printf("  杠杆倍数: %d\n", pos.Leverage)
			fmt.Printf("  保证金类型: %s\n", pos.MarginType)
			fmt.Printf("  仓位方向: %s\n", pos.PositionSide)
			fmt.Printf("  强平价格: %.4f\n", pos.LiquidationPrice)
			fmt.Printf("  名义价值: %.4f\n", pos.Notional)
			fmt.Println("  ---")
		}
	}
}

// CloseAllPositions 市价全平所有持仓
func CloseAllPositions() error {
	positions, err := GetPositions()
	if err != nil {
		return fmt.Errorf("获取仓位信息失败: %v", err)
	}

	if len(positions) == 0 {
		fmt.Println("当前没有持仓，无需平仓")
		return nil
	}

	fmt.Println("=== 开始全平所有持仓 ===")
	successCount := 0
	errorCount := 0

	for _, pos := range positions {
		// 只处理有持仓的交易对
		if pos.PositionAmt != 0 {
			err := closePositionBySymbol(pos.Symbol, pos.PositionAmt, pos.PositionSide)
			if err != nil {
				fmt.Printf("平仓失败 %s: %v\n", pos.Symbol, err)
				errorCount++
			} else {
				fmt.Printf("平仓成功 %s: 数量 %.4f\n", pos.Symbol, pos.PositionAmt)
				successCount++
			}
		}
	}

	fmt.Printf("=== 全平完成 ===\n")
	fmt.Printf("成功平仓: %d 个\n", successCount)
	fmt.Printf("平仓失败: %d 个\n", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("部分平仓失败，请检查")
	}

	return nil
}

// ClosePositionBySymbol 平掉指定交易对的持仓
func ClosePositionBySymbol(symbol string) error {
	position, err := GetPositionBySymbol(symbol)
	if err != nil {
		return fmt.Errorf("获取 %s 仓位信息失败: %v", symbol, err)
	}

	if position.PositionAmt == 0 {
		return fmt.Errorf("%s 当前没有持仓", symbol)
	}

	return closePositionBySymbol(symbol, position.PositionAmt, position.PositionSide)
}

// closePositionBySymbol 内部函数，执行实际的平仓操作
func closePositionBySymbol(symbol string, positionAmt float64, positionSide string) error {
	config := GetConfig()

	// 确定平仓方向
	var side string
	if positionAmt > 0 {
		// 多头持仓，需要卖出平仓
		side = "SELL"
	} else {
		// 空头持仓，需要买入平仓
		side = "BUY"
	}

	// 获取交易对精度信息
	precisionInfo, err := GetSymbolPrecision(symbol)
	if err != nil {
		return fmt.Errorf("获取精度信息失败: %v", err)
	}

	apiUrl := FAPI + API_order

	// 创建请求参数
	data := url.Values{}
	data.Set("symbol", symbol)
	data.Set("side", side)
	data.Set("type", "MARKET") // 市价单
	data.Set("positionSide", positionSide)

	// 使用持仓数量的绝对值
	quantity := math.Abs(positionAmt)
	data.Set("quantity", strconv.FormatFloat(quantity, 'f', precisionInfo.QuantityPrecision, 64))
	data.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	// 生成签名
	signature := GenerateSignature(data.Encode(), config.SECRET_KEY)
	data.Set("signature", signature)

	// 创建并发送请求
	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-MBX-APIKEY", config.APIKEY)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer response.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	if response.StatusCode == http.StatusOK {
		return nil // 平仓成功
	} else {
		return fmt.Errorf("平仓失败，状态码: %d, 响应: %s", response.StatusCode, string(body))
	}
}

// // 设置 保证金
// func AdjustPositionMargin(symbol string, amount float64, adjustType int) {
// 	config := GetConfig()

// 	precisionInfo, err := GetSymbolPrecision(symbol)

// 	if err != nil {
// 		fmt.Println("get precision error ", err)
// 		return
// 	}

// 	// 创建请求参数
// 	data := url.Values{}
// 	data.Set("symbol", symbol)
// 	data.Set("amount", strconv.FormatFloat(amount, 'f', precisionInfo.PricePrecision, 64))
// 	data.Set("type", strconv.Itoa(adjustType))
// 	data.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

// 	// 生成签名
// 	signature := GenerateSignature(config.SECRET_KEY, data.Encode())
// 	data.Set("signature", signature)

// 	// 创建POST请求
// 	req, err := http.NewRequest("POST", FAPI+API_positionMargin, strings.NewReader(data.Encode()))
// 	if err != nil {
// 		log.Fatal("Error creating request: ", err)
// 	}

// 	// 设置请求头
// 	req.Header.Set("X-MBX-APIKEY", config.APIKEY)
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

// 	// 发送请求
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatal("Error sending request: ", err)
// 	}
// 	defer resp.Body.Close()

// 	// 读取响应
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Fatal("Error reading response body: ", err)
// 	}

// 	log.Println("Response: ", string(body))
// }
