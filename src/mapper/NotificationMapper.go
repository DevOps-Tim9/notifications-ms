package mapper

import (
	"notifications-ms/src/dto"
	"notifications-ms/src/model"
)

func NotificationDTOToNotification(notificationDto *dto.NotificationDTO) *model.Notification {
	var notification model.Notification

	notification.Message = notificationDto.Message
	notification.UserAuth0ID = notificationDto.UserAuth0ID
	notification.NotificationType = notificationDto.NotificationType

	return &notification
}

func NotificationToNotificationDTO(notification *model.Notification) *dto.NotificationDTO {
	var notificationDto dto.NotificationDTO

	notificationDto.Message = notification.Message
	notificationDto.UserAuth0ID = notification.UserAuth0ID
	notificationDto.NotificationType = notification.NotificationType

	return &notificationDto
}
