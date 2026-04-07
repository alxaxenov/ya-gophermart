package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_validate(t *testing.T) {
	type fields struct {
		DBURI      string
		AcrualAddr string
	}
	tests := []struct {
		name   string
		fields fields
		expErr error
	}{
		{
			name:   "without DBURI",
			fields: fields{"", "AcrualAddr"},
			expErr: fmt.Errorf("не передана строка подключения к базе данных. Используйте флаг -d или переменную DATABASE_URI"),
		},
		{
			name:   "without AcrualAddr",
			fields: fields{"DBURI", ""},
			expErr: errors.New("не передан адрес системы расчёта начислений. Используйте флаг -r или переменную ACCRUAL_SYSTEM_ADDRESS"),
		},
		{
			name:   "Ok",
			fields: fields{"DBURI", "AcrualAddr"},
			expErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				DBURI:      tt.fields.DBURI,
				AcrualAddr: tt.fields.AcrualAddr,
			}
			err := c.validate()
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
