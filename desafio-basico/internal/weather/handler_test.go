package weather

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeWeather struct {
	result Temperatures
	err    error
	called bool
}

func (f *fakeWeather) Current(_ context.Context, _ string) (Temperatures, error) {
	f.called = true
	return f.result, f.err
}

func TestHandlerResponses(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		weatherErr error
		status     int
		body       string
		called     bool
	}{
		{name: "invalid format", url: "/?cep=1234abcd", status: http.StatusUnprocessableEntity, body: "invalid zipcode\n"},
		{name: "not found", url: "/?cep=00000000", weatherErr: ErrZipcodeNotFound, status: http.StatusNotFound, body: "can not find zipcode\n", called: true},
		{name: "upstream unavailable", url: "/?cep=01001000", weatherErr: errors.New("timeout"), status: http.StatusBadGateway, body: "weather service unavailable\n", called: true},
		{name: "success", url: "/?cep=01001000", status: http.StatusOK, body: "{\"temp_C\":28.5,\"temp_F\":83.3,\"temp_K\":301.5}\n", called: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeWeather{
				result: Temperatures{Celsius: 28.5, Fahrenheit: 83.3, Kelvin: 301.5},
				err:    test.weatherErr,
			}
			request := httptest.NewRequest(http.MethodGet, test.url, nil)
			response := httptest.NewRecorder()
			NewHandler(service).ServeHTTP(response, request)

			if response.Code != test.status {
				t.Errorf("status = %d, want %d", response.Code, test.status)
			}
			if response.Body.String() != test.body {
				t.Errorf("body = %q, want %q", response.Body.String(), test.body)
			}
			if service.called != test.called {
				t.Errorf("service called = %t, want %t", service.called, test.called)
			}
		})
	}
}

func TestHandlerRejectsNonGET(t *testing.T) {
	service := &fakeWeather{}
	request := httptest.NewRequest(http.MethodPost, "/?cep=01001000", nil)
	response := httptest.NewRecorder()
	NewHandler(service).ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if service.called {
		t.Fatal("service should not be called for non-GET requests")
	}
}
