package input

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerValidatesAndForwardsCEP(t *testing.T) {
	var receivedBody string
	called := false
	serviceB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"city":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}`))
	}))
	defer serviceB.Close()

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"cep":"01001000"}`))
	response := httptest.NewRecorder()
	NewHandler(serviceB.URL, serviceB.Client()).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", response.Code, response.Body)
	}
	if receivedBody != `{"cep":"01001000"}` || !called {
		t.Fatalf("forwarded body = %q, called=%t", receivedBody, called)
	}
	if response.Body.String() != `{"city":"São Paulo","temp_C":28.5,"temp_F":83.3,"temp_K":301.5}` {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestHandlerRejectsInvalidCEPWithoutCallingServiceB(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "number", body: `{"cep":29902555}`},
		{name: "wrong length", body: `{"cep":"1234567"}`},
		{name: "non digit", body: `{"cep":"1234abcd"}`},
		{name: "missing", body: `{}`},
		{name: "malformed", body: `{"cep":`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			called := false
			serviceB := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				called = true
			}))
			defer serviceB.Close()

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			response := httptest.NewRecorder()
			NewHandler(serviceB.URL, serviceB.Client()).ServeHTTP(response, request)
			if response.Code != http.StatusUnprocessableEntity || response.Body.String() != "invalid zipcode\n" {
				t.Errorf("response = %d %q, want 422 invalid zipcode", response.Code, response.Body.String())
			}
			if called {
				t.Error("service B was called for an invalid CEP")
			}
		})
	}
}

func TestHandlerRelaysServiceBErrors(t *testing.T) {
	serviceB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("can not find zipcode\n"))
	}))
	defer serviceB.Close()

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"cep":"00000000"}`))
	response := httptest.NewRecorder()
	NewHandler(serviceB.URL, serviceB.Client()).ServeHTTP(response, request)
	if response.Code != http.StatusNotFound || response.Body.String() != "can not find zipcode\n" {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}
