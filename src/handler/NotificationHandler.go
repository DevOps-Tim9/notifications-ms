package handler

import (
	"errors"
	"fmt"
	"net/http"
	"notifications-ms/src/service"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

type NotificationHandler struct {
	Service *service.NotificationService
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

	notifications := handler.Service.GetNotifications(fmt.Sprint(claims["sub"]))

	ctx.JSON(http.StatusOK, notifications)
}

func (handler *NotificationHandler) DeleteNotifications(ctx *gin.Context) {
	span, _ := opentracing.StartSpanFromContext(ctx.Request.Context(), "DELETE /notifications")
	defer span.Finish()

	claims, _ := extractClaims(ctx.Request.Header.Get("Authorization"))

	handler.Service.DeleteNotifications(fmt.Sprint(claims["sub"]))

	ctx.JSON(http.StatusOK, nil)
}
