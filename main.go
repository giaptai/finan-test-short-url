package main

import (
	"github.com/giaptai/finan-test-short-url/database"
	
	"log"
)

func main() {
	db, err := database.Connection()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}
	defer db.Close()
}
