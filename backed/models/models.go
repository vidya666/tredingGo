package models

// Stock represents a stock with its current price
type Stock struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change"`
}

// Order represents a trading order
type Order struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"` // "buy" or "sell"
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Time     string  `json:"time"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Message  string `json:"message"`
}
