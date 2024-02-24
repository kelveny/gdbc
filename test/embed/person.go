package embed

import "time"

//go:generate gdbc -entity Person -table person
type Person struct {
	Id          int        `db:"id"`
	FirstName   string     `db:"first_name"`
	LastName    string     `db:"last_name"`
	Email       *string    `db:"email"`
	Age         *int       `db:"age"`
	CurrentMood *string    `db:"current_mood"`
	AddedAt     *time.Time `db:"added_at"`
}
