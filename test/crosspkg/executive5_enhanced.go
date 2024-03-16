// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive5EntityFields struct {
	Term string
}

type Executive5TableColumns struct {
	Term string
}

func (e *Executive5) TableName() string {
	return "executive"
}

func (e *Executive5) EntityFields() *Executive5EntityFields {
	return &Executive5EntityFields{
		Term: "Term",
	}
}

func (e *Executive5) TableColumns() *Executive5TableColumns {
	return &Executive5TableColumns{
		Term: "term",
	}
}

type Executive5WithUpdateTracker struct {
	Executive5
	trackMap map[string]map[string]bool
}

func (e *Executive5WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive5WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive5WithUpdateTracker) SetTerm(val *string) *Executive5WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive5WithUpdateTracker) SetTitle(val *string) *Executive5WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive5WithUpdateTracker) SetCompany(val *string) *Executive5WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive5WithUpdateTracker) SetId(val int) *Executive5WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive5WithUpdateTracker) SetFirstName(val string) *Executive5WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive5WithUpdateTracker) SetLastName(val string) *Executive5WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive5WithUpdateTracker) SetEmail(val *string) *Executive5WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive5WithUpdateTracker) SetAge(val *int) *Executive5WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive5WithUpdateTracker) SetCurrentMood(val *string) *Executive5WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive5WithUpdateTracker) SetAddedAt(val *time.Time) *Executive5WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
