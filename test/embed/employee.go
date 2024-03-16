package embed

//go:generate gdbc -entity Employee -table employee
type Employee struct {
	Person `db:",table=person"`

	Company *string `db:"company"`
}

//go:generate gdbc -entity Employee2 -table employee
type Employee2 struct {
	*Person `db:",table=person"`

	Company *string `db:"company"`
}
