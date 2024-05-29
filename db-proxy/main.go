package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	//load env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	//create listener
	listener, err := net.Listen("tcp", "0.0.0.0"+":"+os.Getenv("LOCAL_PORT"))
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}

	for {
		connection, err := listener.Accept()
		//add logging for the connection
		log.Printf("Connection accepted: %v", connection.LocalAddr())
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go handleConnection(connection)
	}

}

func handleConnection(connection net.Conn) {
	//connect to actul db server
	db, err := net.Dial("tcp", os.Getenv("REMOTE_DB_HOST")+":"+os.Getenv("REMOTE_DB_PORT"))
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
		return
	}

	defer db.Close()

	go io.Copy(db, connection) //from client to db in seperate

	io.Copy(connection, db)

}

