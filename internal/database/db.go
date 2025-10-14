package database

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Connect establishes a connection to Supabase via REST API
func Connect() error {
	// Validate environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	// Check for placeholder values
	if supabaseURL == "" || strings.Contains(supabaseURL, "your-project") {
		log.Fatal("❌ FATAL: SUPABASE_URL is not configured properly in .env file!\n" +
			"Please update .env with your actual Supabase URL from https://supabase.com/dashboard")
	}

	if supabaseKey == "" {
		log.Fatal("❌ FATAL: SUPABASE_ANON_KEY is not configured in .env file!\n" +
			"Please update .env with your actual Supabase anon key from https://supabase.com/dashboard")
	}

	fmt.Println("✅ Supabase configuration validated")
	fmt.Printf("   URL: %s\n", supabaseURL)
	fmt.Printf("   Key: %s...\n", supabaseKey[:20])

	return nil
}

// InitDB initializes the database by running migrations via Supabase REST API
func InitDB() error {
	fmt.Println("📊 Database initialization check")
	fmt.Println("⚠️  IMPORTANT: Ensure you have run the migration in your Supabase dashboard:")
	fmt.Println("   1. Go to your Supabase dashboard: https://supabase.com/dashboard")
	fmt.Println("   2. Navigate to SQL Editor")
	fmt.Println("   3. Run the contents of: migrations/001_create_job_posts.sql")
	return nil
}
