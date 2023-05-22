package enhancer

//go:generate gdbc -entity Person -table person
type Person struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}
