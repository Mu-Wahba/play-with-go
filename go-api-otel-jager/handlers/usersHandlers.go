package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mu-wahba/go-api-otel-jager/models"
	"github.com/mu-wahba/go-api-otel-jager/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func init() {
	var err error
	tp, err := utils.InitTracer("jaeger:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	Tracer = tp.Tracer("user-handlers")
}

func Login(c *gin.Context) {
	ctx, span := Tracer.Start(c.Request.Context(), "user login")
	defer span.End()
	//GET USer
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		utils.SetErrorOnSpan(span, errors.New("couldn't bind json"))
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Couldn't bind json",
		})
		return
	}

	//Check if email and pass are correct from databse
	err = user.ValidateCreds(ctx)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": err.Error(),
		})
		utils.SetErrorOnSpan(span, err)
		return
	}

	fmt.Println(user.ID)

	//create jwt Token
	token, err := utils.GenerateToken(user.Email, user.ID) // we are getting id from ValidateCreds
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": err.Error()})
		utils.SetErrorOnSpan(span, err)
		return
	}

	utils.SetOkMsgOnSpan(span, codes.Ok, "user verified", "user logged in sucessfully")

	c.JSON(http.StatusOK, gin.H{
		"msg":   " Logged in",
		"token": token,
	})

}

func Signup(c *gin.Context) {
	ctx, span := Tracer.Start(c.Request.Context(), "user Signup")
	defer span.End()
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Couldn't bind json",
		})
		return
	}
	err = user.Save(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Couldn't create user !!",
		})
		return
	}
	span.AddEvent("created a new user", trace.WithAttributes(
		attribute.Int64("user id", user.ID),
	))

	c.JSON(http.StatusCreated, gin.H{
		"msg": "create user successfully",
		"id":  user.ID,
	})
}

func ListUsers(c *gin.Context) {
	// _, span := utils.Tracer.Start(c.Request.Context(), "List Users")
	// defer span.End()

	users, err := models.GetAllUsers(c.Request.Context())
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Couldn't Get users !!",
			"err": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, users)
}
