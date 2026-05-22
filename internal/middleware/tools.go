package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %v: Activity on %s\n", r.Method, time.Now().Format("15:33:02"), r.URL.Path)
		next(w, r)
	}
}
