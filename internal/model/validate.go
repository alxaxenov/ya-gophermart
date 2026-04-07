package model

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Validate Валидирует значения полей структуры, параметры валидации задаются
func Validate(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errors.New("ошибка валидации")
	}

	validate := validator.New()
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	return fmt.Errorf("ошибка валидации %s", err.Error())
}
