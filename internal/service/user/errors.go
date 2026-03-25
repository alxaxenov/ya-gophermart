package user

import (
	"errors"
	"fmt"
)

type AlreadyExistsError struct {
	Login string
	Err   error
}

func (l *AlreadyExistsError) Error() string {
	return fmt.Sprintf("Пользователь %s уже существует", l.Login)
}

func (l *AlreadyExistsError) Unwrap() error {
	return l.Err
}

func NewAlreadyExists(l string, err error) *AlreadyExistsError {
	return &AlreadyExistsError{Login: l, Err: err}
}

var IncorrectCredentialsError = errors.New("некорректный логин или пароль")

type InactiveError struct {
	Login string
	Err   error
}

func (i *InactiveError) Error() string {
	return fmt.Sprintf("Пользователь %s неактивен", i.Login)
}

func (i *InactiveError) Unwrap() error {
	return i.Err
}

func NewInactive(l string, err error) *InactiveError {
	return &InactiveError{Login: l, Err: err}
}

type NotFoundError struct {
	Login string
	Err   error
}

func (n *NotFoundError) Error() string {
	return fmt.Sprintf("Пользователь %s не найден", n.Login)
}

func (n *NotFoundError) Unwrap() error {
	return n.Err
}

func NewNotFound(l string, err error) *NotFoundError {
	return &NotFoundError{Login: l, Err: err}
}
