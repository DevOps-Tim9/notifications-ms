package model

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator"
)

type NotificationType int

const (
	Message NotificationType = iota
	Follow
	Like
	Comment
)

type Notification struct {
	ID               int               `json:"id"`
	Message          string            `json:"message" validate:"required"`
	UserAuth0ID      string            `json:"userAuth0ID" validate:"required"`
	NotificationType *NotificationType `json:"notificationType" validate:"required"`
}

func (n *Notification) Validate() error {
	validate := validator.New()
	return validate.Struct(n)
}

func (n *Notification) FromJSON(r io.Reader) error {
	e := json.NewDecoder(r)
	return e.Decode(n)
}

func (n *Notification) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(n)
}
