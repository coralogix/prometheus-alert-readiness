package responses

import (
	"fmt"
	"net/http"
)

func Ready(writer http.ResponseWriter) {
	writer.Header().Add("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte("ok\n"))
}

func NotReady(writer http.ResponseWriter, err error) {
	writer.Header().Add("Content-Type", "text/plain")
	writer.WriteHeader(http.StatusServiceUnavailable)
	_, _ = writer.Write([]byte(fmt.Sprintf("not ok, err:\n%v\n", err)))
}
