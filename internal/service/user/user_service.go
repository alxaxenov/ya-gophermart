package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alxaxenov/ya-gophermart/internal/repo/model"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secretKey     string
	cookieAuthKey string
	expireHours   int
	userRepo      expectedRepo
}

type expectedRepo interface {
	CheckLogin(context.Context, string) (bool, error)
	CreateUser(context.Context, string, string) (int, error)
	GetUser(context.Context, string) (*model.User, error)
}

type Hash interface {
	HashPass(string) (string, error)
	CheckPass(string, string) bool
}

func NewUserService(
	secretKey string, cookieAuthKey string, expireHours int, userRepo expectedRepo,
) *Service {
	return &Service{
		secretKey,
		cookieAuthKey,
		expireHours,
		userRepo,
	}
}

func (u *Service) RegisterUser(ctx context.Context, login string, password string) (int, error) {
	exist, err := u.userRepo.CheckLogin(ctx, login)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, NewAlreadyExists(login, nil)
	}
	hash, err := u.hashPass(password)
	if err != nil {
		return 0, err
	}
	userID, err := u.userRepo.CreateUser(ctx, login, hash)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (u *Service) LoginUser(ctx context.Context, login string, password string) (int, error) {
	user, err := u.userRepo.GetUser(ctx, login)
	if err != nil {
		return 0, err
	}
	if user == nil {
		return 0, NewNotFound(login, nil)
	}
	if !user.Active {
		return 0, NewInactive(login, nil)
	}
	if !u.checkPass(password, user.PasswordHash) {
		return 0, IncorrectCredentialsError
	}
	return user.ID, nil
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int `json:"user_id"`
}

func (u *Service) GetCookieAuthKey() string {
	return u.cookieAuthKey
}

func (u *Service) GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(u.secretKey), nil
		})
	if err != nil {
		return 0, fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return 0, errors.New("token is not valid")
	}
	return claims.UserID, nil
}

func (u *Service) BuildTokenString(id int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(u.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: id,
	})
	tokenString, err := token.SignedString([]byte(u.secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return tokenString, nil
}

func (u *Service) hashPass(pass string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("HashPass error: %w", err)
	}
	return string(hash), nil
}

func (u *Service) checkPass(pass string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}

type UserIDType string

const UserIDKey UserIDType = "userID"

func (u *Service) GetUserIDCtx(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user id not found in context")
	}
	return userID, nil
}

func (u *Service) SetUserIDCtx(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
