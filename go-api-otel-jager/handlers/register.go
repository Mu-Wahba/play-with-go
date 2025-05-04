package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/models"
	"github.com/mu-wahba/go-api-otel-jager/utils"
	"go.opentelemetry.io/otel/codes"
)

func init() {
	var err error
	tp, err := utils.InitTracer("jaeger:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	Tracer = tp.Tracer("Event-Registration")
}

func EventRegister(c *gin.Context) {
	// ctx := c.MustGet("otel-context").(context.Context) //auth middleware ctx
	// ctx, span := Tracer.Start(ctx, "Event register")

	ctx, span := Tracer.Start(c.Request.Context(), "event registration")
	defer span.End()
	//get user id , event id
	userid := c.GetInt64("userid")
	eventid := c.Param("id")

	//check if the event is in database
	event, err := models.GetEventByID(ctx, eventid)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Event Not found"})
		utils.SetErrorOnSpan(span, err)
		return
	}
	//save it to database
	err = event.Register(ctx, userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to register user for event" + err.Error()})
		utils.SetErrorOnSpan(span, err)
		return
	}

	utils.SetOkMsgOnSpan(span, codes.Ok, "register event ", fmt.Sprintf("user %d registered for eventid %s", userid, eventid))

	c.JSON(http.StatusCreated, gin.H{
		"msg": "register user for event",
	})

}

func DeleteEventRegister(c *gin.Context) {
	//get user id , event id
	userid := c.GetInt64("userid")
	eventid := c.Param("id")
	var event models.Event
	eventID, _ := strconv.ParseInt(eventid, 10, 64) // Convert string to int64
	event.ID = eventID
	fmt.Println("here", userid)
	err := event.CancelRegister(c.Request.Context(), userid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "Failed to cancel registeration user for event  " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"msg": "register cancelled",
	})

}
