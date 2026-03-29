package main

import (
	"fmt"
	notificationservice "notification-service/notification_service"
	types "notification-service/types"
	"time"
)

func main() {
	service := notificationservice.GetNotificationServiceInstance()
	_ = notificationservice.NewLogger()
	notificationEngine := notificationservice.NewNotificationEngine()

	notificationEngine.AddStrategy(notificationservice.NewEmailStrategy("abc@gmail.com"))
	notificationEngine.AddStrategy(notificationservice.NewSMSStrategy("9163167900"))

	notification1 := types.NewSimpleNotification("Worker notification")
	scheduleTime := time.Now().Add(3 * time.Second)
	service.ScheduleNotification(notification1, scheduleTime)

	notification2 := types.NewSimpleNotification("Instant notification")
	service.SendNotification(notification2)
	time.Sleep(5 * time.Second)

	history := service.GetHistory()
	for _, h := range history {
		fmt.Println(h)
	}
}
