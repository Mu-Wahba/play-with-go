package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/handlers"
	"github.com/mu-wahba/go-api-otel-jager/middlewares"
)

func RegisterEventRoutes(server *gin.Engine) {
	auth := server.Group("/")
	// auth.Use(middlewares.OtelMiddleware)
	auth.Use(middlewares.Authenticate)
	auth.POST("/events", handlers.CreateEvent)
	auth.PUT("/event/:id", handlers.UpdateEvent)
	auth.DELETE("/event/:id", handlers.DeleteEvent)
	auth.POST("/events/:id/register", handlers.EventRegister)
	auth.DELETE("/events/:id/register", handlers.DeleteEventRegister)

	server.GET("/", handlers.Home)
	server.GET("/events", handlers.GetEvents)
	server.GET("/registers", handlers.GetRegisters)
	server.GET("/event/:id", handlers.GetEvent)
	server.DELETE("/clear", handlers.Basicauth(), handlers.ClearAll)

}

func RegisterUserRoutes(server *gin.Engine) {
	server.POST("/signup", handlers.Signup)
	server.POST("/login", handlers.Login)
	server.GET("/users", handlers.ListUsers)

}
