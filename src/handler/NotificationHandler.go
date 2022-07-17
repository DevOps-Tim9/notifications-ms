package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"notifications-ms/src/dto"
	"notifications-ms/src/service"
	"notifications-ms/src/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type NotificationHandler struct {
	Service *service.NotificationService
	Logger  *logrus.Entry
}

func extractClaims(tokenStr string) (jwt.MapClaims, bool) {
	token, _ := jwt.Parse(strings.Split(tokenStr, " ")[1], nil)

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, true
	} else {
		fmt.Println("Invalid JWT Token")
		return nil, false
	}
}

func getId(idParam string) (int, error) {
	id, err := strconv.ParseInt(idParam, 10, 32)
	if err != nil {
		return 0, errors.New("ID should be a number")
	}
	return int(id), nil
}

func (handler *NotificationHandler) GetNotifications(ctx *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx.Request.Context(), "GET /notifications")
	defer span.Finish()

	claims, _ := extractClaims(ctx.Request.Header.Get("Authorization"))

	handler.Logger.Info(fmt.Sprintf("Getting notifications for user %s", fmt.Sprint(claims["sub"])))
	notifications := handler.Service.GetNotifications(fmt.Sprint(claims["sub"]))

	ctx.JSON(http.StatusOK, notifications)
}

func (handler *NotificationHandler) DeleteNotifications(ctx *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx.Request.Context(), "DELETE /notifications")
	defer span.Finish()

	claims, _ := extractClaims(ctx.Request.Header.Get("Authorization"))

	handler.Logger.Info(fmt.Sprintf("Deleting notifications for user %s", fmt.Sprint(claims["sub"])))
	handler.Service.DeleteNotifications(fmt.Sprint(claims["sub"]))

	AddSystemEvent(time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf("Notifications deleted for user %s", fmt.Sprint(claims["sub"])))
	ctx.JSON(http.StatusOK, nil)
}

func AddSystemEvent(time string, message string) error {
	logger := utils.Logger()
	event := dto.EventRequestDTO{
		Timestamp: time,
		Message:   message,
	}

	b, _ := json.Marshal(&event)
	endpoint := os.Getenv("EVENTS_MS")
	logger.Info("Sending system event to events-ms")
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))
	req.Header.Set("content-type", "application/json")

	_, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Debug("Error happened during sending system event")
		return err
	}

	return nil
}
