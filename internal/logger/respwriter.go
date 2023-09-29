package logger

import "net/http"

// responseData additional data for request handler log.
type responseData struct {
	status int
	size   int
}

// loggingResponseWriter override base http.ResponseWriter for logging responseData.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// createResponseWriter create method for loggingResponseWriter.
func createResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, &responseData{200, 0}}
}

// createResponseWriter returns responseData.
func (r *loggingResponseWriter) getResponseData() *responseData {
	return r.responseData
}

// Write overridden http.ResponseWriter's interface method.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader overridden http.ResponseWriter's interface method.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
