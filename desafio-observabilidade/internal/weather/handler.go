package weather

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"clima-por-cep/internal/contract"
)

type CurrentWeather interface {
	Current(context.Context, string) (Result, error)
}

func NewHandler(service CurrentWeather) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		zipcode, err := contract.DecodeCEP(r.Body)
		if err != nil {
			http.Error(w, contract.ErrInvalidCEP.Error(), http.StatusUnprocessableEntity)
			return
		}

		result, err := service.Current(r.Context(), zipcode)
		if errors.Is(err, ErrZipcodeNotFound) {
			http.Error(w, ErrZipcodeNotFound.Error(), http.StatusNotFound)
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
