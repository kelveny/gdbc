// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type Manager3EntityFields struct {
	Title string
}

type Manager3TableColumns struct {
	Title string
}

func (e *Manager3) TableName() string {
	return "manager"
}

func (e *Manager3) EntityFields() *Manager3EntityFields {
	return &Manager3EntityFields{
		Title: "Title",
	}
}

func (e *Manager3) TableColumns() *Manager3TableColumns {
	return &Manager3TableColumns{
		Title: "title",
	}
}

type Manager3WithUpdateTracker struct {
	Manager3
	trackMap map[string]map[string]bool
}

func (e *Manager3WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Manager3WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	if tbl == nil {
		tbl = []string{"manager"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

	return cols
}

func (e *Manager3WithUpdateTracker) SetTitle(val *string) *Manager3WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Manager3WithUpdateTracker) SetCompany(val *string) *Manager3WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Manager3WithUpdateTracker) SetId(val int) *Manager3WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Manager3WithUpdateTracker) SetFirstName(val string) *Manager3WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Manager3WithUpdateTracker) SetLastName(val string) *Manager3WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Manager3WithUpdateTracker) SetEmail(val *string) *Manager3WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Manager3WithUpdateTracker) SetAge(val *int) *Manager3WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Manager3WithUpdateTracker) SetCurrentMood(val *string) *Manager3WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Manager3WithUpdateTracker) SetAddedAt(val *time.Time) *Manager3WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
