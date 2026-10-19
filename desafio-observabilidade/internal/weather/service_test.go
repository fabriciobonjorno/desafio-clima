package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCurrentReturnsCityAndConvertsTemperatures(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ws/01001000/json/":
			_, _ = w.Write([]byte(`{"localidade":"São Paulo","uf":"SP"}`))
		case "/v1/search":
			_, _ = w.Write([]byte(`{"results":[{"latitude":-23.55,"longitude":-46.63,"country_code":"BR"}]}`))
		case "/v1/forecast":
			_, _ = w.Write([]byte(`{"current":{"temperature_2m":28.5}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()

	service := NewService(upstream.Client())
	service.viacepURL = upstream.URL + "/ws"
	service.geocodingURL = upstream.URL + "/v1/search"
	service.forecastURL = upstream.URL + "/v1/forecast"

	got, err := service.Current(context.Background(), "01001000")
	if err != nil {
		t.Fatalf("Current() error = %v", err)
	}
	want := Result{City: "São Paulo", Celsius: 28.5, Fahrenheit: 83.3, Kelvin: 301.5}
	if got != want {
		t.Fatalf("Current() = %#v, want %#v", got, want)
	}
}

func TestLookupZipcodeNotFound(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"erro":true}`))
	}))
	defer upstream.Close()

	service := NewService(upstream.Client())
	service.viacepURL = upstream.URL
	_, err := service.Current(context.Background(), "00000000")
	if err != ErrZipcodeNotFound {
		t.Fatalf("Current() error = %v, want %v", err, ErrZipcodeNotFound)
	}
}

func TestCurrentRejectsMissingTemperature(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ws/01001000/json/":
			_, _ = w.Write([]byte(`{"localidade":"São Paulo","uf":"SP"}`))
		case "/v1/search":
			_, _ = w.Write([]byte(`{"results":[{"latitude":-23.55,"longitude":-46.63,"country_code":"BR"}]}`))
		case "/v1/forecast":
			_, _ = w.Write([]byte(`{"current":{}}`))
		}
	}))
	defer upstream.Close()

	service := NewService(upstream.Client())
	service.viacepURL = upstream.URL + "/ws"
	service.geocodingURL = upstream.URL + "/v1/search"
	service.forecastURL = upstream.URL + "/v1/forecast"
	if _, err := service.Current(context.Background(), "01001000"); err == nil {
		t.Fatal("Current() succeeded without a current temperature")
	}
}
