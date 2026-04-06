package utils

import "strconv"

// OrderNumberCheck проверка номера заказа по алгоритму Луна
func OrderNumberCheck(n string) bool {
	if len(n) == 0 {
		return false
	}
	var sum int
	l := len(n)
	for i := l - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(n[i]))
		if err != nil {
			return false
		}
		if (l-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
