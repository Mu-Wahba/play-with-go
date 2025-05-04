package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/models"
	"github.com/mu-wahba/go-api-otel-jager/utils"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

var Tracer trace.Tracer

func init() {
	var err error
	tp, err := utils.InitTracer("jaeger:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	Tracer = tp.Tracer("event-service")
}

func Home(c *gin.Context) {
	c.String(200, "Home page")
}

func GetEvents(c *gin.Context) {
	ctx, span := Tracer.Start(c.Request.Context(), "Get All event")
	defer span.End()
	span.SetAttributes(
		semconv.HTTPClientIP(c.ClientIP()),
		semconv.HTTPUserAgent(c.Request.UserAgent()),
	)

	events, err := models.GetAllEvents(ctx)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{
			"msg": "Error getting all events",
		})
		return
	}
	c.JSON(http.StatusOK, events)
}

func GetRegisters(c *gin.Context) {
	registers, err := models.Registers(c.Request.Context())
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{
			"msg": "Error getting all registers",
		})
		return
	}
	c.JSON(http.StatusOK, registers)
}

func GetEvent(c *gin.Context) {
	// _, span := Tracer.Start(c.Request.Context(), "Get event")
	// defer span.End()
	id := c.Param("id")
	event, err := models.GetEventByID(c.Request.Context(), id)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{
			"msg": "Error getting event with specific id",
		})
		return
	}
	c.JSON(http.StatusOK, event)
}

func CreateEvent(c *gin.Context) {
	ctx, span := Tracer.Start(c.Request.Context(), "Create a new event")
	// ctx := c.MustGet("otel-context").(context.Context) //auth middleware ctx
	// ctx, span := Tracer.Start(ctx, "Create a new event")

	defer span.End()
	var event models.Event

	err := c.ShouldBindJSON(&event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "couldn't parse request",
		})
		return
	}
	userid, _ := c.Get("userid")
	fmt.Println("userid", userid)
	event.UserID = userid.(int64)

	//Save to DB
	err = event.Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Unable to create event",
		})
		return
	}

	c.JSON(http.StatusOK, event)
}

func ClearAll(c *gin.Context) {
	// _, span := Tracer.Start(c.Request.Context(), "Clear All Events")
	// defer span.End()
	err := models.ClearAll(c.Request.Context()) // Ensure this matches the correct function name
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Error clearing events table",
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"msg": "cleared events table",
	})
}

func DeleteEvent(c *gin.Context) {
	// _, span := Tracer.Start(c.Request.Context(), "Delete Event")
	// defer span.End()
	id := c.Param("id")
	//get evetnByID
	event, err := models.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Couldn't get event"})
		return
	}
	idInt := c.GetInt64("userid")

	fmt.Println("event.UserID", event.UserID)
	fmt.Println(idInt)

	if event.UserID != idInt {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "You can't delete what is not yours"})
		return

	}

	err = models.DeleteEventByID(c.Request.Context(), id)
	//Get Event and check userid
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": fmt.Sprintf("Error getting event with id: %s", id),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprintf("Event with id: %s deleted successfully ", id),
	})

}

func UpdateEvent(c *gin.Context) {
	// _, span := Tracer.Start(c.Request.Context(), "Update Event")
	// defer span.End()
	id := c.Param("id")

	event, err := models.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": fmt.Sprintf("Error getting event with id: %s", id),
		})
		return
	}

	userid := c.GetInt64("userid")
	fmt.Println(event)
	fmt.Println(event.UserID)
	fmt.Println(userid)

	if event.UserID != userid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Not yours",
		})
		return
	}

	var updatedEvent models.Event
	err = c.ShouldBindJSON(&updatedEvent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "couldn't parse request",
		})
		return
	}
	// Convert id from string to int64
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "Invalid event ID",
		})
		return
	}
	updatedEvent.ID = idInt
	err = updatedEvent.UpdateEvent(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": fmt.Sprintf("Error Updating event with id: %s", id),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":   fmt.Sprintf("Event with id: %s updated successfully", id),
		"event": updatedEvent,
	})
}

func Basicauth() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		"wahba": "wahba",
	})
}
