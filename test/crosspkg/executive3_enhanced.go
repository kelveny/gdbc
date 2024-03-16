// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import (
	"time"
)

type Executive3EntityFields struct {
	Term string
}

type Executive3TableColumns struct {
	Term string
}

func (e *Executive3) TableName() string {
	return "executive"
}

func (e *Executive3) EntityFields() *Executive3EntityFields {
	return &Executive3EntityFields{
		Term: "Term",
	}
}

func (e *Executive3) TableColumns() *Executive3TableColumns {
	return &Executive3TableColumns{
		Term: "term",
	}
}

type Executive3WithUpdateTracker struct {
	Executive3
	trackMap map[string]map[string]bool
}

func (e *Executive3WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Executive3WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Executive3WithUpdateTracker) SetTerm(val *string) *Executive3WithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *Executive3WithUpdateTracker) SetTitle(val *string) *Executive3WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Executive3WithUpdateTracker) SetCompany(val *string) *Executive3WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Executive3WithUpdateTracker) SetId(val int) *Executive3WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Executive3WithUpdateTracker) SetFirstName(val string) *Executive3WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Executive3WithUpdateTracker) SetLastName(val string) *Executive3WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Executive3WithUpdateTracker) SetEmail(val *string) *Executive3WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Executive3WithUpdateTracker) SetAge(val *int) *Executive3WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Executive3WithUpdateTracker) SetCurrentMood(val *string) *Executive3WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Executive3WithUpdateTracker) SetAddedAt(val *time.Time) *Executive3WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
