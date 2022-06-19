package main

import (
	"fmt"
	"net/http"
	"notifications-ms/src/handler"
	"notifications-ms/src/model"
	"notifications-ms/src/rabbitmq"
	"notifications-ms/src/repository"
	"notifications-ms/src/service"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rs/cors"
)

var db *gorm.DB
var err error

func initDB() (*gorm.DB, error) {
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
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
	db, _ = gorm.Open("postgres", connectionString)

	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(model.Notification{})
	return db, err
}

func initNotificationRepo(database *gorm.DB) *repository.NotificationRepository {
	return &repository.NotificationRepository{Database: database}
}

func initNotificationService(notificationRepo *repository.NotificationRepository) *service.NotificationService {
	return &service.NotificationService{NotificationRepo: notificationRepo}
}

func initNotificationHandler(service *service.NotificationService) *handler.NotificationHandler {
	return &handler.NotificationHandler{Service: service}
}

func handleNotificationFunc(handler *handler.NotificationHandler, router *gin.Engine) {
	router.GET("/notifications", handler.GetNotifications)
	router.DELETE("/notifications", handler.DeleteNotifications)
}

func main() {
	database, _ := initDB()

	port := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))

	notificationRepo := initNotificationRepo(database)
	notificationService := initNotificationService(notificationRepo)
	notificationHandler := initNotificationHandler(notificationService)

	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	rabbit := rabbitmq.RMQConsumer{
		ConnectionString:    amqpServerURL,
		NotificationService: notificationService,
	}

	channel, _ := rabbit.StartRabbitMQ()

	defer channel.Close()

	messages, _ := channel.Consume(
		"AddNotification-MS",          // queue name
		"AddNotification-MS-consumer", // consumer
		true,                          // auto-ack
		false,                         // exclusive
		false,                         // no local
		false,                         // no wait
		nil,                           // arguments
	)

	go rabbit.Worker(messages)

	router := gin.Default()

	handleNotificationFunc(notificationHandler, router)

	http.ListenAndServe(port, cors.AllowAll().Handler(router))
}