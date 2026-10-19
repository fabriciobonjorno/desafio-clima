package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrZipcodeNotFound = errors.New("can not find zipcode")

const (
	viacepURL    = "https://viacep.com.br/ws"
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
)

type Temperatures struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
}

type Service struct {
	client       *http.Client
	viacepURL    string
	geocodingURL string
	forecastURL  string
}

func NewService(client *http.Client) *Service {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	return &Service{
		client:       client,
		viacepURL:    viacepURL,
		geocodingURL: geocodingURL,
		forecastURL:  forecastURL,
	}
}

func (s *Service) Current(ctx context.Context, zipcode string) (Temperatures, error) {
	location, err := s.lookupZipcode(ctx, zipcode)
	if err != nil {
		return Temperatures{}, err
	}

	coordinates, err := s.lookupCoordinates(ctx, location)
	if err != nil {
		return Temperatures{}, err
	}

	celsius, err := s.lookupTemperature(ctx, coordinates)
	if err != nil {
		return Temperatures{}, err
	}

	return Temperatures{
		Celsius:    round(celsius, 1),
		Fahrenheit: round(celsius*1.8+32, 1),
		Kelvin:     round(celsius+273, 2),
	}, nil
}

type location struct {
	City  string `json:"localidade"`
	State string `json:"uf"`
	Error bool   `json:"erro"`
}

func (s *Service) lookupZipcode(ctx context.Context, zipcode string) (location, error) {
	endpoint := fmt.Sprintf("%s/%s/json/", s.viacepURL, zipcode)
	var result location
	if err := s.getJSON(ctx, endpoint, &result); err != nil {
		return location{}, err
	}
	if result.Error || result.City == "" || result.State == "" {
		return location{}, ErrZipcodeNotFound
	}
	return result, nil
}

type coordinates struct {
	Latitude  float64
	Longitude float64
}

type geocodingResponse struct {
	Results []struct {
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		CountryCode string  `json:"country_code"`
	} `json:"results"`
}

func (s *Service) lookupCoordinates(ctx context.Context, place location) (coordinates, error) {
	query := url.Values{
		"name":     {place.City + ", " + place.State + ", Brazil"},
		"count":    {"10"},
		"language": {"pt"},
		"format":   {"json"},
	}
	endpoint := s.geocodingURL + "?" + query.Encode()
	var result geocodingResponse
	if err := s.getJSON(ctx, endpoint, &result); err != nil {
		return coordinates{}, err
	}
	for _, match := range result.Results {
		if strings.EqualFold(match.CountryCode, "BR") {
			return coordinates{Latitude: match.Latitude, Longitude: match.Longitude}, nil
		}
	}
	return coordinates{}, errors.New("city not found in weather geocoding service")
}

type forecastResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
	} `json:"current"`
}

func (s *Service) lookupTemperature(ctx context.Context, point coordinates) (float64, error) {
	query := url.Values{
		"latitude":  {fmt.Sprintf("%f", point.Latitude)},
		"longitude": {fmt.Sprintf("%f", point.Longitude)},
		"current":   {"temperature_2m"},
	}
	endpoint := s.forecastURL + "?" + query.Encode()
	var result forecastResponse
	if err := s.getJSON(ctx, endpoint, &result); err != nil {
		return 0, err
	}
	return result.Current.Temperature, nil
}

func (s *Service) getJSON(ctx context.Context, endpoint string, destination any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create upstream request: %w", err)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return fmt.Errorf("call upstream service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("upstream service returned status %d", response.StatusCode)
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode upstream response: %w", err)
	}
	return nil
}

func round(value float64, places int) float64 {
	factor := 1.0
	for range places {
		factor *= 10
	}
	return math.Round(value*factor) / factor
}
