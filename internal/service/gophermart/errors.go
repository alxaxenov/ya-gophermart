package gophermart

import "errors"

var BalanceNotFoundError = errors.New("баланс пользователя не найден")

var OrderNumberIncorrectError = errors.New("некорректный номер заказа")

var InsufficientFundsError = errors.New("на счету недостаточно средств")

var AnotherUserError = errors.New("номер заказа уже был загружен другим пользователем")
