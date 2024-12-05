package middleware

import (
	"context"
	"net/http"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/token"

	"github.com/go-chi/render"
)

type contextKey string

const claimsKey contextKey = "claims"

// UserMustBeAuthenticated is a middleware that checks if the user is authenticated.
func UserMustBeAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if the user is authenticated
		// if not, return an error

		// decode the token from the request header
		claims, err := token.DecodeTokenFromRequest(r)
		if err != nil {
			render.Status(r, err.(*apierrors.APIError).HTTPStatus)
			render.JSON(w, r, err)
			return
		}

		// check whether the token is not expired
		if time.Now().After(claims.ExpiresAt.Time) {
			e := apierrors.ErrTokenExpired
			e.Message = "token is expired"
			render.Status(r, e.HTTPStatus)
			render.JSON(w, r, e)
			return
		}

		// add the claims to the request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, claimsKey, claims)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
