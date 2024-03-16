// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive8EntityFields struct {
	Term string
}

type Executive8TableColumns struct {
	Term string
}

func (e *Executive8) TableName() string {
	return "executive"
}

func (e *Executive8) EntityFields() *Executive8EntityFields {
	return &Executive8EntityFields{
		Term: "Term",
	}
}

func (e *Executive8) TableColumns() *Executive8TableColumns {
	return &Executive8TableColumns{
		Term: "term",
	}
}

type Executive8WithUpdateTracker struct {
	Executive8
	trackMap map[string]map[string]bool
}

func (e *Executive8WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive8WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive8WithUpdateTracker) SetTerm(val *string) *Executive8WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive8WithUpdateTracker) SetTitle(val *string) *Executive8WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive8WithUpdateTracker) SetCompany(val *string) *Executive8WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive8WithUpdateTracker) SetId(val int) *Executive8WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive8WithUpdateTracker) SetFirstName(val string) *Executive8WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive8WithUpdateTracker) SetLastName(val string) *Executive8WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive8WithUpdateTracker) SetEmail(val *string) *Executive8WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive8WithUpdateTracker) SetAge(val *int) *Executive8WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive8WithUpdateTracker) SetCurrentMood(val *string) *Executive8WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive8WithUpdateTracker) SetAddedAt(val *time.Time) *Executive8WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
