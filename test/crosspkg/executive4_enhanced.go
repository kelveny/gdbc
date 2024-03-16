// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive4EntityFields struct {
	Term string
}

type Executive4TableColumns struct {
	Term string
}

func (e *Executive4) TableName() string {
	return "executive"
}

func (e *Executive4) EntityFields() *Executive4EntityFields {
	return &Executive4EntityFields{
		Term: "Term",
	}
}

func (e *Executive4) TableColumns() *Executive4TableColumns {
	return &Executive4TableColumns{
		Term: "term",
	}
}

type Executive4WithUpdateTracker struct {
	Executive4
	trackMap map[string]map[string]bool
}

func (e *Executive4WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive4WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive4WithUpdateTracker) SetTerm(val *string) *Executive4WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive4WithUpdateTracker) SetTitle(val *string) *Executive4WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive4WithUpdateTracker) SetCompany(val *string) *Executive4WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive4WithUpdateTracker) SetId(val int) *Executive4WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive4WithUpdateTracker) SetFirstName(val string) *Executive4WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive4WithUpdateTracker) SetLastName(val string) *Executive4WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive4WithUpdateTracker) SetEmail(val *string) *Executive4WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive4WithUpdateTracker) SetAge(val *int) *Executive4WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive4WithUpdateTracker) SetCurrentMood(val *string) *Executive4WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive4WithUpdateTracker) SetAddedAt(val *time.Time) *Executive4WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
