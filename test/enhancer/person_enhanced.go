// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package enhancer

type PersonEntityFields struct {
	FirstName string
	LastName  string
	Email     string
}

type PersonTableColumns struct {
	FirstName string
	LastName  string
	Email     string
}

func (e *Person) TableName() string {
	return "person"
}

func (e *Person) EntityFields() *PersonEntityFields {
	return &PersonEntityFields{
		FirstName: "FirstName",
		LastName:  "LastName",
		Email:     "Email",
	}
}

func (e *Person) TableColumns() *PersonTableColumns {
	return &PersonTableColumns{
		FirstName: "first_name",
		LastName:  "last_name",
		Email:     "email",
	}
}

type PersonWithUpdateTracker struct {
	Person
	trackMap map[string]bool
}

func (e *PersonWithUpdateTracker) ColumnsChanged() []string {
	cols := []string{}

	for col, _ := range e.trackMap {
		cols = append(cols, col)
	}

	return cols
}

func (e *PersonWithUpdateTracker) SetFirstName(val string) *PersonWithUpdateTracker {
	e.FirstName = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["first_name"] = true

	return e
}

func (e *PersonWithUpdateTracker) SetLastName(val string) *PersonWithUpdateTracker {
	e.LastName = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["last_name"] = true

	return e
}

func (e *PersonWithUpdateTracker) SetEmail(val string) *PersonWithUpdateTracker {
	e.Email = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["email"] = true

	return e
}
