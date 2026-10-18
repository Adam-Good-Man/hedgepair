package util

import (
	"log"

	"github.com/spf13/viper"
)

const (
	TEST_API   = "https://testnet.binance.vision"
	STREAM_WSS = "stream.binance.com:9443"
	FAPI       = "https://fapi.binance.com"
	API_0      = "https://api.binance.com"
	API_1      = "https://api1.binance.com"
	API_2      = "https://api2.binance.com"
	API_3      = "https://api3.binance.com"
	API_4      = "https://api4.binance.com"
)

const (
	API_exchangeInfo    = "/fapi/v1/exchangeInfo"
	API_leverage        = "/fapi/v1/leverage"
	API_price           = "/api/v3/ticker/price"
	API_order           = "/fapi/v1/order"
	API_marginType      = "/fapi/v1/marginType"
	API_leverageBracket = "/fapi/v1/leverageBracket"
	API_positionMargin  = "/fapi/v1/positionMargin"
	API_position        = "/fapi/v2/positionRisk"
	//当前最优的挂单(最高买单，最低卖单) GET
	API_bookTicker = "/api/v3/ticker/bookTicker"
	// 委托列表
	API_openOrders = "/fapi/v1/openOrders"
)

type Config struct {
	APIKEY     string `mapstructure:"API_KEY"`
	SECRET_KEY string `mapstructure:"SECRET_KEY"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

var config Config

func init() {
	var err error
	config, err = LoadConfig(".")
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}
}

func GetConfig() Config {
	return config
}
