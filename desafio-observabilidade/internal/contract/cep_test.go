package contract

import (
	"strings"
	"testing"
)

func TestDecodeCEP(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    string
		wantErr bool
	}{
		{name: "valid string", body: `{"cep":"01001000"}`, want: "01001000"},
		{name: "number is not a string", body: `{"cep":1001000}`, wantErr: true},
		{name: "wrong length", body: `{"cep":"1234567"}`, wantErr: true},
		{name: "non digits", body: `{"cep":"1234abcd"}`, wantErr: true},
		{name: "missing field", body: `{}`, wantErr: true},
		{name: "malformed JSON", body: `{"cep":`, wantErr: true},
		{name: "multiple JSON values", body: `{"cep":"01001000"} {}`, wantErr: true},
		{name: "oversized body", body: `{"cep":"01001000","extra":"` + strings.Repeat("x", maxRequestBody) + `"}`, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecodeCEP(strings.NewReader(test.body))
			if (err != nil) != test.wantErr {
				t.Fatalf("DecodeCEP() error = %v, wantErr %t", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("DecodeCEP() = %q, want %q", got, test.want)
			}
		})
	}
}
