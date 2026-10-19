package input

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"clima-por-cep/internal/contract"
)

const maxResponseBody = 1 << 20

type Handler struct {
	serviceBURL string
	client      *http.Client
}

func NewHandler(serviceBURL string, client *http.Client) *Handler {
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Handler{serviceBURL: serviceBURL, client: client}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cep, err := contract.DecodeCEP(r.Body)
	if err != nil {
		http.Error(w, contract.ErrInvalidCEP.Error(), http.StatusUnprocessableEntity)
		return
	}

	body := []byte(fmt.Sprintf(`{"cep":%q}`, cep))
	request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.serviceBURL, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := h.client.Do(request)
	if err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBody+1))
	if err != nil || len(responseBody) > maxResponseBody {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(response.StatusCode)
	_, _ = w.Write(responseBody)
}

var _ http.Handler = (*Handler)(nil)
