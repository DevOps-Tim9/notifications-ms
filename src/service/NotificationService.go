package service

import (
	"fmt"
	"notifications-ms/src/dto"
	"notifications-ms/src/mapper"
	"notifications-ms/src/repository"

	"github.com/sirupsen/logrus"
)

type NotificationService struct {
	NotificationRepo repository.INotificationRepository
	Logger           *logrus.Entry
}

type INotificationService interface {
	AddNotification(*dto.NotificationDTO) error
	GetNotifications(userAuth0ID string) []*dto.NotificationDTO
	DeleteNotifications(userAuth0ID string)
}

func NewNotificationService(notificationRepository repository.INotificationRepository, logger *logrus.Entry) INotificationService {
	return &NotificationService{
		notificationRepository,
		logger,
	}
}

func (service *NotificationService) AddNotification(notificationDto *dto.NotificationDTO) error {
	notification := mapper.NotificationDTOToNotification(notificationDto)

	err := notification.Validate()
	if err != nil {
		service.Logger.Debug(err.Error())
		return err
	}
	service.Logger.Info(fmt.Sprintf("Adding notification for user %s", notification.UserAuth0ID))
	errr := service.NotificationRepo.AddNotification(notification)
	if errr != nil {
		service.Logger.Debug(errr.Error())
		return errr
	}

	service.Logger.Info(fmt.Sprintf("Successfully added notification for user %s", notification.UserAuth0ID))
	return nil
}

func (service *NotificationService) GetNotifications(userAuth0ID string) []*dto.NotificationDTO {
	service.Logger.Info(fmt.Sprintf("Getting notifications for user %s in database", userAuth0ID))
	notifications := service.NotificationRepo.GetNotificationsByUserAuth0ID(userAuth0ID)

	res := make([]*dto.NotificationDTO, len(notifications))
	for i := 0; i < len(notifications); i++ {
		res[i] = mapper.NotificationToNotificationDTO(notifications[i])
	}

	service.Logger.Info(fmt.Sprintf("Successfully got notifications for user %s", userAuth0ID))
	return res
}

func (service *NotificationService) DeleteNotifications(userAuth0ID string) {
	service.NotificationRepo.DeleteNotificationsByUserAuth0ID(userAuth0ID)
	service.Logger.Info(fmt.Sprintf("Successfully deleted notifications for user %s", userAuth0ID))
}
