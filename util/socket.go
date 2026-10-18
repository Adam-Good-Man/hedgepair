package util

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	Symbol       string
	PriceChan    chan<- float64
	MaxRetries   int
	RetryDelay   time.Duration
	PingPeriod   time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// socket 最新价格推送 - 改进版本，支持自动重连
func StreamPrice(symbol string, priceChan chan<- float64) {
	config := &WebSocketConfig{
		Symbol:       symbol,
		PriceChan:    priceChan,
		MaxRetries:   10,               // 最大重试次数
		RetryDelay:   5 * time.Second,  // 重试延迟
		PingPeriod:   30 * time.Second, // 心跳间隔
		ReadTimeout:  60 * time.Second, // 读取超时
		WriteTimeout: 10 * time.Second, // 写入超时
	}

	streamPriceWithReconnect(config)
}

// streamPriceWithReconnect 带重连功能的价格流
func streamPriceWithReconnect(config *WebSocketConfig) {
	symbol := strings.ToLower(config.Symbol)
	path := fmt.Sprintf("/ws/%s@trade", symbol)

	retryCount := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		if retryCount >= config.MaxRetries {
			log.Printf("达到最大重试次数 %d，停止重连 %s", config.MaxRetries, config.Symbol)
			return
		}

		log.Printf("正在连接 %s WebSocket... (尝试 %d/%d)", config.Symbol, retryCount+1, config.MaxRetries)

		// 建立WebSocket连接
		conn, err := establishConnection(path)
		if err != nil {
			log.Printf("连接失败 %s: %v", config.Symbol, err)
			retryCount++
			time.Sleep(config.RetryDelay)
			continue
		}

		log.Printf("成功连接 %s WebSocket", config.Symbol)
		retryCount = 0 // 重置重试计数

		// 启动心跳和消息处理
		var wg sync.WaitGroup
		wg.Add(2)

		// 启动心跳
		go func() {
			defer wg.Done()
			pingTicker := time.NewTicker(config.PingPeriod)
			defer pingTicker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-pingTicker.C:
					if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(config.WriteTimeout)); err != nil {
						log.Printf("发送心跳失败 %s: %v", config.Symbol, err)
						return
					}
				}
			}
		}()

		// 启动消息处理
		go func() {
			defer wg.Done()
			handleMessages(conn, config, ctx)
		}()

		// 等待连接断开
		wg.Wait()
		conn.Close()

		log.Printf("%s WebSocket连接断开，准备重连...", config.Symbol)
		time.Sleep(config.RetryDelay)
	}
}

// establishConnection 建立WebSocket连接
func establishConnection(path string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "wss", Host: STREAM_WSS, Path: path}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	// 设置读写超时
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	return conn, nil
}

// handleMessages 处理WebSocket消息
func handleMessages(conn *websocket.Conn, config *WebSocketConfig, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))

			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket意外关闭 %s: %v", config.Symbol, err)
				} else {
					log.Printf("读取消息失败 %s: %v", config.Symbol, err)
				}
				return
			}

			// 处理ping消息
			if messageType == websocket.PingMessage {
				log.Printf("收到ping消息 %s，发送pong", config.Symbol)
				conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
				if err := conn.WriteMessage(websocket.PongMessage, message); err != nil {
					log.Printf("发送pong失败 %s: %v", config.Symbol, err)
					return
				}
				continue
			}

			// 处理普通消息
			if messageType == websocket.TextMessage {
				processPriceMessage(message, config)
			}
		}
	}
}

// processPriceMessage 处理价格消息
func processPriceMessage(message []byte, config *WebSocketConfig) {
	var msg map[string]interface{}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("JSON解析失败 %s: %v", config.Symbol, err)
		return
	}

	if price, ok := msg["p"].(string); ok {
		priceFloat, err := strconv.ParseFloat(price, 64)
		if err != nil {
			log.Printf("价格解析失败 %s: %v", config.Symbol, err)
			return
		}

		// 非阻塞发送价格
		select {
		case config.PriceChan <- priceFloat:
			// 成功发送
		default:
			// 通道已满，记录警告但不阻塞
			// log.Printf("价格通道已满，丢弃价格数据 %s: %.4f", config.Symbol, priceFloat)
		}
	}
}

// WebSocketManager WebSocket连接管理器
type WebSocketManager struct {
	connections map[string]*WebSocketConfig
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWebSocketManager 创建新的WebSocket管理器
func NewWebSocketManager() *WebSocketManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketManager{
		connections: make(map[string]*WebSocketConfig),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// AddConnection 添加新的WebSocket连接
func (wm *WebSocketManager) AddConnection(symbol string, priceChan chan<- float64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	config := &WebSocketConfig{
		Symbol:       symbol,
		PriceChan:    priceChan,
		MaxRetries:   10,
		RetryDelay:   5 * time.Second,
		PingPeriod:   30 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	wm.connections[symbol] = config

	// 启动连接
	go streamPriceWithReconnect(config)

	log.Printf("已添加 %s 到WebSocket管理器", symbol)
}

// RemoveConnection 移除WebSocket连接
func (wm *WebSocketManager) RemoveConnection(symbol string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if _, exists := wm.connections[symbol]; exists {
		delete(wm.connections, symbol)
		log.Printf("已从WebSocket管理器移除 %s", symbol)
	}
}

// GetConnectionCount 获取当前连接数量
func (wm *WebSocketManager) GetConnectionCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.connections)
}

// Stop 停止所有连接
func (wm *WebSocketManager) Stop() {
	wm.cancel()
	wm.mu.Lock()
	defer wm.mu.Unlock()

	log.Printf("正在停止 %d 个WebSocket连接", len(wm.connections))
	wm.connections = make(map[string]*WebSocketConfig)
}

// StreamPriceWithManager 使用管理器的方式启动价格流
func StreamPriceWithManager(symbol string, priceChan chan<- float64) {
	// 这里可以创建一个全局管理器或者使用单例模式
	// 为了简化，我们仍然使用原来的方式，但可以很容易地扩展
	StreamPrice(symbol, priceChan)
}
