// Package httputil provides shared HTTP utilities for API responses.
//
// This package contains common HTTP response helpers that ensure consistent
// JSON responses across all services in the system.
//
// Example usage:
//
//	func MyHandler(w http.ResponseWriter, r *http.Request) {
//	    data := map[string]string{"message": "success"}
//	    httputil.RespondJSON(w, http.StatusOK, data)
//	}
//
//	func ErrorHandler(w http.ResponseWriter, r *http.Request) {
//	    httputil.RespondError(w, http.StatusBadRequest, "invalid input")
//	}
package httputil
