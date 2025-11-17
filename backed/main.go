package main

import (
	// "encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Stock represents a stock with its current price
type Stock struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change"`
}

// Order represents a trading order
type Order struct {
	ID       int     `json:"id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"` // "buy" or "sell"
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Time     string  `json:"time"`
}

// Global state
var (
	stocks = map[string]*Stock{
		"AAPL": {Symbol: "AAPL", Price: 175.50, Change: 0},
		"TSLA": {Symbol: "TSLA", Price: 242.80, Change: 0},
		"AMZN": {Symbol: "AMZN", Price: 145.30, Change: 0},
		"INFY": {Symbol: "INFY", Price: 18.75, Change: 0},
		"TCS":  {Symbol: "TCS", Price: 3645.20, Change: 0},
	}
	orders         []Order
	ordersMutex    sync.RWMutex
	stocksMutex    sync.RWMutex
	orderIDCounter = 1

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
	broadcast    = make(chan map[string]*Stock)
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Start price update goroutine
	go updatePrices()
	go handleBroadcast()

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// API routes
	r.GET("/prices", getPrices)
	r.POST("/orders", createOrder)
	r.GET("/orders", getOrders)
	r.GET("/ws", handleWebSocket)

	log.Println("Server starting on :8080")
	r.Run(":8080")
}

// getPrices returns current stock prices
func getPrices(c *gin.Context) {
	stocksMutex.RLock()
	defer stocksMutex.RUnlock()

	priceList := make([]*Stock, 0, len(stocks))
	for _, stock := range stocks {
		priceList = append(priceList, stock)
	}

	c.JSON(http.StatusOK, priceList)
}

// createOrder handles order creation
func createOrder(c *gin.Context) {
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate order
	if order.Symbol == "" || order.Side == "" || order.Quantity <= 0 || order.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})
		return
	}

	if order.Side != "buy" && order.Side != "sell" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Side must be 'buy' or 'sell'"})
		return
	}

	stocksMutex.RLock()
	_, exists := stocks[order.Symbol]
	stocksMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock symbol"})
		return
	}

	ordersMutex.Lock()
	order.ID = orderIDCounter
	orderIDCounter++
	order.Time = time.Now().Format("2006-01-02 15:04:05")
	orders = append(orders, order)
	ordersMutex.Unlock()

	c.JSON(http.StatusCreated, order)
}

// getOrders returns all orders
func getOrders(c *gin.Context) {
	ordersMutex.RLock()
	defer ordersMutex.RUnlock()

	c.JSON(http.StatusOK, orders)
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	// Send initial prices
	stocksMutex.RLock()
	initialData := make(map[string]*Stock)
	for k, v := range stocks {
		initialData[k] = &Stock{
			Symbol: v.Symbol,
			Price:  v.Price,
			Change: v.Change,
		}
	}
	stocksMutex.RUnlock()

	if err := conn.WriteJSON(initialData); err != nil {
		log.Println("Error sending initial data:", err)
	}

	// Handle client disconnect
	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		conn.Close()
	}()

	// Keep connection alive and listen for messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// updatePrices simulates price changes
func updatePrices() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stocksMutex.Lock()
		updatedStocks := make(map[string]*Stock)

		for symbol, stock := range stocks {
			// Random price change between -2% and +2%
			changePercent := (rand.Float64()*4 - 2) / 100
			newPrice := stock.Price * (1 + changePercent)

			// Round to 2 decimal places
			newPrice = float64(int(newPrice*100)) / 100

			stock.Price = newPrice
			stock.Change = changePercent * 100

			updatedStocks[symbol] = &Stock{
				Symbol: stock.Symbol,
				Price:  stock.Price,
				Change: stock.Change,
			}
		}
		stocksMutex.Unlock()

		// Broadcast to all clients
		broadcast <- updatedStocks
	}
}

// handleBroadcast sends updates to all connected clients
func handleBroadcast() {
	for {
		stocks := <-broadcast

		clientsMutex.Lock()
		for client := range clients {
			err := client.WriteJSON(stocks)
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientsMutex.Unlock()
	}
}
