package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func init() {
	var err error
	tp, err := utils.InitTracer("jaeger:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	Tracer = tp.Tracer("Auth Middleware")
}

func Authenticate(c *gin.Context) {
	// ctx := c.MustGet("otel-context").(context.Context) //auth middleware ctx
	// ctx, span := Tracer.Start(ctx, "Auth Middleware")
	fmt.Println("otel-contextotel-contextotel-contextotel-contextotel-contextotel-contextotel-context", "ctx")

	ctx, span := Tracer.Start(c.Request.Context(), "Auth Middleware")

	defer span.End()
	// Pass the context to the Gin context so it can be accessed later
	c.Set("otel-context", ctx)
	//Authentication
	authToken := c.Request.Header.Get("Authorization")
	if authToken == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "not authorized"})
		span.SetStatus(codes.Error, "not authorized")
		return
	}
	//Verify Token
	userid, err := utils.ValidateToken(authToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
		span.SetStatus(codes.Error, err.Error())
		return
	}
	//Attach userid to c
	c.Set("userid", int64(userid))
	span.AddEvent("Authentication", trace.WithAttributes(
		attribute.Int64("user id", int64(userid)),
	))

	c.Next()
}

// func OtelMiddleware(c *gin.Context) {
// 	request := fmt.Sprintf("Request %s", c.Request.URL.Path)
// 	ctx, span := Tracer.Start(c.Request.Context(), request)
// 	defer span.End()

// 	c.Set("otel-context", ctx)
// 	fmt.Println(ctx)

// 	c.Next()
// }
