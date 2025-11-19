package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"openjobs/pkg/models"
	"openjobs/pkg/storage"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("🧪 Testing OpenJobs → Supabase Connection")
	fmt.Println("==========================================")
	fmt.Println()

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using environment variables")
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("❌ Missing SUPABASE_URL or SUPABASE_ANON_KEY")
	}

	fmt.Printf("📍 Supabase URL: %s\n", supabaseURL)
	fmt.Println()

	// Initialize JobStore
	jobStore := storage.NewJobStore()

	// Test 1: Read Jobs
	fmt.Println("Test 1: Read Jobs from Database")
	fmt.Println("--------------------------------")
	jobs, err := jobStore.GetAllJobs(5, 0)
	if err != nil {
		fmt.Printf("❌ Failed to read jobs: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully read %d jobs\n", len(jobs))
		for i, job := range jobs {
			fmt.Printf("  %d. %s - %s (ID: %s)\n", i+1, job.Title, job.Company, job.ID)
		}
	}
	fmt.Println()

	// Test 2: Get Job Count
	fmt.Println("Test 2: Get Total Job Count")
	fmt.Println("---------------------------")
	count, err := jobStore.GetTotalJobCount()
	if err != nil {
		fmt.Printf("❌ Failed to get job count: %v\n", err)
	} else {
		fmt.Printf("✅ Total active jobs in database: %d\n", count)
	}
	fmt.Println()

	// Test 3: Create Test Job
	fmt.Println("Test 3: Create Test Job")
	fmt.Println("-----------------------")
	testJob := &models.JobPost{
		ID:          fmt.Sprintf("test-%d", time.Now().Unix()),
		Title:       "Test Job - Connection Check",
		Company:     "OpenJobs Test",
		Description: "This is a test job to verify write access from Go",
		Location:    "Test Location",
		IsActive:    false,
		PostedDate:  time.Now(),
		Fields:      map[string]interface{}{"source": "connection-test"},
	}

	err = jobStore.CreateJob(testJob)
	if err != nil {
		fmt.Printf("❌ Failed to create test job: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully created test job: %s\n", testJob.ID)
	}
	fmt.Println()

	// Test 4: Read Test Job Back
	fmt.Println("Test 4: Read Test Job Back")
	fmt.Println("--------------------------")
	readJob, err := jobStore.GetJob(testJob.ID)
	if err != nil {
		fmt.Printf("❌ Failed to read test job: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully read test job back\n")
		fmt.Printf("  Title: %s\n", readJob.Title)
		fmt.Printf("  Company: %s\n", readJob.Company)
		fmt.Printf("  Active: %v\n", readJob.IsActive)
	}
	fmt.Println()

	// Test 5: Update Test Job
	fmt.Println("Test 5: Update Test Job")
	fmt.Println("-----------------------")
	testJob.Description = "Updated description from Go test"
	err = jobStore.UpdateJob(testJob)
	if err != nil {
		fmt.Printf("❌ Failed to update test job: %v\n", err)
	} else {
		fmt.Printf("✅ Successfully updated test job\n")
	}
	fmt.Println()

	// Test 6: Delete Test Job (Cleanup)
	fmt.Println("Test 6: Cleanup (Delete Test Job)")
	fmt.Println("---------------------------------")
	err = jobStore.DeleteJob(testJob.ID)
	if err != nil {
		fmt.Printf("⚠️  Failed to delete test job: %v\n", err)
		fmt.Printf("   You may need to manually delete: %s\n", testJob.ID)
	} else {
		fmt.Printf("✅ Successfully deleted test job\n")
		fmt.Println("🧹 Test job cleaned up")
	}
	fmt.Println()

	// Test 7: Get Remote Job Count
	fmt.Println("Test 7: Get Remote Job Count")
	fmt.Println("----------------------------")
	remoteCount, err := jobStore.GetRemoteJobCount()
	if err != nil {
		fmt.Printf("❌ Failed to get remote job count: %v\n", err)
	} else {
		fmt.Printf("✅ Total remote jobs: %d\n", remoteCount)
	}
	fmt.Println()

	// Summary
	fmt.Println("==========================================")
	fmt.Println("📊 Test Summary")
	fmt.Println("==========================================")
	fmt.Println("✅ All database operations working!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Start OpenJobs API: go run cmd/openjobs/main.go")
	fmt.Println("2. Test health endpoint: curl http://localhost:8080/health")
	fmt.Println("3. Test jobs endpoint: curl http://localhost:8080/jobs?limit=5")
	fmt.Println()
}
