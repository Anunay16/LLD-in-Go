package main

import (
	types "notification-service/types"
	notificationservice "notification-service/notification_service"
)

func main() {
	service := notificationservice.GetNotificationServiceInstance()
	_ = notificationservice.NewLogger()
	notificationEngine := notificationservice.NewNotificationEngine()

	notificationEngine.AddStrategy(notificationservice.NewEmailStrategy("abc@gmail.com"))
	notificationEngine.AddStrategy(notificationservice.NewSMSStrategy("9163167900"))

	notification := types.NewSimpleNotification("Your order has been shipped")

	service.SendNotification(notification)
}
