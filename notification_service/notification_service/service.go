package notificationservice

import (
	"notification-service/repository"
	types "notification-service/types"
	"sync"
	"time"
)

type NotificationService struct {
	observable *NotificationObservable
	scheduler  *Scheduler
}

var (
	once     sync.Once
	instance *NotificationService
)

func GetNotificationServiceInstance() *NotificationService {
	repo := repository.NewNotificationObservableRepository()
	newObservable := NewNotificationObservable(repo)
	newScheduler := NewScheduler()
	once.Do(func() {
		instance = &NotificationService{
			observable: newObservable,
			scheduler:  newScheduler,
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

func (s *NotificationService) ScheduleNotification(notification types.INotification, executeAt time.Time) {
	s.scheduler.Schedule(executeAt, func() {
		s.SendNotification(notification)
	})
}
