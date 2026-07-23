package middleware

import "net/http"

// CORS returns a middleware that allows cross-origin requests from the given
// client origin (CLIENT_URL). It echoes the allowed origin back, permits
// credentials, and answers preflight OPTIONS requests directly with 204.
func CORS(clientURL string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")

			if origin != "" && origin == clientURL {
				header := w.Header()
				header.Set("Access-Control-Allow-Origin", origin)
				header.Set("Access-Control-Allow-Credentials", "true")
				header.Set("Access-Control-Expose-Headers", "Set-Cookie")

				if req.Method == http.MethodOptions {
					header.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS")
					header.Set("Access-Control-Allow-Headers", "authorization,content-type,content-length")
					header.Set("Access-Control-Max-Age", "86400")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}

			next.ServeHTTP(w, req)
		})
	}
}
