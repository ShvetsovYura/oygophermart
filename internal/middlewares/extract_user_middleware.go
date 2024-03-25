package middlewares

import (
	"context"
	"net/http"

	"github.com/ShvetsovYura/oygophermart/internal/models"
)

type Extractor interface {
	ExtractUserID(token string) (uint64, error)
}

func ExtractUserID(ex Extractor) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Убрать в мидлварю
			c, _ := r.Cookie("token")
			if c != nil {
				userID, err := ex.ExtractUserID(c.Value)

				if err != nil {
					http.Error(w, "Unable get user", http.StatusInternalServerError)
					return
				}
				ctx := context.WithValue(r.Context(), models.UIDKey, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Unable get cookies", http.StatusBadRequest)
				return
			}

		})
	}
}
