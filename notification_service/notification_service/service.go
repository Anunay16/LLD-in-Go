package notificationservice

import (
	"notification-service/repository"
	types "notification-service/types"
	"sync"
	"time"

	"github.com/google/uuid"
)

type NotificationService struct {
	observable  *NotificationObservable
	scheduler   *Scheduler
	historyRepo repository.NotificationHistoryRepository
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
			observable:  newObservable,
			scheduler:   newScheduler,
			historyRepo: repository.NewInMemoryHistoryRepository(),
		}
	})
	return instance
}

func (s *NotificationService) GetObservable() *NotificationObservable {
	return s.observable
}

func (s *NotificationService) SendNotification(notification types.INotification) {
	s.observable.SetNotification(notification)

	now := time.Now()

	history := types.NotificationHistory{
		Id:           uuid.NewString(),
		Notification: notification,
		Status:       types.Sent,
		SentAt:       &now,
	}

	s.historyRepo.Save(history)
}

func (s *NotificationService) ScheduleNotification(notification types.INotification, executeAt time.Time) {

	history := types.NotificationHistory{
		Id:           uuid.NewString(),
		Notification: notification,
		Status:       types.Scheduled,
		CreatedAt:    time.Now(),
	}

	s.historyRepo.Save(history)

	s.scheduler.Schedule(executeAt, func() {
		s.SendNotification(notification)
	})
}

func (s *NotificationService) GetHistory() []types.NotificationHistory {
	return s.historyRepo.GetAll()
}
