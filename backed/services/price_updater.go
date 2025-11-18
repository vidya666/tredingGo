package services

import (
	"math/rand"
	"time"
	"trading-go/handlers"
	"trading-go/models"
)

func UpdatePrices() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		handlers.StocksMutex.Lock()
		updatedStocks := make(map[string]*models.Stock)

		for symbol, stock := range handlers.Stocks {
			// Random price change between -2% and +2%
			changePercent := (rand.Float64()*4 - 2) / 100
			newPrice := stock.Price * (1 + changePercent)

			// Round to 2 decimal places
			newPrice = float64(int(newPrice*100)) / 100

			stock.Price = newPrice
			stock.Change = changePercent * 100

			updatedStocks[symbol] = &models.Stock{
				Symbol: stock.Symbol,
				Price:  stock.Price,
				Change: stock.Change,
			}
		}
		handlers.StocksMutex.Unlock()

		handlers.Broadcast <- updatedStocks
	}
}
