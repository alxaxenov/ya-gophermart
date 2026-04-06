package repo

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	"errors"

	dbMock "github.com/alxaxenov/ya-gophermart/internal/config/db/mocks"
	"github.com/alxaxenov/ya-gophermart/internal/repo/mocks"
	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserRepo_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		want       *model.User
		setupMocks func(c *mocks.Mockconnector, db *dbMock.MockDB)
		expErr     func(*testing.T, error)
	}{
		{
			name: "ErrNoRows",
			want: nil,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(sql.ErrNoRows)
			},
			expErr: nil,
		},
		{
			name: "Another error",
			want: nil,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo GetUser select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Ok",
			want: &model.User{
				ID:           10,
				Login:        "test_login",
				PasswordHash: "test_password_hash",
				Active:       true,
			},
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(
					gomock.Any(),
					&model.User{},
					"SELECT id, login, password_hash, active FROM users WHERE login = $1",
					"test_login",
				).Times(1).DoAndReturn(
					func(ctx context.Context, user *model.User, login, password string) error {
						user.ID = 10
						user.Login = "test_login"
						user.PasswordHash = "test_password_hash"
						user.Active = true
						return nil
					})
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo GetUser select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			con := mocks.NewMockconnector(ctrl)
			db := dbMock.NewMockDB(ctrl)
			tt.setupMocks(con, db)

			d := &UserRepo{con}
			got, err := d.GetUser(context.Background(), "test_login")
			if err != nil {
				assert.Error(t, err, tt.name)
				tt.expErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRepo_CheckLogin(t *testing.T) {
	tests := []struct {
		name       string
		want       bool
		setupMocks func(c *mocks.Mockconnector, db *dbMock.MockDB)
		expErr     func(*testing.T, error)
	}{
		{
			name: "ErrNoRows",
			want: false,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(sql.ErrNoRows)
			},
			expErr: nil,
		},
		{
			name: "Another error",
			want: false,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CheckLogin select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Ok not zero",
			want: true,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(
					gomock.Any(),
					gomock.AssignableToTypeOf(new(int)),
					"SELECT id FROM users WHERE login = $1",
					"test_login",
				).Times(1).DoAndReturn(func(ctx context.Context, user *int, login, password string) error {
					*user = 10
					return nil
				})
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CheckLogin select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Ok zero",
			want: false,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().GetContext(
					gomock.Any(),
					gomock.AssignableToTypeOf(new(int)),
					"SELECT id FROM users WHERE login = $1",
					"test_login",
				).Times(1).DoAndReturn(func(ctx context.Context, user *int, login, password string) error {
					*user = 0
					return nil
				})
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CheckLogin select error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			con := mocks.NewMockconnector(ctrl)
			db := dbMock.NewMockDB(ctrl)
			tt.setupMocks(con, db)

			d := &UserRepo{con}
			got, err := d.CheckLogin(context.Background(), "test_login")
			if err != nil {
				assert.Error(t, err, tt.name)
				tt.expErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRepo_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		want       int
		setupMocks func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx)
		expErr     func(*testing.T, error)
	}{
		{
			name: "BeginTxx error",
			want: 0,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CreateUser begin tx error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "GetContext error",
			want: 0,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback().Times(1)
				tx.EXPECT().GetContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CreateUser user query error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "ExecContext error",
			want: 0,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback().Times(1)
				tx.EXPECT().GetContext(
					gomock.Any(),
					gomock.AssignableToTypeOf(new(int)),
					"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
					"test_login",
					"test_password",
				).DoAndReturn(func(ctx context.Context, user *int, queryUser string, args ...string) error {
					*user = 10
					return nil
				})
				tx.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CreateUser balance query error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Commit error",
			want: 0,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback().Times(1)
				tx.EXPECT().GetContext(
					gomock.Any(),
					gomock.AssignableToTypeOf(new(int)),
					"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
					"test_login",
					"test_password",
				).DoAndReturn(func(ctx context.Context, user *int, queryUser string, args ...string) error {
					*user = 10
					return nil
				})
				tx.EXPECT().ExecContext(
					gomock.Any(),
					"INSERT INTO balance (user_id) VALUES ($1)",
					10,
				).Return(nil, nil)
				tx.EXPECT().Commit().Return(errors.New("test error"))
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CreateUser commit error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
		{
			name: "Ok",
			want: 10,
			setupMocks: func(c *mocks.Mockconnector, db *dbMock.MockDB, tx *dbMock.MockTx) {
				c.EXPECT().GetDB().Times(1).Return(db)
				db.EXPECT().BeginTxx(gomock.Any(), gomock.Any()).Return(tx, nil)
				tx.EXPECT().Rollback().Times(1)
				tx.EXPECT().GetContext(
					gomock.Any(),
					gomock.AssignableToTypeOf(new(int)),
					"INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id",
					"test_login",
					"test_password",
				).DoAndReturn(func(ctx context.Context, user *int, queryUser string, args ...string) error {
					*user = 10
					return nil
				})
				tx.EXPECT().ExecContext(
					gomock.Any(),
					"INSERT INTO balance (user_id) VALUES ($1)",
					10,
				).Return(nil, nil)
				tx.EXPECT().Commit().Return(nil)
			},
			expErr: func(t *testing.T, err error) {
				expectedPrefix := "UserRepo CreateUser commit error:"
				if !strings.Contains(err.Error(), expectedPrefix) {
					t.Errorf("error message %q does not contain %q", err.Error(), expectedPrefix)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			con := mocks.NewMockconnector(ctrl)
			db := dbMock.NewMockDB(ctrl)
			tx := dbMock.NewMockTx(ctrl)
			tt.setupMocks(con, db, tx)

			d := &UserRepo{con}
			got, err := d.CreateUser(context.Background(), "test_login", "test_password")
			if err != nil {
				assert.Error(t, err, tt.name)
				tt.expErr(t, err)
			} else {
				assert.NoError(t, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
