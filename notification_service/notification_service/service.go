package notificationservice

import (
	"notification-service/repository"
	types "notification-service/types"
	"sync"
)

type NotificationService struct {
	observable *NotificationObservable
}

var (
	once     sync.Once
	instance *NotificationService
)

func GetNotificationServiceInstance() *NotificationService {
	repo := repository.NewNotificationObservableRepository()
	newObservable := NewNotificationObservable(repo)
	once.Do(func() {
		instance = &NotificationService{
			observable: newObservable,
		}
	})
	return instance
}

func (s *NotificationService) GetObservable() *NotificationObservable {
	return s.observable
}

func (s *NotificationService) SendNotification(notification types.INotification) {
	s.observable.SetNotification(notification)
}
