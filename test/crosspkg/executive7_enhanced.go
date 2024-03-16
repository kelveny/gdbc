// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive7EntityFields struct {
	Term string
}

type Executive7TableColumns struct {
	Term string
}

func (e *Executive7) TableName() string {
	return "executive"
}

func (e *Executive7) EntityFields() *Executive7EntityFields {
	return &Executive7EntityFields{
		Term: "Term",
	}
}

func (e *Executive7) TableColumns() *Executive7TableColumns {
	return &Executive7TableColumns{
		Term: "term",
	}
}

type Executive7WithUpdateTracker struct {
	Executive7
	trackMap map[string]map[string]bool
}

func (e *Executive7WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive7WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive7WithUpdateTracker) SetTerm(val *string) *Executive7WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive7WithUpdateTracker) SetTitle(val *string) *Executive7WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive7WithUpdateTracker) SetCompany(val *string) *Executive7WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive7WithUpdateTracker) SetId(val int) *Executive7WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive7WithUpdateTracker) SetFirstName(val string) *Executive7WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive7WithUpdateTracker) SetLastName(val string) *Executive7WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive7WithUpdateTracker) SetEmail(val *string) *Executive7WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive7WithUpdateTracker) SetAge(val *int) *Executive7WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive7WithUpdateTracker) SetCurrentMood(val *string) *Executive7WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive7WithUpdateTracker) SetAddedAt(val *time.Time) *Executive7WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
