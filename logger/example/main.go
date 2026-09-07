package main

import (
	"fmt"
	"time"

	"github.com/anunay/logger"
	"github.com/anunay/logger/appender"
	"github.com/anunay/logger/formatter"
)

func main() {
	fmt.Println("=== 1. Text Formatter with Metadata ===")
	consoleAppender := appender.NewConsoleAppender(formatter.NewTextFormatter(true))

	textLogger := logger.NewBuilder().
		SetLevel(logger.DebugLevel).
		SetAppender(consoleAppender).
		SetDefaultFields(map[string]any{
			"app": "ecommerce-service",
			"env": "production",
		}).
		Build()

	textLogger.Info("Application initialized successfully", logger.String("version", "v1.2.0"))
	textLogger.Debug("Connecting to database", logger.String("host", "10.0.0.1"), logger.Int("port", 5432))
	textLogger.Warn("Cache connection high latency", logger.Duration("latency", 250*time.Millisecond))

	fmt.Println("\n=== 2. JSON Formatter with Structured Metadata ===")
	jsonConsoleApp := appender.NewConsoleAppender(formatter.NewJSONFormatter(true))

	jsonLogger := logger.NewBuilder().
		SetLevel(logger.InfoLevel).
		SetAppender(jsonConsoleApp).
		Build()

	jsonLogger.Info("User login attempt",
		logger.String("user_id", "usr_9981"),
		logger.String("username", "john_doe"),
		logger.String("ip", "192.168.1.1"),
	)

	fmt.Println("\n=== 3. Contextual / Child Loggers (With) ===")
	reqLogger := textLogger.With(
		logger.String("trace_id", "trace-abc-123"),
		logger.String("user_ip", "192.168.1.50"),
	)

	reqLogger.Info("Order checkout initiated", logger.String("cart_id", "cart_771"))
	reqLogger.Info("Payment verified", logger.Float64("amount", 149.99), logger.String("currency", "USD"))

	fmt.Println("\n=== 4. Multi-Appender & File Output + Async Processing ===")
	fileApp, err := appender.NewFileAppender("logs/app.log", formatter.NewJSONFormatter(false))
	if err != nil {
		fmt.Printf("File appender error: %v\n", err)
		return
	}
	defer fileApp.Close()

	multiApp := appender.NewMultiAppender(consoleAppender, fileApp)
	asyncApp := appender.NewAsyncAppender(multiApp, 500, 2)
	defer asyncApp.Close()

	asyncLogger := logger.NewBuilder().
		SetLevel(logger.InfoLevel).
		SetAppender(asyncApp).
		Build()

	for i := 1; i <= 5; i++ {
		asyncLogger.Info("Batch processing event", logger.Int("item_index", i), logger.String("status", "processed"))
	}

	// Give async worker time to flush
	time.Sleep(50 * time.Millisecond)

	fmt.Println("\nDemo completed successfully! Check logs/app.log for persisted logs.")
}
