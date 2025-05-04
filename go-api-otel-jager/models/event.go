package models

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mu-wahba/go-api-otel-jager/db"
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
	Tracer = tp.Tracer("event-service")
}

type Event struct {
	ID          int64     `json:"id"`
	Name        string    `binding:"required" json:"name"`
	Description string    `binding:"required" json:"description"`
	Location    string    `binding:"required" json:"location"`
	DateTime    time.Time `json:"date_time"`
	UserID      int64     `json:"user_id"`
}

type Registration struct {
	ID      int64  `json:"id"`
	UserID  string `binding:"required" json:"user_id"`
	EventID string `binding:"required" json:"event_id"`
}

// Save Function to save an event in the database table `events`
func (e *Event) Save(ctx context.Context) error {
	_, span := Tracer.Start(ctx, "Create new event")
	defer span.End()
	query := `INSERT INTO events(name, description, location, dateTime, user_id) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	span.SetAttributes(
		attribute.String("db.statement", query),
	)

	// Execute query and get the newly inserted event ID
	err := db.DB.QueryRowContext(ctx, query, e.Name, e.Description, e.Location, e.DateTime, e.UserID).Scan(&e.ID)
	if err != nil {
		fmt.Println("Couldn't insert event:", err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("error.message", err.Error()),
			attribute.String("error.details", "Error,  couldn't insert event to db."),
		)
		return err
	}

	span.AddEvent("created a new event", trace.WithAttributes(
		attribute.Int("user id", int(e.UserID)),
		attribute.String("Event Name", e.Name),
	))

	return nil
}

// GetAllEvents retrieves all events from the database
func GetAllEvents(ctx context.Context) ([]Event, error) {
	_, span := Tracer.Start(ctx, "Get All event db")
	defer span.End()

	var events []Event
	query := `SELECT id, name, description, location, dateTime, user_id FROM events`
	span.SetAttributes(
		attribute.String("db.statement", query),
	)

	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
		err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.DateTime, &event.UserID)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
		span.AddEvent("Rows fetched", trace.WithAttributes(
			attribute.Int("id", int(event.ID)),
			attribute.String("name", event.Name),
		))
	}

	return events, nil
}

// Registers retrieves all registrations from the database
func Registers(ctx context.Context) ([]Registration, error) {
	_, span := Tracer.Start(ctx, "DB: Register Event")
	defer span.End()

	var registers []Registration
	query := `SELECT id, event_id, user_id FROM registrations`
	now := time.Now()
	rows, err := db.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var register Registration
		err := rows.Scan(&register.ID, &register.EventID, &register.UserID)
		if err != nil {
			return nil, err
		}

		registers = append(registers, register)
	}
	duration := time.Since(now)
	span.SetAttributes(utils.DbQueryAttributes(query, 0, duration, "GET")...)

	return registers, nil
}

// GetEventByID retrieves an event by its ID
func GetEventByID(ctx context.Context, id string) (Event, error) {
	_, span := Tracer.Start(ctx, "DB: Get All event db")
	defer span.End()
	now := time.Now()
	var event Event
	query := `SELECT id, name, description, location, dateTime, user_id FROM events WHERE id = $1`
	err := db.DB.QueryRowContext(ctx, query, id).Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.DateTime, &event.UserID)
	if err != nil {
		return event, err
	}
	duration := time.Since(now)
	span.SetAttributes(utils.DbQueryAttributes(query, 0, duration, "GET")...)

	return event, nil
}

// ClearAll deletes all records from the `events` table and resets its sequence
func ClearAll(ctx context.Context) error {
	_, err := db.DB.ExecContext(ctx, "DELETE FROM events")
	if err != nil {
		return err
	}

	// Reset the sequence in PostgreSQL
	_, err = db.DB.ExecContext(ctx, "ALTER SEQUENCE events_id_seq RESTART WITH 1")
	if err != nil {
		return err
	}

	return nil
}

// DeleteEventByID deletes an event by its ID
func DeleteEventByID(ctx context.Context, id string) error {
	query := `DELETE FROM events WHERE id = $1`
	res, err := db.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no event with id: %s found", id)
	}

	return nil
}

// UpdateEvent updates an existing event
func (e Event) UpdateEvent(ctx context.Context) error {
	query := `
		UPDATE events SET 
		name = $1, description = $2, location = $3, dateTime = $4 
		WHERE id = $5
	`
	res, err := db.DB.ExecContext(ctx, query, e.Name, e.Description, e.Location, e.DateTime, e.ID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no event with id: %d found to update", e.ID)
	}

	return nil
}

// Register a user for an event
func (e *Event) Register(ctx context.Context, id int64) error {
	_, span := Tracer.Start(ctx, "DB: Register event")
	defer span.End()
	now := time.Now()
	query := `INSERT INTO registrations (event_id, user_id) VALUES ($1, $2)`
	_, err := db.DB.ExecContext(ctx, query, e.ID, id)
	if err != nil {
		utils.SetErrorOnSpan(span, err)

		return errors.New("Couldn't create registration: " + err.Error())
	}
	duration := time.Since(now)
	span.SetAttributes(utils.DbQueryAttributes(query, 0, duration, "Insert")...)

	return nil
}

// CancelRegister removes a user's registration for an event
func (e Event) CancelRegister(ctx context.Context, id int64) error {
	query := `DELETE FROM registrations WHERE event_id = $1 AND user_id = $2`
	res, err := db.DB.ExecContext(ctx, query, e.ID, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("unauthorized: no registration found to delete")
	}

	return nil
}

// var events = []Event{}

// func (e *Event) Save() error {
// 	query := `INSERT INTO events(name,description,location,dateTime,user_id) VALUES (?,?,?,?,?)`
// 	res, err := db.DB.Exec(query, e.Name, e.Description, e.Location, e.DateTime, e.UserID)
// 	if err != nil {
// 		fmt.Println("couldn't insert event", err)
// 		return err
// 	}
// 	id, err := res.LastInsertId()
// 	e.ID = id
// 	if err != nil {
// 		fmt.Println("couldn't insert event", err)
// 		return err
// 	}
// 	return nil
// }

// func GetAllEvents() ([]Event, error) {
// 	var events []Event
// 	rows, err := db.DB.Query("SELECT * from events")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var event Event
// 		err := rows.Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.DateTime, &event.UserID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		events = append(events, event)
// 	}

// 	return events, nil
// }

// func Regsiters() ([]Registers, error) {
// 	var registers []Registers
// 	rows, err := db.DB.Query("SELECT * from registerations")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	for rows.Next() {
// 		var register Registers
// 		err := rows.Scan(&register.ID, &register.EventID, &register.UserID)
// 		if err != nil {
// 			return nil, err
// 		}
// 		registers = append(registers, register)

// 	}
// 	return registers, nil
// }
// func GetEventByID(id string) (Event, error) {
// 	var event Event
// 	err := db.DB.QueryRow("SELECT * FROM events WHERE id = ?", id).Scan(&event.ID, &event.Name, &event.Description, &event.Location, &event.DateTime, &event.UserID)
// 	if err != nil {
// 		return event, err
// 	}
// 	return event, nil
// }

// func ClearAll() error {
// 	_, err := db.DB.Exec("DELETE FROM events")
// 	if err != nil {
// 		return err
// 	}
// 	// Reset the auto-increment counter
// 	_, err = db.DB.Exec("DELETE FROM sqlite_sequence WHERE name='events'") // For SQLite
// 	// OR
// 	// _, err = db.DB.Exec("ALTER TABLE events AUTO_INCREMENT = 1") // For MySQL
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func DeleteEventByID(id string) error {
// 	query := `DELETE From events Where id=?`
// 	res, err := db.DB.Exec(query, id)
// 	if err != nil {
// 		return err
// 	}
// 	rowsAffected, err := res.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("no event with id: %s found", id)
// 	}

// 	return nil
// }

// func (e Event) UpdateEvent() error {
// 	query := `
// 	UPDATE events SET
// 	name = ? , description=?,location=?,dateTime=?
// 	WHERE id = ?
// 	`
// 	res, err := db.DB.Exec(query, e.Name, e.Description, e.Location, e.DateTime, e.ID)
// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := res.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("no event with id: %d found to update", e.ID)
// 	}

// 	return nil
// }

// func (e *Event) Register(id int64) error {
// 	query := `INSERT INTO registerations (event_id, user_id) VALUES(?,?)`
// 	stmt, err := db.DB.Prepare(query)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	_, err = stmt.Exec(e.ID, id)
// 	if err != nil {
// 		return errors.New("Couldn't create registeration" + err.Error())
// 	}

// 	return nil

// }
// func (e Event) CancelRegister(id int64) error {
// 	query := `DELETE FROM registerations WHERE event_id=? AND user_id=? `
// 	stmt, err := db.DB.Prepare(query)
// 	if err != nil {
// 		return err
// 	}
// 	defer stmt.Close()
// 	res, err := stmt.Exec(e.ID, id)
// 	n, _ := res.RowsAffected()

// 	if n == 0 {
// 		fmt.Println("nnnnn", n)
// 		return errors.New("un authorized")
// 	}
// 	// if numrows, err := res.RowsAffected(); numrows == 0 || err != nil {
// 	// 	fmt.Println("sdfsdfs", numrows, err.Error())
// 	// 	// return errors.New("Unaithorized" + err.Error())
// 	// }
// 	if err != nil {
// 		return errors.New("Couldn't Delete registeration" + err.Error())
// 	}

// 	return nil

// }
