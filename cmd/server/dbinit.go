package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Invoke with `go run internal/server/dbinit.go cmd/server/schema.sql /tmp/lbx.db`

func main() {
	// Read schema filename and db filename from command line
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <schema.sql> <database.db>", os.Args[0])
	}

	// Read the schema
	schema, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read schema.sql: %v", err)
	}

	// Open a connection to the SQLite database
	db, err := sql.Open("sqlite3", os.Args[2])
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Execute the schema
	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalf("Failed to execute schema: %v", err)
	}

	fmt.Println("Schema created successfully")
}
