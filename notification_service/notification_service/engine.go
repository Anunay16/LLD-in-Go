package notificationservice

import (
	types "notification-service/types"
)

// implements IObserver interface
type NotificationEngine struct {
	notificationObservable *NotificationObservable
	strategies             []INotificationStrategy
}

func NewNotificationEngine() *NotificationEngine {
	observable := GetNotificationServiceInstance().GetObservable()
	newEngine := &NotificationEngine{
		notificationObservable: observable,
	}
	observable.AddObserver(newEngine)
	return newEngine
}

func (e *NotificationEngine) Update(notification types.INotification) {
	for _, strategy := range e.strategies {
		strategy.SendNotification(notification.GetContent())
	}
}

func (e *NotificationEngine) AddStrategy(strategy INotificationStrategy) {
	e.strategies = append(e.strategies, strategy)
}
