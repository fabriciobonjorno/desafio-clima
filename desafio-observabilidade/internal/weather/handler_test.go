package weather

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeCurrentWeather struct {
	result Result
	err    error
}

func (f fakeCurrentWeather) Current(context.Context, string) (Result, error) {
	return f.result, f.err
}

func TestHandlerResponses(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		result     Result
		serviceErr error
		status     int
		response   string
	}{
		{
			name:     "success",
			body:     `{"cep":"01001000"}`,
			result:   Result{City: "São Paulo", Celsius: 28.5, Fahrenheit: 83.3, Kelvin: 301.5},
			status:   http.StatusOK,
			response: `{"city":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}` + "\n",
		},
		{
			name:     "invalid CEP",
			body:     `{"cep":29902555}`,
			status:   http.StatusUnprocessableEntity,
			response: "invalid zipcode\n",
		},
		{
			name:       "not found",
			body:       `{"cep":"00000000"}`,
			serviceErr: ErrZipcodeNotFound,
			status:     http.StatusNotFound,
			response:   "can not find zipcode\n",
		},
		{
			name:       "upstream failure",
			body:       `{"cep":"01001000"}`,
			serviceErr: errors.New("upstream timeout"),
			status:     http.StatusBadGateway,
			response:   "weather service unavailable\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := fakeCurrentWeather{result: test.result, err: test.serviceErr}
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			NewHandler(service).ServeHTTP(response, request)
			if response.Code != test.status || response.Body.String() != test.response {
				t.Fatalf("response = %d %q, want %d %q", response.Code, response.Body.String(), test.status, test.response)
			}
		})
	}
}

func TestHandlerRejectsNonPOST(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	NewHandler(fakeCurrentWeather{}).ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", response.Code)
	}
}
