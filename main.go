package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	lotel "github.com/rhiadc/grpc_api/client/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func main() {

	ctx := context.Background()
	exp, err := lotel.NewExporter(ctx)

	if err != nil {
		log.Fatalf("Error: failed to initialize exporter: %v", err)
	}

	tp := lotel.NewTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	tracer := tp.Tracer("restapi", trace.WithInstrumentationVersion("1.0.0"))

	g := gin.Default()
	g.Use(otelgin.Middleware("restapi"))

	g.GET("/add/:a/:b", func(ctx *gin.Context) {

		a := parseint(ctx.Param("a"), ctx)
		b := parseint(ctx.Param("b"), ctx)

		_, span := tracer.Start(ctx, "add")
		defer span.End()
		result := *a + *b
		span.SetAttributes(
			attribute.String("value", string(result)),
		)
		// setting span as successful
		span.SetStatus(codes.Ok, "Success")
		ctx.JSON(http.StatusOK, gin.H{"result": result})

	})

	if err := g.Run(":8089"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func parseint(param string, ctx *gin.Context) *uint64 {
	a, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter"})
		return nil
	}
	return &a
}
