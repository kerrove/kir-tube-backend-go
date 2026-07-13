package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, req)
			return
		}
		header := w.Header()
		header.Set("Access-Control-Origin", origin)
		header.Set("Access-Control-Credentials", "true")
		if req.Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,HEAD")
			header.Set("Access-Control-Allow-Headers", "authorization,content-type,content-length")
			header.Set("Access-Control-Max-Age", "86400")
		}
		next.ServeHTTP(w, req)
	})
}
