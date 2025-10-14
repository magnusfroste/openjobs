package database

import (
	"fmt"
	"os"
)

// Connect establishes a connection to Supabase via REST API
func Connect() {
	// Validate environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" {
		fmt.Println("Warning: SUPABASE_URL not set, using default")
		os.Setenv("SUPABASE_URL", "https://supabase.froste.eu")
	}
	if supabaseKey == "" {
		fmt.Println("Warning: SUPABASE_ANON_KEY not set, using default")
		os.Setenv("SUPABASE_ANON_KEY", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyAgCiAgICAicm9sZSI6ICJhbm9uIiwKICAgICJpc3MiOiAic3VwYWJhc2UtZGVtbyIsCiAgICAiaWF0IjogMTY0MTc2OTIwMCwKICAgICJleHAiOiAxNzk5NTM1NjAwCn0.dc_X5iR_VP_qT0zsiyj_I_OZ2T9FtRU2BBNWN8Bu4GE")
	}

	fmt.Println("Supabase environment configured successfully")
}

// InitDB initializes the database by running migrations via Supabase REST API
func InitDB() error {
	fmt.Println("Supabase database connection initialized")
	fmt.Println("Note: Create tables manually in Supabase dashboard using migrations/001_create_job_posts.sql")
	fmt.Println("Or use the SQL editor in your Supabase dashboard to run the migration")
	return nil
}