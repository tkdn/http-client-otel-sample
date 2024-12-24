package instrument_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tkdn/http-client-otel-sample/instrument"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// see: [httptrace test examples] and [InMemoryExporter test examples]
//
// [httptrace test examples]: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/00786cc3fe5b58f2c0e14235eab8fb4d8ef1ae43/instrumentation/net/http/otelhttp/test/transport_test.go#L136
// [InMemoryExporter test examples]: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/00786cc3fe5b58f2c0e14235eab8fb4d8ef1ae43/instrumentation/github.com/labstack/echo/otelecho/test/echo_test.go#L204
func TestHTTPConnTracer(t *testing.T) {
	ctx := context.Background()
	imsb := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(imsb))
	content := []byte("大事な指標を取るためのテスト!")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(content)
		if err != nil {
			t.Fatal(err)
		}
	}))

	r, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	r = r.WithContext(ctx)

	tr := instrument.NewHTTPTransport(provider)
	c := http.Client{Transport: tr}
	_, err = c.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		imsb.Reset()
	}()

	spans := imsb.GetSpans()
	if spans[0].Name != "めっちゃ大事なスパン" {
		t.Errorf("span has no `めっちゃ大事なスパン`:\n%+v\n", spans[0])
	}
}
