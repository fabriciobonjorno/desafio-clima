package contract

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const maxRequestBody = 1024

var ErrInvalidCEP = errors.New("invalid zipcode")

var cepPattern = regexp.MustCompile(`^[0-9]{8}$`)

type CEPRequest struct {
	CEP string `json:"cep"`
}

func DecodeCEP(body io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxRequestBody+1))
	if err != nil || len(data) > maxRequestBody {
		return "", ErrInvalidCEP
	}

	var request CEPRequest
	if err := json.Unmarshal(data, &request); err != nil || !cepPattern.MatchString(request.CEP) {
		return "", ErrInvalidCEP
	}
	return request.CEP, nil
}
