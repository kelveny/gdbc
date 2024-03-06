// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type PersonEntityFields struct {
	Id          string
	FirstName   string
	LastName    string
	Email       string
	Age         string
	CurrentMood string
	AddedAt     string
}

type PersonTableColumns struct {
	Id          string
	FirstName   string
	LastName    string
	Email       string
	Age         string
	CurrentMood string
	AddedAt     string
}

func (e *Person) TableName() string {
	return "person"
}

func (e *Person) EntityFields() *PersonEntityFields {
	return &PersonEntityFields{
		Id:          "Id",
		FirstName:   "FirstName",
		LastName:    "LastName",
		Email:       "Email",
		Age:         "Age",
		CurrentMood: "CurrentMood",
		AddedAt:     "AddedAt",
	}
}

func (e *Person) TableColumns() *PersonTableColumns {
	return &PersonTableColumns{
		Id:          "id",
		FirstName:   "first_name",
		LastName:    "last_name",
		Email:       "email",
		Age:         "age",
		CurrentMood: "current_mood",
		AddedAt:     "added_at",
	}
}

type PersonWithUpdateTracker struct {
	Person
	trackMap map[string]map[string]bool
}

func (e *PersonWithUpdateTracker) registerChange(tbl string, col string) {
	if e.trackMap == nil {
		e.trackMap = make(map[string]map[string]bool)
	}

	if m, ok := e.trackMap[tbl]; ok {
		m[col] = true
	} else {
		m = make(map[string]bool)
		e.trackMap[tbl] = m

		m[col] = true
	}
}

func (e *PersonWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	if tbl == nil {
		tbl = []string{"person"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

	return cols
}

func (e *PersonWithUpdateTracker) SetId(val int) *PersonWithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *PersonWithUpdateTracker) SetFirstName(val string) *PersonWithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *PersonWithUpdateTracker) SetLastName(val string) *PersonWithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *PersonWithUpdateTracker) SetEmail(val *string) *PersonWithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *PersonWithUpdateTracker) SetAge(val *int) *PersonWithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *PersonWithUpdateTracker) SetCurrentMood(val *string) *PersonWithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *PersonWithUpdateTracker) SetAddedAt(val *time.Time) *PersonWithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
