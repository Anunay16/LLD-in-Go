package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/anunay/logger"
	"github.com/anunay/logger/appender"
	"github.com/anunay/logger/formatter"
)

type customBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (c *customBuffer) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Write(p)
}

func (c *customBuffer) Close() error {
	return nil
}

func (c *customBuffer) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

func TestLogLevelsAndFiltering(t *testing.T) {
	cb := &customBuffer{}
	app := appender.NewCustomWriterAppender(cb, formatter.NewTextFormatter(false))

	log := logger.NewBuilder().
		SetLevel(logger.WarnLevel).
		SetAppender(app).
		Build()

	log.Debug("This debug log should not appear")
	log.Info("This info log should not appear")
	log.Warn("This warning log should appear")
	log.Error("This error log should appear")

	output := cb.String()
	if strings.Contains(output, "debug log") {
		t.Errorf("expected no debug log in output, got: %s", output)
	}
	if strings.Contains(output, "info log") {
		t.Errorf("expected no info log in output, got: %s", output)
	}
	if !strings.Contains(output, "warning log") {
		t.Errorf("expected warning log in output, got: %s", output)
	}
	if !strings.Contains(output, "error log") {
		t.Errorf("expected error log in output, got: %s", output)
	}
}

func TestJSONFormatterAndMetadataFields(t *testing.T) {
	cb := &customBuffer{}
	app := appender.NewCustomWriterAppender(cb, formatter.NewJSONFormatter(false))

	log := logger.NewBuilder().
		SetLevel(logger.DebugLevel).
		SetAppender(app).
		SetDefaultFields(map[string]any{"service": "payment-svc", "env": "prod"}).
		Build()

	userLog := log.With(logger.String("userId", "usr_12345"), logger.Int("attempt", 3))
	userLog.Info("User payment processed", logger.Float64("amount", 99.99), logger.Bool("success", true))

	output := cb.String()
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v, raw: %s", err, output)
	}

	if parsed["level"] != "INFO" {
		t.Errorf("expected level INFO, got %v", parsed["level"])
	}
	if parsed["message"] != "User payment processed" {
		t.Errorf("expected correct message, got %v", parsed["message"])
	}

	fields, ok := parsed["fields"].(map[string]any)
	if !ok {
		t.Fatalf("expected fields map, got %v", parsed["fields"])
	}

	if fields["service"] != "payment-svc" || fields["userId"] != "usr_12345" || fields["attempt"] != float64(3) || fields["amount"] != 99.99 || fields["success"] != true {
		t.Errorf("unexpected fields content: %+v", fields)
	}
}

func TestAsyncAppender(t *testing.T) {
	cb := &customBuffer{}
	rawApp := appender.NewCustomWriterAppender(cb, formatter.NewTextFormatter(false))
	asyncApp := appender.NewAsyncAppender(rawApp, 100, 2)

	log := logger.NewBuilder().
		SetLevel(logger.InfoLevel).
		SetAppender(asyncApp).
		Build()

	for i := 0; i < 50; i++ {
		log.Info(fmt.Sprintf("Async message %d", i))
	}

	// Close waits for all queue items to flush
	if err := log.Close(); err != nil {
		t.Fatalf("failed to close async logger: %v", err)
	}

	output := cb.String()
	for i := 0; i < 50; i++ {
		expected := "Async message "
		if !strings.Contains(output, expected) {
			t.Errorf("missing async message in output")
			break
		}
	}
}

func TestConcurrency(t *testing.T) {
	cb := &customBuffer{}
	app := appender.NewCustomWriterAppender(cb, formatter.NewJSONFormatter(false))

	log := logger.NewBuilder().
		SetLevel(logger.DebugLevel).
		SetAppender(app).
		Build()

	var wg sync.WaitGroup
	workers := 10
	logsPerWorker := 50

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerLogger := log.With(logger.Int("workerId", workerID))
			for i := 0; i < logsPerWorker; i++ {
				workerLogger.Info(fmt.Sprintf("Log entry %d", i))
			}
		}(w)
	}

	wg.Wait()
}
