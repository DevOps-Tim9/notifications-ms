package dto

import "notifications-ms/src/model"

type NotificationDTO struct {
	Message          string
	UserAuth0ID      string
	NotificationType *model.NotificationType
}
