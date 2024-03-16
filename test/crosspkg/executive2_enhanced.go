// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive2EntityFields struct {
	Term string
}

type Executive2TableColumns struct {
	Term string
}

func (e *Executive2) TableName() string {
	return "executive"
}

func (e *Executive2) EntityFields() *Executive2EntityFields {
	return &Executive2EntityFields{
		Term: "Term",
	}
}

func (e *Executive2) TableColumns() *Executive2TableColumns {
	return &Executive2TableColumns{
		Term: "term",
	}
}

type Executive2WithUpdateTracker struct {
	Executive2
	trackMap map[string]map[string]bool
}

func (e *Executive2WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive2WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	if tbl == nil {
		tbl = []string{"executive"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

	return cols
}

func (e *Executive2WithUpdateTracker) SetTerm(val *string) *Executive2WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive2WithUpdateTracker) SetTitle(val *string) *Executive2WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive2WithUpdateTracker) SetCompany(val *string) *Executive2WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive2WithUpdateTracker) SetId(val int) *Executive2WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive2WithUpdateTracker) SetFirstName(val string) *Executive2WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive2WithUpdateTracker) SetLastName(val string) *Executive2WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive2WithUpdateTracker) SetEmail(val *string) *Executive2WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive2WithUpdateTracker) SetAge(val *int) *Executive2WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive2WithUpdateTracker) SetCurrentMood(val *string) *Executive2WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive2WithUpdateTracker) SetAddedAt(val *time.Time) *Executive2WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
