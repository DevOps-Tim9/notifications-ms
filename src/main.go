package main

import (
	"fmt"
	"io"
	"net/http"
	"notifications-ms/src/handler"
	"notifications-ms/src/model"
	"notifications-ms/src/rabbitmq"
	"notifications-ms/src/repository"
	"notifications-ms/src/service"
	"notifications-ms/src/utils"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

var db *gorm.DB
var err error

func initDB() (*gorm.DB, error) {
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

func InitJaeger() (opentracing.Tracer, io.Closer, error) {
	cfg := config.Configuration{
		ServiceName: "notifications-ms",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "jaeger:6831",
		},
	}

	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	return tracer, closer, err
}

func initNotificationRepo(database *gorm.DB) *repository.NotificationRepository {
	return &repository.NotificationRepository{Database: database}
}

func initNotificationService(notificationRepo *repository.NotificationRepository) *service.NotificationService {
	return &service.NotificationService{NotificationRepo: notificationRepo, Logger: utils.Logger()}
}

func initNotificationHandler(service *service.NotificationService) *handler.NotificationHandler {
	return &handler.NotificationHandler{Service: service, Logger: utils.Logger()}
}

func handleNotificationFunc(handler *handler.NotificationHandler, router *gin.Engine) {
	router.GET("/notifications", handler.GetNotifications)
	router.DELETE("/notifications", handler.DeleteNotifications)
}

var totalTrafficSizeInGB = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "http_requests_total_traffic_size_in_gb",
		Help: "Total traffic size in GB.",
	},
)

var total404Requests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total_404",
		Help: "Total number of 404 requests.",
	},
	[]string{"path"},
)

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

var uniqueClients = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "http_unique_clients",
	Help: "Number of unique clients.",
}, []string{"ip", "timestamp", "browser"})

func prometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		path := ctx.Request.RequestURI

		requestSize := ctx.Request.ContentLength

		ip := ctx.ClientIP()
		browser := ctx.Request.UserAgent()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))

		ctx.Next()

		responseSize := ctx.Writer.Size()

		responseStatus.WithLabelValues(strconv.Itoa(ctx.Writer.Status())).Inc()
		totalRequests.WithLabelValues(path).Inc()
		uniqueClients.WithLabelValues(ip, time.Now().Format(time.UnixDate), browser).Inc()

		if responseSize < 0 {
			responseSize = 0
		}
		totalTrafficSizeInGB.Add((float64(requestSize + int64(responseSize))) / 1073741824)

		if ctx.Writer.Status() == 404 {
			total404Requests.WithLabelValues(path).Inc()
		}

		timer.ObserveDuration()
	})
}

func setupPrometherus() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
	prometheus.Register(total404Requests)
	prometheus.Register(totalTrafficSizeInGB)
}

func prometheusGin() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func main() {
	logger := utils.Logger()

	logger.Info("Connecting with DB")

	database, _ := initDB()

	port := fmt.Sprintf(":%s", os.Getenv("SERVER_PORT"))

	logger.Info("Initializing Jaeger")
	tracer, trCloser, err := InitJaeger()
	if err != nil {
		logger.Debug(err.Error())
		fmt.Printf("error init jaeger %v", err)
	} else {
		defer trCloser.Close()
		opentracing.SetGlobalTracer(tracer)
	}

	notificationRepo := initNotificationRepo(database)
	notificationService := initNotificationService(notificationRepo)
	notificationHandler := initNotificationHandler(notificationService)

	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	rabbit := rabbitmq.RMQConsumer{
		ConnectionString:    amqpServerURL,
		NotificationService: notificationService,
	}

	logger.Info("Starting RabbitMQ")
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

	setupPrometherus()

	router.Use(prometheusMiddleware())

	router.GET("/api/metrics", prometheusGin())

	handleNotificationFunc(notificationHandler, router)

	logger.Info(fmt.Sprintf("Starting server on port %s", os.Getenv("SERVER_PORT")))
	http.ListenAndServe(port, cors.AllowAll().Handler(router))
}
