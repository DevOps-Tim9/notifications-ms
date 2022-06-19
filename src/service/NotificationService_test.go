package service

import (
	"notifications-ms/src/dto"
	"notifications-ms/src/model"
	"notifications-ms/src/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type NotificationServiceUnitTestsSuite struct {
	suite.Suite
	notificationRepositoryMock *repository.NotificationRepositoryMock
	service                    INotificationService
}

func TestNotificationServiceUnitTestsSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceUnitTestsSuite))
}

func (suite *NotificationServiceUnitTestsSuite) SetupSuite() {
	suite.notificationRepositoryMock = new(repository.NotificationRepositoryMock)
	suite.service = NewNotificationService(suite.notificationRepositoryMock)
}

func (suite *NotificationServiceUnitTestsSuite) TestNewNotificationService() {
	assert.NotNil(suite.T(), suite.service, "Service is nil")
}

func (suite *NotificationServiceUnitTestsSuite) TestNotificationService_AddNotification_NotificationWithoutMessageShouldFail() {
	ntype := model.Comment
	notificationDTO := dto.NotificationDTO{
		UserAuth0ID:      "auth0",
		NotificationType: &ntype,
	}

	err := suite.service.AddNotification(&notificationDTO)

	assert.NotEqual(suite.T(), nil, err)
}

func (suite *NotificationServiceUnitTestsSuite) TestNotificationService_AddNotification_ValidNotificationProvided() {
	ntype := model.Like
	message := "Someone liked your post"
	auth0Id := "auth0id"
	notificationDTO := dto.NotificationDTO{
		Message:          message,
		UserAuth0ID:      auth0Id,
		NotificationType: &ntype,
	}

	notificationEntity := model.Notification{
		Message:          message,
		UserAuth0ID:      auth0Id,
		NotificationType: &ntype,
	}

	suite.notificationRepositoryMock.On("AddNotification", &notificationEntity).Return(nil)

	err := suite.service.AddNotification(&notificationDTO)

	assert.Equal(suite.T(), nil, err)
}

func (suite *NotificationServiceUnitTestsSuite) TestNotificationService_GetNotifications_NoNotificationsForUser() {
	auth0Id := "auth0id"
	suite.notificationRepositoryMock.On("GetNotificationsByUserAuth0ID", auth0Id).Return([]*model.Notification{}).Once()

	notifications := suite.service.GetNotifications(auth0Id)

	assert.Equal(suite.T(), 0, len(notifications))
}

func (suite *NotificationServiceUnitTestsSuite) TestNotificationService_GetNotifications_ReturnedListOfNotifications() {
	ntype := model.Like
	message := "Someone liked your post"
	auth0Id := "auth0id"

	notificationEntity := model.Notification{
		Message:          message,
		UserAuth0ID:      auth0Id,
		NotificationType: &ntype,
	}
	var list []*model.Notification
	list = append(list, &notificationEntity)

	suite.notificationRepositoryMock.On("GetNotificationsByUserAuth0ID", auth0Id).Return(list).Once()

	notifications := suite.service.GetNotifications(auth0Id)

	assert.Equal(suite.T(), len(list), len(notifications))
	for i := 0; i < len(notifications); i++ {
		assert.Equal(suite.T(), list[i].Message, notifications[i].Message)
		assert.Equal(suite.T(), list[i].UserAuth0ID, notifications[i].UserAuth0ID)
		assert.Equal(suite.T(), list[i].NotificationType, notifications[i].NotificationType)
	}
}
