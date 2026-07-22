package http

import (
	"net/http"
	"strings"
)

var systemHeadersList = []string{
	"Content-Disposition",
	"Cache-Control",
	"Content-Encoding",
	"Content-Language",
}

func ExtractMetadata(r *http.Request) (systemMeta map[string]string, userMeta map[string]string) {
	systemMeta = make(map[string]string)
	userMeta = make(map[string]string)

	for key, values := range r.Header {
		lowerKey := strings.ToLower(key)
		if strings.HasPrefix(lowerKey, "x-amz-meta-") {
			if len(values) > 0 {
				userMeta[lowerKey] = values[0]
			}
		}
	}

	for _, headerName := range systemHeadersList {
		val := r.Header.Get(headerName)
		if val != "" {
			systemMeta[strings.ToLower(headerName)] = val
		}
	}

	return systemMeta, userMeta
}
