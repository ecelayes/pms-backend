
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/migrate/main.go [setup|create-db|migrate-up|migrate-down|status]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "setup":
		setup()
	case "create-db":
		createDatabases()
	case "migrate-up":
		migrateUp()
	case "migrate-down":
		migrateDown()
	case "status":
		migrateStatus()
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func getDBURL(dbName string) string {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbName)
}

func getMigrateBin() string {
	bin := filepath.Join(os.Getenv("HOME"), "go", "bin", "migrate")
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	return "migrate"
}

func ensureMigrate() {
	bin := getMigrateBin()
	if _, err := os.Stat(bin); os.IsNotExist(err) {
		fmt.Println("Installing golang-migrate...")
		cmd := exec.Command("go", "install", "-tags", "postgres", "github.com/golang-migrate/migrate/v4/cmd/migrate@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to install migrate: %v", err)
		}
	}
}

func createDatabases() {
	ctx := context.Background()
	
	mainDB := getDBURL("postgres")
	conn, err := pgx.Connect(ctx, mainDB)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer conn.Close(ctx)

	dbName := os.Getenv("DB_NAME")
	testDBName := os.Getenv("DB_TEST_NAME")
	if testDBName == "" {
		testDBName = dbName + "_test"
	}

	for _, db := range []string{dbName, testDBName} {
		_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", db))
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				fmt.Printf("DB %s already exists\n", db)
			} else {
				log.Printf("Warning creating %s: %v", db, err)
			}
		} else {
			fmt.Printf("Created database: %s\n", db)
		}
	}
}

func runMigrate(args ...string) {
	ensureMigrate()
	
	migrateBin := getMigrateBin()
	dbURL := getDBURL(os.Getenv("DB_NAME"))
	
	cmd := exec.Command(migrateBin, append([]string{
		"-path", "db/migrations",
		"-database", dbURL,
	}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
}

func migrateUp() {
	fmt.Println("Running migrations...")
	runMigrate("up")
}

func migrateDown() {
	fmt.Println("Rolling back last migration...")
	runMigrate("down", "1")
}

func migrateStatus() {
	ensureMigrate()
	
	migrateBin := getMigrateBin()
	dbURL := getDBURL(os.Getenv("DB_NAME"))
	
	cmd := exec.Command(migrateBin, "-path", "db/migrations", "-database", dbURL, "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func setup() {
	fmt.Println("=== Full database setup ===")
	createDatabases()
	
	fmt.Println("\n=== Running migrations ===")
	runMigrate("up")
	
	fmt.Println("\n=== Migration status ===")
	migrateStatus()
	
	testDB := getDBURL(os.Getenv("DB_TEST_NAME"))
	fmt.Printf("\nTest DB URL: %s\n", testDB)
	fmt.Println("\nRun tests with:")
	fmt.Printf("  TEST_DATABASE_URL=\"%s\" go test ./tests/...\n", testDB)
}
