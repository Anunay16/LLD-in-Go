package strategy

import (
	"context"
	"fmt"
	"time"
)

// TaskStrategy defines the execution logic for a specific type of work.
type TaskStrategy interface {
	Execute(ctx context.Context) error
}

// EmailTask sends an email.
type EmailTask struct {
	To      string
	Subject string
	Body    string
}

func NewEmailTask(to, subject, body string) *EmailTask {
	return &EmailTask{
		To:      to,
		Subject: subject,
		Body:    body,
	}
}

func (t *EmailTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("[EmailTask] Sending email to %s | Subject: %s\n", t.To, t.Subject)
		// Simulate network / processing latency
		time.Sleep(50 * time.Millisecond)
		return nil
	}
}

// ReportTask generates and publishes a business report.
type ReportTask struct {
	ReportName string
	Format     string
}

func NewReportTask(reportName, format string) *ReportTask {
	return &ReportTask{
		ReportName: reportName,
		Format:     format,
	}
}

func (t *ReportTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("[ReportTask] Generating %s report (%s format)...\n", t.ReportName, t.Format)
		time.Sleep(80 * time.Millisecond)
		return nil
	}
}

// NotificationTask sends push or webhook notifications.
type NotificationTask struct {
	Channel string
	Message string
}

func NewNotificationTask(channel, message string) *NotificationTask {
	return &NotificationTask{
		Channel: channel,
		Message: message,
	}
}

func (t *NotificationTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("[NotificationTask] [%s] Alert: %s\n", t.Channel, t.Message)
		time.Sleep(30 * time.Millisecond)
		return nil
	}
}
