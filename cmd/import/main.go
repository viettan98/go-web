package main

import (
	"log"
	"web/internal/database"
)

func main() {
	database.InitMySQL("root:Tan0819151575@tcp(127.0.0.5:3306)/gobook_db?parseTime=true")

	log.Println("Starting CSV Import Process...")
	err := database.ImportBooksFromCSVFile("books_import_sample.csv")
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}
	log.Println("Import process completed!")
}