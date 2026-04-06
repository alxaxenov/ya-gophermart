package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/config/db"
	"github.com/alxaxenov/ya-gophermart/internal/domain/order"
	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sqlx.DB
var c *db.ConnectorPG
var testContainer testcontainers.Container

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(60 * time.Second). // Увеличьте таймаут для CI
			WithPollInterval(2 * time.Second),    // Установите интервал опроса
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatal(err)
	}
	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	testContainer = container

	c = db.NewPGConnector(connStr)
	testDB, err = c.Open()
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	c.Close()
	if err := testContainer.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate container: %v", err)
	}
	os.Exit(code)
}

func truncateTables(db *sqlx.DB) error {
	tables := []string{"orders", "balance"}
	for _, t := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + t + " RESTART IDENTITY CASCADE")
		if err != nil {
			return err
		}
	}
	return nil
}

func TestGophermartRepo_GetOrderByNumber(t *testing.T) {
	tests := []struct {
		name     string
		orderNum string
		fillDB   func() error
		want     *model.Order
		expErr   func(*testing.T, error)
	}{
		{
			name:     "success",
			orderNum: "12345",
			fillDB: func() error {
				_, err := testDB.Exec("INSERT INTO orders (number, user_id) VALUES ('12345', 100)")
				return err
			},
			want:   &model.Order{1, 100},
			expErr: nil,
		},
		{
			name:     "not found",
			orderNum: "12345",
			fillDB: func() error {
				return nil
			},
			want: nil,
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "GophermartRepo GetOrderByNumber select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			err = tt.fillDB()
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			got, err := g.GetOrderByNumber(context.Background(), tt.orderNum)
			if err != nil {
				assert.Error(t, err, tt.name)
				tt.expErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equalf(t, tt.want, got, "GetOrderByNumber(%v)", tt.orderNum)
		})
	}
}

func TestGophermartRepo_CreateOrder(t *testing.T) {
	tests := []struct {
		name     string
		want     bool
		orderNum string
	}{
		{
			name:     "success",
			want:     true,
			orderNum: "123456",
		},
		{
			name:     "success",
			want:     false,
			orderNum: "12345",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			_, err = testDB.Exec("INSERT INTO orders (number, user_id) VALUES ('12345', 100)")
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			got, err := g.CreateOrder(context.Background(), 1, tt.orderNum)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "CreateOrder(%v, %v, %v)", tt.orderNum)
		})
	}
}

func TestGophermartRepo_UpdateOrderStatus(t *testing.T) {
	tests := []struct {
		name      string
		orderNum  string
		status    order.Status
		expStatus order.Status
	}{
		{
			name:      "to PROCESSING",
			orderNum:  "12345",
			status:    order.PROCESSING,
			expStatus: order.PROCESSING,
		},
		{
			name:      "to INVALID",
			orderNum:  "12345",
			status:    order.INVALID,
			expStatus: order.INVALID,
		},
		{
			name:      "to PROCESSED",
			orderNum:  "12345",
			status:    order.PROCESSED,
			expStatus: order.PROCESSED,
		},
		{
			name:      "another order",
			orderNum:  "123456",
			status:    order.PROCESSED,
			expStatus: order.NEW,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			_, err = testDB.Exec("INSERT INTO orders (number, user_id) VALUES ('12345', 100)")
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			err = g.UpdateOrderStatus(context.Background(), tt.orderNum, tt.status)
			assert.NoError(t, err)

			o := struct {
				Status order.Status `db:"status"`
			}{}
			testDB.Get(&o, "SELECT status FROM orders WHERE number = $1", "12345")
			assert.Equal(t, tt.expStatus, o.Status)
		})
	}
}

func TestGophermartRepo_GetOrdersToProcess(t *testing.T) {
	tests := []struct {
		name   string
		fillDB func() error
		want   *[]model.OrderToProcess
	}{
		{
			name: "no orders in DB",
			fillDB: func() error {
				return nil
			},
			want: &[]model.OrderToProcess{},
		},
		{
			name: "no orders in progress",
			fillDB: func() error {
				execStrings := []string{
					"INSERT INTO orders (number, user_id, status) VALUES ('12345', 100, 'INVALID')",
					"INSERT INTO orders (number, user_id, status) VALUES ('67890', 200, 'PROCESSED')",
				}
				for _, execString := range execStrings {
					_, err := testDB.Exec(execString)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: &[]model.OrderToProcess{},
		},
		{
			name: "ok",
			fillDB: func() error {
				execStrings := []string{
					"INSERT INTO orders (number, user_id, status) VALUES ('12345', 100, 'INVALID')",
					"INSERT INTO orders (number, user_id, status) VALUES ('67890', 200, 'PROCESSED')",
					"INSERT INTO orders (number, user_id, status) VALUES ('1', 1, 'NEW')",
					"INSERT INTO orders (number, user_id, status) VALUES ('2', 2, 'PROCESSING')",
				}
				for _, execString := range execStrings {
					_, err := testDB.Exec(execString)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: &[]model.OrderToProcess{
				{"1", order.NEW, 1},
				{"2", order.PROCESSING, 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			err = tt.fillDB()
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			got, err := g.GetOrdersToProcess(context.Background(), 10)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetOrdersToProcess() - %s", tt.name)
		})
	}
}

func TestGophermartRepo_GetOrders(t *testing.T) {
	tests := []struct {
		name   string
		fillDB func() error
		want   *[]model.OrderForList
	}{
		{
			name: "no orders in DB",
			fillDB: func() error {
				return nil
			},
			want: &[]model.OrderForList{},
		},
		{
			name: "no orders for user",
			fillDB: func() error {
				execStrings := []string{
					"INSERT INTO orders (number, user_id) VALUES ('12345', 100)",
				}
				for _, execString := range execStrings {
					_, err := testDB.Exec(execString)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: &[]model.OrderForList{},
		},
		{
			name: "ok",
			fillDB: func() error {
				execStrings := []string{
					"INSERT INTO orders (number, user_id) VALUES ('12345', 100)",
					"INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ('1', 10, 'PROCESSED', 25.55, '2024-12-07 14:30:00+00')",
					"INSERT INTO orders (number, user_id, status, accrual, uploaded_at) VALUES ('2', 10, 'NEW', NULL, '2026-05-17 08:00:30+00')",
				}
				for _, execString := range execStrings {
					_, err := testDB.Exec(execString)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: &[]model.OrderForList{
				//time.Date(2024, 12, 7, 14, 30, 5, 123456000, time.UTC)
				{"2", order.NEW, sql.NullFloat64{0, false}, time.Date(2026, time.May, 17, 8, 00, 30, 0, time.UTC)},
				{"1", order.PROCESSED, sql.NullFloat64{25.55, true}, time.Date(2024, time.December, 07, 14, 30, 0, 0, time.UTC)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			err = tt.fillDB()
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			got, err := g.GetOrders(context.Background(), 10)
			for i := range *got {
				(*got)[i].UploadedAt = (*got)[i].UploadedAt.UTC()
			}
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetOrders() - %s", tt.name)
		})
	}
}

func TestGophermartRepo_GetBalance(t *testing.T) {
	tests := []struct {
		name   string
		fillDB func() error
		want   *model.Balance
	}{
		{
			name: "no orders in DB",
			fillDB: func() error {
				return nil
			},
			want: nil,
		},
		{
			name: "ok",
			fillDB: func() error {
				execStrings := []string{
					"INSERT INTO balance (user_id, current, withdrawn) VALUES (10, 100, 200)",
					"INSERT INTO balance (user_id, current, withdrawn) VALUES (20, 0, 1000)",
				}
				for _, execString := range execStrings {
					_, err := testDB.Exec(execString)
					if err != nil {
						return err
					}
				}
				return nil
			},
			want: &model.Balance{1, 10, 100, 200},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := truncateTables(testDB)
			assert.NoError(t, err)
			err = tt.fillDB()
			assert.NoError(t, err)

			g := &GophermartRepo{Connector: c}
			got, err := g.GetBalance(context.Background(), 10)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, got, "GetBalance() - %s", tt.name)
		})
	}
}
