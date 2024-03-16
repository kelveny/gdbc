// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type Manager4EntityFields struct {
	Title string
}

type Manager4TableColumns struct {
	Title string
}

func (e *Manager4) TableName() string {
	return "manager"
}

func (e *Manager4) EntityFields() *Manager4EntityFields {
	return &Manager4EntityFields{
		Title: "Title",
	}
}

func (e *Manager4) TableColumns() *Manager4TableColumns {
	return &Manager4TableColumns{
		Title: "title",
	}
}

type Manager4WithUpdateTracker struct {
	Manager4
	trackMap map[string]map[string]bool
}

func (e *Manager4WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Manager4WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Manager4WithUpdateTracker) SetTitle(val *string) *Manager4WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Manager4WithUpdateTracker) SetCompany(val *string) *Manager4WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Manager4WithUpdateTracker) SetId(val int) *Manager4WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Manager4WithUpdateTracker) SetFirstName(val string) *Manager4WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Manager4WithUpdateTracker) SetLastName(val string) *Manager4WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Manager4WithUpdateTracker) SetEmail(val *string) *Manager4WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Manager4WithUpdateTracker) SetAge(val *int) *Manager4WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Manager4WithUpdateTracker) SetCurrentMood(val *string) *Manager4WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Manager4WithUpdateTracker) SetAddedAt(val *time.Time) *Manager4WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
