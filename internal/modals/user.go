package modals

// User represents an entry in Postgres Table Users
type User struct {
	ID        int
	Username  string
	Email     string
	CreatedAt string
}