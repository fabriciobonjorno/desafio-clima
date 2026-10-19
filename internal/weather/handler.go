package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

var zipcodePattern = regexp.MustCompile(`^[0-9]{8}$`)

type CurrentWeather interface {
	Current(ctx context.Context, zipcode string) (Temperatures, error)
}

func NewHandler(service CurrentWeather) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		zipcode := r.URL.Query().Get("cep")
		if !zipcodePattern.MatchString(zipcode) {
			http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
			return
		}

		result, err := service.Current(r.Context(), zipcode)
		if errors.Is(err, ErrZipcodeNotFound) {
			http.Error(w, "can not find zipcode", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "weather service unavailable", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			return
		}
	})
}
