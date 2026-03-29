package notificationservice

import "fmt"

type INotificationStrategy interface {
	SendNotification(content string)
}

// =========== Email Strategy ==============

type EmailStrategy struct {
	email string
}

func NewEmailStrategy(email string) *EmailStrategy {
	return &EmailStrategy{email: email}
}

func (s *EmailStrategy) SendNotification(content string) {
	fmt.Printf("sending EMail notification to %s: %s\n", s.email, content)
}

// =========== SMS Strategy ==============

type SMSStrategy struct {
	phoneNo string
}

func NewSMSStrategy(phoneNo string) *SMSStrategy {
	return &SMSStrategy{phoneNo: phoneNo}
}

func (s *SMSStrategy) SendNotification(content string) {
	fmt.Printf("sending SMS notification to %s: %s\n", s.phoneNo, content)
}
