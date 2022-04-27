package main

import (
	"database/sql"
	"fmt"
	"log"
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

	db, _ = sql.Open("postgres", psqlInfo)
	defer db.Close()

	err := db.Ping()
	if err != nil {
		panic(err)
	}

	log.Println("Successfully connected!")
}
