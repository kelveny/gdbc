// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type Employee2EntityFields struct {
	Company string
}

type Employee2TableColumns struct {
	Company string
}

func (e *Employee2) TableName() string {
	return "employee"
}

func (e *Employee2) EntityFields() *Employee2EntityFields {
	return &Employee2EntityFields{
		Company: "Company",
	}
}

func (e *Employee2) TableColumns() *Employee2TableColumns {
	return &Employee2TableColumns{
		Company: "company",
	}
}

type Employee2WithUpdateTracker struct {
	Employee2
	trackMap map[string]map[string]bool
}

func (e *Employee2WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Employee2WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	if tbl == nil {
		tbl = []string{"employee"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

	return cols
}

func (e *Employee2WithUpdateTracker) SetCompany(val *string) *Employee2WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Employee2WithUpdateTracker) SetId(val int) *Employee2WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Employee2WithUpdateTracker) SetFirstName(val string) *Employee2WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Employee2WithUpdateTracker) SetLastName(val string) *Employee2WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Employee2WithUpdateTracker) SetEmail(val *string) *Employee2WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Employee2WithUpdateTracker) SetAge(val *int) *Employee2WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Employee2WithUpdateTracker) SetCurrentMood(val *string) *Employee2WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Employee2WithUpdateTracker) SetAddedAt(val *time.Time) *Employee2WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
