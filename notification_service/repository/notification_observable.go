package repository

import "notification-service/types"

/*
TODO: This NotificationObservableReposiory should be an interface which would be imported by
the client

The concretions can be: InMemoryStorage and DB
*/
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
