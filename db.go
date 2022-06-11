package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

var (
	dbHost     = ""
	dbPort     = 0
	dbUser     = ""
	dbPassword = ""
	dbName     = ""
)

var db *sql.DB

func setDb() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	for {
		var err error
		db, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	err := db.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected!")
}
