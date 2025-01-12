package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"receipt-processor/internal/api/handlers"
	"receipt-processor/internal/api/middleware"
	"receipt-processor/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize dependencies
	receiptService := service.NewReceiptService()
	receiptHandler := handlers.NewReceiptHandler(receiptService)

	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.Logger)
	router.Use(middleware.ContentType)
	router.Use(middleware.Recover)

	// Register routes
	router.HandleFunc("/receipts/process", receiptHandler.ProcessReceipt).Methods(http.MethodPost)
	router.HandleFunc("/receipts/{id}/points", receiptHandler.GetPoints).Methods(http.MethodGet)

	// Configure server
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Server starting on port 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
