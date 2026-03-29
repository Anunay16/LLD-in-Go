package notificationservice

import (
	"notification-service/repository"
	"notification-service/types"
)

// concrete observable, implements IObservable interface
type NotificationObservable struct {
	currentNotification    types.INotification
	notificationRepository *repository.NotificationObservableReposiory
}

func NewNotificationObservable(repo *repository.NotificationObservableReposiory) *NotificationObservable {
	return &NotificationObservable{
		currentNotification:    nil,
		notificationRepository: repo,
	}
}

func (n *NotificationObservable) AddObserver(ob types.IObserver) {
	n.notificationRepository.Add(ob)
}

func (n *NotificationObservable) RemoveObserver(ob types.IObserver) {

}

func (n *NotificationObservable) Notify() {
	observers := n.notificationRepository.GetAllObservers()
	for _, observer := range observers {
		observer.Update(n.currentNotification)
	}
}

func (n *NotificationObservable) SetNotification(notification types.INotification) {
	n.currentNotification = notification
	n.Notify()
}

func (n *NotificationObservable) GetNotification() types.INotification {
	return n.currentNotification
}

func (n *NotificationObservable) GetNotificationContent() string {
	return n.currentNotification.GetContent()
}
