package embed

//go:generate gdbc -entity Employee -table employee
type Employee struct {
	Person `db:",table=person"`

	Company *string `db:"company"`
}
