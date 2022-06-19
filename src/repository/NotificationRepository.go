package repository

import (
	"fmt"
	"notifications-ms/src/model"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type INotificationRepository interface {
	AddNotification(*model.Notification) error
	GetNotificationsByUserAuth0ID(userAuth0ID string) []*model.Notification
	DeleteNotificationsByUserAuth0ID(userAuth0ID string)
}

func NewNotificationRepository(database *gorm.DB) INotificationRepository {
	return &NotificationRepository{
		database,
	}
}

type NotificationRepository struct {
	Database *gorm.DB
}

func (repo *NotificationRepository) AddNotification(notification *model.Notification) error {
	result := repo.Database.Create(notification)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (repo *NotificationRepository) GetNotificationsByUserAuth0ID(userAuth0ID string) []*model.Notification {
	var notifications = []*model.Notification{}
	if result := repo.Database.Find(&notifications, "user_auth0_id = ?", userAuth0ID); result.Error != nil {
		return nil
	}

	return notifications
}

func (repo *NotificationRepository) DeleteNotificationsByUserAuth0ID(userAuth0ID string) {
	param := "'%" + strings.Split(userAuth0ID, "|")[1] + "'"
	repo.Database.Exec(fmt.Sprintf("delete from notifications where user_auth0_id LIKE %s", param))
}
