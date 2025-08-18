package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"rag-searchbot-backend/config"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connection established successfully")

	// Read migration file
	migrationPath := filepath.Join("migrations", "001_add_user_social_fields.sql")
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Split SQL statements (split by semicolon)
	statements := strings.Split(string(migrationSQL), ";")

	// Execute each statement
	for i, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		log.Printf("🔧 Executing statement %d...", i+1)

		_, err := db.Exec(statement)
		if err != nil {
			log.Printf("❌ Failed to execute statement %d: %v", i+1, err)
			log.Printf("Statement: %s", statement)
			os.Exit(1)
		}

		log.Printf("✅ Statement %d executed successfully", i+1)
	}

	log.Println("🎉 Migration completed successfully!")
	log.Println("✅ Added social media fields to users table")
	log.Println("✅ Added indexes for better performance")
}
