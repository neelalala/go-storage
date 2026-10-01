package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Content-Length, X-Requested-With, ETag, *")
		w.Header().Set("Access-Control-Expose-Headers", "ETag, Content-Length, Content-Type, X-Go-Bucket-Name, X-Go-Owner-Id, X-Go-Creation-Date, X-Go-Storage-Node-Id, X-Go-Created-At, X-Go-Updated-At")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
