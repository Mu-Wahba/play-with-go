package main

import (
	"net/http"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/db"
	"github.com/mu-wahba/go-api-otel-jager/routes"
)

func init() {
	db.InitDB()

}

func main() {
	// tp, err := utils.InitTracer("localhost:4318") // Jaeger collector endpoint
	// if err != nil {
	// 	log.Fatalf("Failed to initialize tracer: %v", err)
	// }
	// defer func() {
	// 	if err := tp.Shutdown(context.Background()); err != nil {
	// 		log.Printf("Error shutting down tracer provider: %v", err)
	// 	}
	// }()

	router := gin.Default()
	// router.Use(otelgin.Middleware("go-api"))
	routes.RegisterEventRoutes(router)
	routes.RegisterUserRoutes(router)

	fmt.Println("starting server on port 3000")
	if err := http.ListenAndServe(":3000", router); err != nil {
		log.Fatal("Couldn't start server ", err)
	}
}
