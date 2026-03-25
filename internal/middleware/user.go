package middleware

import (
	"context"
	"net/http"
)

type UserService interface {
	GetCookieAuthKey() string
	GetUserID(string) (int, error)
	SetUserIDCtx(context.Context, int) context.Context
}

type UserAuthMiddleware struct {
	userService UserService
}

func (u *UserAuthMiddleware) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(u.userService.GetCookieAuthKey())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		userID, err := u.userService.GetUserID(cookie.Value)
		if err != nil || userID == 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		ctx := u.userService.SetUserIDCtx(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewUserAuthMiddleware(userService UserService) *UserAuthMiddleware {
	return &UserAuthMiddleware{userService}
}
