package handlers

// Transaction represents a row in the PostgreSQL 'transaction' table.
type Transaction struct {
	ID     int     `json:"id"`
	Amount float64 `json:"amount"`  // Use float64 for monetary values in Go
	UserID int     `json:"user_id"` // Foreign key to users table
}
