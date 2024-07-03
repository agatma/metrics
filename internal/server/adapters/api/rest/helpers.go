package rest

import (
	"errors"
	"metrics/internal/server/core/domain"
	"net/http"
)

func handleSetMetricError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrItemNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrIncorrectMetricType) || errors.Is(err, domain.ErrIncorrectMetricValue):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleGetMetricError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrIncorrectMetricType) || errors.Is(err, domain.ErrItemNotFound) {
		http.Error(w, domain.ErrItemNotFound.Error(), http.StatusNotFound)
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
