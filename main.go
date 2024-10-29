package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	rdb            = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	timeoutSeconds = 60 * time.Second
	ctx            = context.Background()
)

type Message struct {
	ClientID int    `json:"client_id"`
	Content  string `json:"content"`
}

func sendMessage(c *gin.Context) {
	var message Message
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key := "message:" + strconv.Itoa(message.ClientID)
	err := rdb.SetEx(ctx, key, message.Content, timeoutSeconds).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Message sent"})
}

func receiveMessage(c *gin.Context) {
	clientID, err := strconv.Atoi(c.Param("client_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid client_id"})
		return
	}

	key := "message:" + strconv.Itoa(clientID)
	startTime := time.Now()

	for {
		elapsed := time.Since(startTime)
		if elapsed > timeoutSeconds {
			c.JSON(http.StatusOK, gin.H{"status": "No new messages"})
			return
		}

		message, err := rdb.Get(ctx, key).Result()
		if err == redis.Nil {
			time.Sleep(1 * time.Second) // Wait for 1 second before retrying
			continue
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve message"})
			return
		}

		rdb.Del(ctx, key)
		c.JSON(http.StatusOK, gin.H{"message": message})
		return
	}
}

func main() {
	router := gin.Default()
	router.POST("/send", sendMessage)
	router.GET("/receive/:client_id", receiveMessage)

	srv := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	// Run server in a goroutine so it doesn't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server started on http://localhost:8000")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Context with timeout to allow server to finish current operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited gracefully")
}
