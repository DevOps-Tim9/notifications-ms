package service

import (
	"fmt"
	"notifications-ms/src/dto"
	"notifications-ms/src/mapper"
	"notifications-ms/src/repository"
)

type NotificationService struct {
	NotificationRepo repository.INotificationRepository
}

type INotificationService interface {
	AddNotification(*dto.NotificationDTO) error
	GetNotifications(userAuth0ID string) []*dto.NotificationDTO
	DeleteNotifications(userAuth0ID string)
}

func NewNotificationService(notificationRepository repository.INotificationRepository) INotificationService {
	return &NotificationService{
		notificationRepository,
	}
}

func (service *NotificationService) AddNotification(notificationDto *dto.NotificationDTO) error {
	notification := mapper.NotificationDTOToNotification(notificationDto)

	err := notification.Validate()
	if err != nil {
		return err
	}

	errr := service.NotificationRepo.AddNotification(notification)
	if errr != nil {
		fmt.Println(errr)
		return errr
	}

	return nil
}

func (service *NotificationService) GetNotifications(userAuth0ID string) []*dto.NotificationDTO {
	notifications := service.NotificationRepo.GetNotificationsByUserAuth0ID(userAuth0ID)

	res := make([]*dto.NotificationDTO, len(notifications))
	for i := 0; i < len(notifications); i++ {
		res[i] = mapper.NotificationToNotificationDTO(notifications[i])
	}

	return res
}

func (service *NotificationService) DeleteNotifications(userAuth0ID string) {
	service.NotificationRepo.DeleteNotificationsByUserAuth0ID(userAuth0ID)
}
