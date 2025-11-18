package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"
	"trading-go/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	Stocks = map[string]*models.Stock{
		"AAPL": {Symbol: "AAPL", Price: 175.50, Change: 0},
		"TSLA": {Symbol: "TSLA", Price: 242.80, Change: 0},
		"AMZN": {Symbol: "AMZN", Price: 145.30, Change: 0},
		"INFY": {Symbol: "INFY", Price: 18.75, Change: 0},
		"TCS":  {Symbol: "TCS", Price: 3645.20, Change: 0},
	}
	Orders         []models.Order
	OrdersMutex    sync.RWMutex
	StocksMutex    sync.RWMutex
	OrderIDCounter = 1

	Upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins
		},
	}
	Clients      = make(map[*websocket.Conn]bool)
	ClientsMutex sync.Mutex
	Broadcast    = make(chan map[string]*models.Stock)
)

func GetPrices(c *gin.Context) {
	StocksMutex.RLock()
	defer StocksMutex.RUnlock()

	priceList := make([]*models.Stock, 0, len(Stocks))
	for _, stock := range Stocks {
		priceList = append(priceList, stock)
	}

	c.JSON(http.StatusOK, priceList)
}

func CreateOrder(c *gin.Context) {
	username, _ := c.Get("username")

	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.Symbol == "" || order.Side == "" || order.Quantity <= 0 || order.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order data"})
		return
	}

	if order.Side != "buy" && order.Side != "sell" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Side must be 'buy' or 'sell'"})
		return
	}

	StocksMutex.RLock()
	_, exists := Stocks[order.Symbol]
	StocksMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock symbol"})
		return
	}

	OrdersMutex.Lock()
	order.ID = OrderIDCounter
	OrderIDCounter++
	order.Username = username.(string)
	order.Time = time.Now().Format("2006-01-02 15:04:05")
	Orders = append(Orders, order)
	OrdersMutex.Unlock()

	c.JSON(http.StatusCreated, order)
}

func GetOrders(c *gin.Context) {
	username, _ := c.Get("username")

	OrdersMutex.RLock()
	defer OrdersMutex.RUnlock()

	userOrders := []models.Order{}
	for _, order := range Orders {
		if order.Username == username.(string) {
			userOrders = append(userOrders, order)
		}
	}

	c.JSON(http.StatusOK, userOrders)
}

func GetAllOrders(c *gin.Context) {
	OrdersMutex.RLock()
	defer OrdersMutex.RUnlock()

	c.JSON(http.StatusOK, Orders)
}

func HandleWebSocket(c *gin.Context) {
	conn, err := Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	ClientsMutex.Lock()
	Clients[conn] = true
	ClientsMutex.Unlock()

	StocksMutex.RLock()
	initialData := make(map[string]*models.Stock)
	for k, v := range Stocks {
		initialData[k] = &models.Stock{
			Symbol: v.Symbol,
			Price:  v.Price,
			Change: v.Change,
		}
	}
	StocksMutex.RUnlock()

	if err := conn.WriteJSON(initialData); err != nil {
		log.Println("Error sending initial data:", err)
	}

	// client disconnect
	defer func() {
		ClientsMutex.Lock()
		delete(Clients, conn)
		ClientsMutex.Unlock()
		conn.Close()
	}()

	// connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func BroadcastPrices() {
	for {
		stocks := <-Broadcast

		ClientsMutex.Lock()
		for client := range Clients {
			err := client.WriteJSON(stocks)
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
		ClientsMutex.Unlock()
	}
}
