// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive6EntityFields struct {
	Term string
}

type Executive6TableColumns struct {
	Term string
}

func (e *Executive6) TableName() string {
	return "executive"
}

func (e *Executive6) EntityFields() *Executive6EntityFields {
	return &Executive6EntityFields{
		Term: "Term",
	}
}

func (e *Executive6) TableColumns() *Executive6TableColumns {
	return &Executive6TableColumns{
		Term: "term",
	}
}

type Executive6WithUpdateTracker struct {
	Executive6
	trackMap map[string]map[string]bool
}

func (e *Executive6WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive6WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive6WithUpdateTracker) SetTerm(val *string) *Executive6WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive6WithUpdateTracker) SetTitle(val *string) *Executive6WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive6WithUpdateTracker) SetCompany(val *string) *Executive6WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive6WithUpdateTracker) SetId(val int) *Executive6WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive6WithUpdateTracker) SetFirstName(val string) *Executive6WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive6WithUpdateTracker) SetLastName(val string) *Executive6WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive6WithUpdateTracker) SetEmail(val *string) *Executive6WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive6WithUpdateTracker) SetAge(val *int) *Executive6WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive6WithUpdateTracker) SetCurrentMood(val *string) *Executive6WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive6WithUpdateTracker) SetAddedAt(val *time.Time) *Executive6WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
