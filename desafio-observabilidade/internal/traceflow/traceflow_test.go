package traceflow_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clima-por-cep/internal/input"
	"clima-por-cep/internal/telemetry"
	"clima-por-cep/internal/weather"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestTracePropagatesFromServiceAToServiceBAndExternalLookups(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Errorf("shutdown tracer provider: %v", err)
		}
	})

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

	weatherService := weather.NewService(upstream.Client())
	serviceB := httptest.NewServer(telemetry.HTTPHandler("service-b", weather.NewHandler(weatherService)))
	defer serviceB.Close()

	serviceAClient := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	serviceA := httptest.NewServer(telemetry.HTTPHandler("service-a", input.NewHandler(serviceB.URL, serviceAClient)))
	defer serviceA.Close()

	request, err := http.NewRequest(http.MethodPost, serviceA.URL+"/", strings.NewReader(`{"cep":"01001000"}`))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("request Service A: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.StatusCode)
	}

	spans := exporter.GetSpans()
	traceID := "4bf92f3577b34da6a3ce929d0e0e4736"
	wantNames := map[string]bool{
		"service-a":          false,
		"service-b":          false,
		"zipcode.lookup":     false,
		"temperature.lookup": false,
	}
	var serviceAServer, serviceBServer, serviceAClientSpan trace.SpanID
	var serviceAClientParent, serviceBServerParent trace.SpanID
	manualParents := map[string]trace.SpanID{}
	for _, span := range spans {
		if span.SpanContext.TraceID().String() != traceID {
			t.Errorf("span %q has trace ID %s, want %s", span.Name, span.SpanContext.TraceID(), traceID)
		}
		if _, ok := wantNames[span.Name]; ok {
			wantNames[span.Name] = true
		}
		if span.Name == "service-a" && span.SpanKind == trace.SpanKindServer {
			serviceAServer = span.SpanContext.SpanID()
		}
		if span.Name == "service-b" && span.SpanKind == trace.SpanKindServer {
			serviceBServer = span.SpanContext.SpanID()
			serviceBServerParent = span.Parent.SpanID()
		}
		if span.SpanKind == trace.SpanKindClient {
			serviceAClientSpan = span.SpanContext.SpanID()
			serviceAClientParent = span.Parent.SpanID()
		}
		if span.Name == "zipcode.lookup" || span.Name == "temperature.lookup" {
			manualParents[span.Name] = span.Parent.SpanID()
		}
	}
	spanNames := make([]string, 0, len(spans))
	for _, span := range spans {
		spanNames = append(spanNames, span.Name)
	}
	for name, found := range wantNames {
		if !found {
			t.Errorf("trace is missing span %q; captured %d spans: %v", name, len(spans), spanNames)
		}
	}
	if serviceAServer == (trace.SpanID{}) || serviceAClientSpan == (trace.SpanID{}) || serviceBServer == (trace.SpanID{}) {
		t.Fatal("trace is missing an HTTP server or client span")
	}
	if serviceAClientParent != serviceAServer {
		t.Error("Service A outbound HTTP span is not a child of its inbound request span")
	}
	if serviceBServerParent != serviceAClientSpan {
		t.Error("Service B server span did not continue the Service A outbound span")
	}
	for name, parent := range manualParents {
		if parent != serviceBServer {
			t.Errorf("manual span %q is not a child of Service B's request span", name)
		}
	}
}
