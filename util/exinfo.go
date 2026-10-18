package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"example.com/m/cache"
)

// 获取 交易对 下单要求精度
func GetSymbolPrecision(symbol string) (*Precision, error) {
	// apiUrl := "https://fapi.binance.com/fapi/v1/exchangeInfo"
	values, found := cache.GlobalCache.Get("Precision_" + symbol)
	if found {
		precisionInfo := &Precision{}
		if precision, ok := values.(Precision); ok {

			precisionInfo.BaseAssetPrecision = precision.BaseAssetPrecision
			precisionInfo.PricePrecision = precision.PricePrecision
			precisionInfo.QuantityPrecision = precision.QuantityPrecision
			precisionInfo.QuotePrecision = precision.QuotePrecision
			precisionInfo.Symbol = precision.Symbol

			return precisionInfo, nil
		}

	} else {
		apiUrl := FAPI + API_exchangeInfo

		resp, err := http.Get(apiUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var exchangeInfo ExchangeInfo
		err = json.Unmarshal(body, &exchangeInfo)
		if err != nil {
			return nil, err
		}

		for _, s := range exchangeInfo.Symbols {
			if s.Symbol == symbol {

				cache.GlobalCache.Set("Precision_"+symbol, s, time.Hour*36)

				return &s, nil // 或 s.BaseAssetPrecision 根据需求选择
			}
		}

		return nil, fmt.Errorf("Symbol not found")

	}

	return nil, nil

}

// 通过 REST API 的方式暂未被使用，目前 通过 socket 已经可以拿到最新价格推送
// getLatestPrice 发送请求并解析返回的价格
func GetLatestPrice(symbol string) (string, error) {

	url := fmt.Sprintf("%s%s?symbol=%s", API_0, API_price, symbol)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tickerResponse TickerResponse
	err = json.Unmarshal(body, &tickerResponse)
	if err != nil {
		return "", err
	}

	return tickerResponse.Price, nil
}

// 获取 交易对 的 一些 信息
func GetMaxLeverage(symbol string) *LeverageInfo {

	values, found := cache.GlobalCache.Get("Leverage_" + symbol)
	if found {
		everageInfo := &LeverageInfo{}
		if leverage, ok := values.(LeverageInfo); ok {

			everageInfo.Leverage = leverage.Leverage
			everageInfo.MaintMarginRatio = leverage.MaintMarginRatio

			return everageInfo
		}

	} else {

		config := GetConfig()

		info := &LeverageInfo{}

		// apiUrl := config.FAPI + API_leverageBracket

		timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
		queryString := fmt.Sprintf("symbol=%s&timestamp=%s", symbol, timestamp)
		signature := GenerateSignature(queryString, config.SECRET_KEY)

		apiUrl := fmt.Sprintf("%s%s?%s&signature=%s", FAPI, API_leverageBracket, queryString, signature)

		client := &http.Client{}
		req, err := http.NewRequest("GET", apiUrl, nil)
		if err != nil {
			log.Fatal("Error creating request: ", err)
		}

		req.Header.Set("X-MBX-APIKEY", config.APIKEY)

		response, err := client.Do(req)
		if err != nil {
			log.Fatal("Error sending request: ", err)
		}
		defer response.Body.Close()

		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal("Error reading response body: ", err)
		}

		var leverageInfo []LeverageBracket
		err = json.Unmarshal(responseBody, &leverageInfo)
		if err != nil {
			log.Fatal("Error unmarshalling JSON: ", err)
		}

		maxLeverage := 0
		for _, bracket := range leverageInfo {
			for _, br := range bracket.Brackets {
				if br.Leverage > maxLeverage {
					maxLeverage = br.Leverage

					info.Leverage = br.Leverage
					info.MaintMarginRatio = float64(br.MaintMarginRatio)
				}
			}
		}

		return info

	}

	return nil

}

func GetOpenOrders() ([]OpenOrder, error) {

	config := GetConfig()

	timestamp := strconv.FormatInt(time.Now().Unix()*1000, 10)
	queryString := fmt.Sprintf("timestamp=%s", timestamp)
	signature := GenerateSignature(queryString, config.SECRET_KEY)

	apiUrl := fmt.Sprintf("%s%s?%s&signature=%s", FAPI, API_openOrders, queryString, signature)

	client := &http.Client{}
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("X-MBX-APIKEY", config.APIKEY)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var openOrders []OpenOrder
	err = json.Unmarshal(body, &openOrders)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return openOrders, nil
}
