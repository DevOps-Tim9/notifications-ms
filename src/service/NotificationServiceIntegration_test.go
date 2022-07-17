package service

import (
	"fmt"
	"notifications-ms/src/dto"
	"notifications-ms/src/model"
	"notifications-ms/src/repository"
	"notifications-ms/src/utils"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type NotificationServiceIntegrationTestSuite struct {
	suite.Suite
	service       NotificationService
	db            *gorm.DB
	notifications []model.Notification
}

func (suite *NotificationServiceIntegrationTestSuite) SetupSuite() {
	host := os.Getenv("DATABASE_DOMAIN")
	user := os.Getenv("DATABASE_USERNAME")
	password := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_SCHEMA")
	port := os.Getenv("DATABASE_PORT")

	connectionString := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host,
		user,
		password,
		name,
		port,
	)
	db, _ := gorm.Open("postgres", connectionString)

	db.AutoMigrate(model.Notification{})

	db.Where("1=1").Delete(model.Notification{})

	notificationRepository := repository.NotificationRepository{Database: db}

	suite.db = db

	suite.service = NotificationService{
		NotificationRepo: &notificationRepository,
		Logger:           utils.Logger(),
	}

	ntype := model.Like
	ntypeFollow := model.Follow
	suite.notifications = []model.Notification{
		{
			Message:          "Test message",
			UserAuth0ID:      "auth0Id1",
			NotificationType: &ntype,
		},
		{
			Message:          "Test message2",
			UserAuth0ID:      "auth0Id2",
			NotificationType: &ntypeFollow,
		},
		{
			Message:          "Test message3",
			UserAuth0ID:      "auth0Id1",
			NotificationType: &ntypeFollow,
		},
	}

	tx := suite.db.Begin()

	tx.Create(&suite.notifications[0])
	tx.Create(&suite.notifications[1])
	tx.Create(&suite.notifications[2])

	tx.Commit()
}

func TestNotificationServiceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationServiceIntegrationTestSuite))
}

func (suite *NotificationServiceIntegrationTestSuite) TestIntegrationNotificationService_AddNotification_Pass() {
	ntype := model.Comment
	notificationDTO := dto.NotificationDTO{
		UserAuth0ID:      "auth0",
		Message:          "Test message",
		NotificationType: &ntype,
	}

	err := suite.service.AddNotification(&notificationDTO)

	assert.Equal(suite.T(), nil, err)
}

func (suite *NotificationServiceIntegrationTestSuite) TestIntegrationNotificationService_AddNotification_RequiredFieldMissing() {
	ntype := model.Comment
	notificationDTO := dto.NotificationDTO{
		NotificationType: &ntype,
		Message:          "Test message",
	}

	err := suite.service.AddNotification(&notificationDTO)

	assert.NotNil(suite.T(), err)
}

func (suite *NotificationServiceIntegrationTestSuite) TestIntegrationNotificationService_DeleteNotifications_Pass() {
	userId := "auth0Id2|auth0"

	suite.service.DeleteNotifications(userId)

	assert.True(suite.T(), true)
}

func (suite *NotificationServiceIntegrationTestSuite) TestIntegrationNotificationService_GetNotifications_NoNotifications() {
	userId := "authId3"

	notifications := suite.service.GetNotifications(userId)

	assert.NotNil(suite.T(), notifications)
	assert.Equal(suite.T(), 0, len(notifications))
}

func (suite *NotificationServiceIntegrationTestSuite) TestIntegrationNotificationService_GetNotifications_NotificationsExist() {
	userId := "auth0Id1"

	notifications := suite.service.GetNotifications(userId)

	assert.NotNil(suite.T(), notifications)
	assert.Equal(suite.T(), 2, len(notifications))
}
