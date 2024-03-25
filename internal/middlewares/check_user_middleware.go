package middlewares

import "net/http"

type Vaidator interface {
	ValidateSign(token string) (bool, error)
}

func CheckAuthCookie(v Vaidator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "error get auth cookies", http.StatusUnauthorized)
				return
			}
			valid, err := v.ValidateSign(c.Value)
			if err != nil {
				http.Error(w, "error on check auth cookie", http.StatusInternalServerError)
				return
			}
			if !valid {
				http.Error(w, "not valid auth cookies", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
