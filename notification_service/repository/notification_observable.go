package repository

import "notification-service/types"

type NotificationObservableReposiory struct {
	observers []types.IObserver
}

func NewNotificationObservableRepository() *NotificationObservableReposiory {
	return &NotificationObservableReposiory{observers: []types.IObserver{}}
}

func (r *NotificationObservableReposiory) Add(obs types.IObserver) {
	r.observers = append(r.observers, obs)
}

func (r *NotificationObservableReposiory) GetAllObservers() []types.IObserver {
	return r.observers
}
