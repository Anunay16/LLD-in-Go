package types

type INotification interface {
	GetContent() string
}

// ============ Simple Notification =================

type SimpleNotification struct {
	content string
}

func (s *SimpleNotification) GetContent() string {
	return s.content
}

func NewSimpleNotification(content string) *SimpleNotification {
	return &SimpleNotification{content: content}
}

// =========== Observer Design Pattern ================

type IObservable interface {
	AddObserver(ob IObserver)
	RemoveObserver(ob IObserver)
	Notify()
}

type IObserver interface {
	Update(notification INotification)
}
