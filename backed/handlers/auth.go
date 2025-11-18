package handlers

import (
	"net/http"
	"sync"
	"trading-go/auth"
	"trading-go/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	// In-memory user storage
	Users      = make(map[string]string) // username -> hashed password
	UsersMutex sync.RWMutex
)

func init() {
	// Create a demo user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("demo123"), bcrypt.DefaultCost)
	Users["demo"] = string(hashedPassword)
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	UsersMutex.Lock()
	defer UsersMutex.Unlock()

	if _, exists := Users[req.Username]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	Users[req.Username] = string(hashedPassword)

	token, err := auth.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token:    token,
		Username: req.Username,
		Message:  "Registration successful",
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	UsersMutex.RLock()
	hashedPassword, exists := Users[req.Username]
	UsersMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := auth.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:    token,
		Username: req.Username,
		Message:  "Login successful",
	})
}

// GetProfile for current user's profile
func GetProfile(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{
		"username": username,
		"message":  "Profile retrieved successfully",
	})
}
