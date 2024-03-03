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
	trackMap map[string]bool
}

func (e *PersonWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	for col, _ := range e.trackMap {
		cols = append(cols, col)
	}

	return cols
}

func (e *PersonWithUpdateTracker) SetId(val int) *PersonWithUpdateTracker {
	e.Id = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["id"] = true

	return e
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

func (e *PersonWithUpdateTracker) SetEmail(val *string) *PersonWithUpdateTracker {
	e.Email = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["email"] = true

	return e
}

func (e *PersonWithUpdateTracker) SetAge(val *int) *PersonWithUpdateTracker {
	e.Age = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["age"] = true

	return e
}

func (e *PersonWithUpdateTracker) SetCurrentMood(val *string) *PersonWithUpdateTracker {
	e.CurrentMood = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["current_mood"] = true

	return e
}

func (e *PersonWithUpdateTracker) SetAddedAt(val *time.Time) *PersonWithUpdateTracker {
	e.AddedAt = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["added_at"] = true

	return e
}
