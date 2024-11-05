package database

import (
	"context"
	"database/sql"
	"fmt"
	"jjr-tec-backend/internal/modals"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// CreateUser inserts a new user into Users Table
func (s *service) CreateUser(username string, email string, password string) error {
    query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`

    _, err := s.db.Exec(query, username, email, password)
    if err != nil {
        log.Printf("Error inserting user: %v", err)
        return err
    }
    return nil
}

// CreateRole inserts a new user into Users Table
func (s *service) CreateRole(role_name string) error {
    query := `INSERT INTO roles (role_name) VALUES ($1)`

    _, err := s.db.Exec(query, role_name)
    if err != nil {
        log.Printf("Error inserting role: %v", err)
        return err
    }
    return nil
}

func (s *service) GetUserByUsernameAndPassword(username string, password string) (*modals.User, error) {
    var user modals.User
    query := `SELECT id, username, email, created_at FROM users WHERE username = $1`
    err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
    if err == sql.ErrNoRows {
        return nil, nil // User not found
    } else if err != nil {
        return nil, err
    }
    return &user, nil
}

func (s *service) GetRolesByUsername(username string, password string) ([]string, error) {
    var roles []string
    query := `
        SELECT r.role_name
        FROM roles r
        INNER JOIN user_roles ur ON ur.role_id = r.id
        INNER JOIN users u ON u.id = ur.user_id
        WHERE u.username = $1
    `
    rows, err := s.db.Query(query, username)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Collect roles into the slice
    for rows.Next() {
        var role string
        if err := rows.Scan(&role); err != nil {
            return nil, err
        }
        roles = append(roles, role)
    }

    // Check for any errors encountered during iteration
    if err = rows.Err(); err != nil {
        return nil, err
    }

    return roles, nil
}


type Service interface {
    // Health returns a map of health status information.
    // The keys and values in the map are service-specific.
    Health() map[string]string

    // Close terminates the database connection.
    // It returns an error if the connection cannot be closed.
    Close() error

    // Creates a User in the Postgres DB, Table Users
    CreateUser(username string, email string, password string) error
    GetUserByUsernameAndPassword(username string, password string) (*modals.User, error)

    // Creates a Role in Postgres DB, Table Roles
    CreateRole(role_name string) error
    GetRolesByUsername(username string, password string) ([]string, error)
}

type service struct {
    db *sql.DB
}

var (
    database   = os.Getenv("BLUEPRINT_DB_DATABASE")
    password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
    username   = os.Getenv("BLUEPRINT_DB_USERNAME")
    port       = os.Getenv("BLUEPRINT_DB_PORT")
    host       = os.Getenv("BLUEPRINT_DB_HOST")
    schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
    dbInstance *service
)

func New() Service {
    // Reuse Connection
    if dbInstance != nil {
        return dbInstance
    }
    connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
    db, err := sql.Open("pgx", connStr)
    if err != nil {
        log.Fatal(err)
    }
    dbInstance = &service{
        db: db,
    }
    return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    stats := make(map[string]string)

    // Ping the database
    err := s.db.PingContext(ctx)
    if err != nil {
        stats["status"] = "down"
        stats["error"] = fmt.Sprintf("db down: %v", err)
        log.Fatalf("db down: %v", err) // Log the error and terminate the program
        return stats
    }

    // Database is up, add more statistics
    stats["status"] = "up"
    stats["message"] = "It's healthy"

    // Get database stats (like open connections, in use, idle, etc.)
    dbStats := s.db.Stats()
    stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
    stats["in_use"] = strconv.Itoa(dbStats.InUse)
    stats["idle"] = strconv.Itoa(dbStats.Idle)
    stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
    stats["wait_duration"] = dbStats.WaitDuration.String()
    stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
    stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

    // Evaluate stats to provide a health message
    if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
        stats["message"] = "The database is experiencing heavy load."
    }

    if dbStats.WaitCount > 1000 {
        stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
    }

    if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
        stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
    }

    if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
        stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
    }

    return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
    log.Printf("Disconnected from database: %s", database)
    return s.db.Close()
}
