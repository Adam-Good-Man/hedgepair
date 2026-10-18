# hedgepair

币安 U 本位合约对冲程序。启动后对同一个交易对同时建立多头和空头，再按当前仓位补挂限价单。

量化交易的一个试验品。默认会向币安实盘发市价单，不要用真实资金运行。

## 它做什么

`main` 里写死了一组交易参数（当前是 `DOGEUSDT`、5 倍、逐仓、名义本金 2 USDT）。进程起来之后：

1. 把该交易对设为逐仓，并设置杠杆。
2. 订阅币安成交 WebSocket，收到价格后按交易对精度过滤掉过小的跳动。
3. 第一次有效价格到来时，同时发两笔市价单：买入开多、卖出开空。数量由 `名义本金 × 杠杆 / 价格` 算出，再按精度截断。
4. 之后每次价格跨过最小变动单位，就拉取持仓和当前委托。没有委托时，按仓位方向各挂一笔限价单。

收到 `SIGINT` 或 `SIGTERM` 后退出。退出不会自动平仓。

## 架构

```text
main
  ├─ 初始化内存缓存
  ├─ 设置保证金模式 / 杠杆          util  ──HMAC──▶  fapi.binance.com
  └─ 价格循环
       ├─ 成交流                     util  ◀──WSS──  stream.binance.com
       ├─ 过滤过小的价格变化          main
       ├─ 首次：多空市价开仓          util
       └─ 之后：查仓位和委托，补限价单  util

cache   进程内 map，带过期时间。交易对精度缓存 36 小时。
util    配置、签名、REST、WebSocket、仓位和订单结构体。
```

对外只有两类连接：

| 方向 | 地址 | 用途 |
| --- | --- | --- |
| WebSocket | `wss://stream.binance.com:9443/ws/<symbol>@trade` | 最新成交价，断线后最多重连 10 次 |
| REST | `https://fapi.binance.com` | 交易规则、杠杆、保证金模式、下单、持仓、当前委托 |

私有 REST 请求用 API Key 放在 `X-MBX-APIKEY`，参数按币安规则做 HMAC-SHA256 签名。密钥从当前目录的 `app.env` 读取。

## 目录

```text
main.go              启动、价格过滤、开仓和对冲循环
cache/init.go        带读写锁的内存缓存
util/config.go       读取 app.env
util/sign.go         HMAC-SHA256
util/socket.go       成交价 WebSocket
util/order.go        杠杆、保证金模式、下单、持仓
util/exinfo.go       交易规则、精度、当前委托
util/type.go         配置和接口响应结构体
app.env.example      配置模板
```

## 配置

需要一个开启了合约交易权限、未开启提现的币安 API Key。

```bash
cp app.env.example app.env
```

`app.env`：

```env
API_KEY=your_binance_api_key
SECRET_KEY=your_binance_secret_key
```

交易参数在 `main.go` 的 `hedgeConfigs` 里，不在环境变量中。

## 运行

Go 1.21 或更高版本。

```bash
go run .
```

## 现状

- 只跑一个交易对，参数写在代码里， 但实际可以在数组中填入多组交易对。
- 没有委托时补的限价价格是写死的，没有按强平价计算。
- `GetLatestPrice` 保留了 REST 询价，主循环没有使用它。
- 程序不会在退出时平掉已有仓位。
