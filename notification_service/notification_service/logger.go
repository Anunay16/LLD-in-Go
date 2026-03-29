package notificationservice

import (
	"fmt"
	types "notification-service/types"
)

// concrete observer, implements IObserver interface
type Logger struct {
	notificationObservable *NotificationObservable
}

func NewLogger() *Logger {
	observable := GetNotificationServiceInstance().GetObservable()
	logger := &Logger{notificationObservable: observable}
	logger.notificationObservable.AddObserver(logger)
	return logger
}

func (l *Logger) Update(notification types.INotification) {
	fmt.Printf("Logging new notification: %s\n", notification.GetContent())
}
