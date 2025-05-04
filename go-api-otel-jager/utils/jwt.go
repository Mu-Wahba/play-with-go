package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secretKey = "sdfmppr5sdfaka"

func GenerateToken(email string, userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":  email,
		"userID": userID,
		"exp":    time.Now().Add(time.Hour * 2).Unix(), //Valid for 2 hours
	})
	return token.SignedString([]byte(secretKey))

}

func ValidateToken(mtoken string) (float64, error) {
	parsedToken, err := jwt.Parse(mtoken, func(token *jwt.Token) (interface{}, error) {
		//check signing method
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("wrong signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, err
	}
	//we have parsed tolen now we can extract any data
	isValid := parsedToken.Valid
	if !isValid {
		return 0, errors.New("invalid token")
	}
	claims, ok := parsedToken.Claims.(jwt.MapClaims) // Correctly assert claims
	if !ok {
		return 0, errors.New("could not extract claims")
	}
	// email := claims["email"].(string)
	userId := claims["userID"].(float64)
	return userId, nil

}
