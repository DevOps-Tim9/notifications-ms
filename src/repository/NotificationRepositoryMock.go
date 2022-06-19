package repository

import (
	"notifications-ms/src/model"

	"github.com/stretchr/testify/mock"
)

type NotificationRepositoryMock struct {
	mock.Mock
}

func (n *NotificationRepositoryMock) AddNotification(notification *model.Notification) error {
	args := n.Called(notification)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(error)
}

func (n *NotificationRepositoryMock) DeleteNotificationsByUserAuth0ID(userAuth0ID string) {

}

func (n *NotificationRepositoryMock) GetNotificationsByUserAuth0ID(userAuth0ID string) []*model.Notification {
	args := n.Called(userAuth0ID)
	return args.Get(0).([]*model.Notification)
}
