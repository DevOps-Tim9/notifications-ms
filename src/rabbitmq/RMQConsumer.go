package rabbitmq

import (
	"encoding/json"
	"fmt"
	"notifications-ms/src/dto"
	"notifications-ms/src/handler"
	"notifications-ms/src/service"
	"time"

	"github.com/streadway/amqp"
)

type RMQConsumer struct {
	ConnectionString    string
	NotificationService service.INotificationService
}

func (r RMQConsumer) StartRabbitMQ() (*amqp.Channel, error) {
	connectRabbitMQ, errr := amqp.Dial(r.ConnectionString)

	if errr != nil {
		return nil, errr
	}

	channelRabbitMQ, _ := connectRabbitMQ.Channel()

	err := channelRabbitMQ.ExchangeDeclare(
		"AddNotification-MS-exchange",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	queue, err := channelRabbitMQ.QueueDeclare(
		"AddNotification-MS",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = channelRabbitMQ.QueueBind(
		queue.Name,
		"AddNotification-MS-routing-key",
		"AddNotification-MS-exchange",
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	err = channelRabbitMQ.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)

	if err != nil {
		return nil, err
	}

	return channelRabbitMQ, nil
}

func (r RMQConsumer) HandleAddNotification(message []byte) {
	var notificationDto dto.NotificationDTO

	json.Unmarshal([]byte(message), &notificationDto)

	r.NotificationService.AddNotification(&notificationDto)
	handler.AddSystemEvent(time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf("New notification added for user %s, type: %d", notificationDto.UserAuth0ID, notificationDto.NotificationType))
}

func (r RMQConsumer) Worker(messages <-chan amqp.Delivery) {
	for delivery := range messages {
		r.HandleAddNotification(delivery.Body)
	}
}
