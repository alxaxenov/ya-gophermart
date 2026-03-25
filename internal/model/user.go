package model

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Credentials struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (c Credentials) Validate() error {
	validate := validator.New()
	err := validate.Struct(c)
	if err == nil {
		return nil
	}
	return fmt.Errorf("ошибка валидации %s", err.Error())
}
