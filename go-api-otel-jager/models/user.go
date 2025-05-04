package models

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/mu-wahba/go-api-otel-jager/db"
	"github.com/mu-wahba/go-api-otel-jager/utils"
)

func init() {
	var err error
	tp, err := utils.InitTracer("jaeger:4318")
	if err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}
	Tracer = tp.Tracer("user models")
}

type User struct {
	ID       int64  `json:"id"`
	Email    string `binding:"required" json:"email"`
	Password string `binding:"required" json:"password"`
}

// Save Function to save user in database table users
func (u *User) Save(ctx context.Context) error {
	_, span := Tracer.Start(ctx, "DB: Create a new user")
	defer span.End()
	query := `INSERT INTO users(email, password) VALUES ($1, $2) RETURNING id`

	hashedPass, err := utils.HashPassword(u.Password)
	if err != nil {
		utils.SetErrorOnSpan(span, err)
		return err
	}

	// Execute query and get the newly inserted user ID
	startTime := time.Now()
	err = db.DB.QueryRowContext(ctx, query, u.Email, hashedPass).Scan(&u.ID)
	if err != nil {
		utils.SetErrorOnSpan(span, err)
		return err

	}
	duration := time.Since(startTime)
	span.SetAttributes(utils.DbQueryAttributes(query, 0, duration, "Insert")...)

	return nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers(ctx context.Context) ([]User, error) {
	if db.DB == nil {
		return nil, errors.New("database connection is not initialized")
	}
	var users []User
	query := `SELECT id, email, password FROM users`

	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.Password)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// ValidateCreds checks if user credentials are correct
func (u *User) ValidateCreds(ctx context.Context) error {
	_, span := Tracer.Start(ctx, "user ValidateCreds from db")
	defer span.End()

	query := `SELECT id, password FROM users WHERE email = $1`
	row := db.DB.QueryRowContext(ctx, query, u.Email)
	span.SetAttributes(utils.DbQueryAttributes(query, 0, 0, "SELECT")...)

	var hashedPass string
	err := row.Scan(&u.ID, &hashedPass) // Get user ID and hashed password
	if err != nil {
		utils.SetErrorOnSpan(span, err)
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	// Validate password
	err = utils.ComparePassHash(hashedPass, u.Password)
	if err != nil {
		utils.SetErrorOnSpan(span, err)
		return errors.New("invalid credentials")
	}

	return nil
}

// // Save Function to save user in database table users
// func (u *User) Save() error {
// 	query := `INSERT INTO users(email,password) VALUES (?,?)`
// 	stmt, err := db.DB.Prepare(query)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	hashedPass, err := utils.HashPassword(u.Password)
// 	if err != nil {
// 		return err
// 	}
// 	res, err := stmt.Exec(u.Email, hashedPass)
// 	if err != nil {
// 		return err
// 	}
// 	lastInsertedId, err := res.LastInsertId()
// 	if err != nil {
// 		return err
// 	}
// 	u.ID = lastInsertedId
// 	return nil
// }

// func GetAllUsers(ctx context.Context) ([]User, error) {
// 	var users []User
// 	query := `SELECT * FROM users`
// 	rows, err := db.DB.QueryContext(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var user User
// 		err := rows.Scan(&user.ID, &user.Email, &user.Password)
// 		if err != nil {
// 			return nil, err
// 		}
// 		users = append(users, user)
// 	}
// 	return users, nil

// }

// func (u *User) ValidateCreds() error {
// 	//fetch user
// 	query := `SELECT id, password FROM users Where email = ?`
// 	row := db.DB.QueryRow(query, u.Email)
// 	var hashedPass string
// 	// var id int64
// 	err := row.Scan(&u.ID, &hashedPass) //Get user id
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			// If no rows are found, return a not-found error
// 			return errors.New(" User Not Found")
// 		}
// 		return err // Return any other error
// 	}

// 	//Validate pass
// 	err = utils.ComparePassHash(hashedPass, u.Password)
// 	if err != nil {
// 		return errors.New("invalid creds")
// 	}
// 	return nil
// }
