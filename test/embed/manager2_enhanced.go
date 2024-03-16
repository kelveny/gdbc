// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type Manager2EntityFields struct {
	Title string
}

type Manager2TableColumns struct {
	Title string
}

func (e *Manager2) TableName() string {
	return "manager"
}

func (e *Manager2) EntityFields() *Manager2EntityFields {
	return &Manager2EntityFields{
		Title: "Title",
	}
}

func (e *Manager2) TableColumns() *Manager2TableColumns {
	return &Manager2TableColumns{
		Title: "title",
	}
}

type Manager2WithUpdateTracker struct {
	Manager2
	trackMap map[string]map[string]bool
}

func (e *Manager2WithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *Manager2WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *Manager2WithUpdateTracker) SetTitle(val *string) *Manager2WithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *Manager2WithUpdateTracker) SetCompany(val *string) *Manager2WithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *Manager2WithUpdateTracker) SetId(val int) *Manager2WithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *Manager2WithUpdateTracker) SetFirstName(val string) *Manager2WithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *Manager2WithUpdateTracker) SetLastName(val string) *Manager2WithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *Manager2WithUpdateTracker) SetEmail(val *string) *Manager2WithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *Manager2WithUpdateTracker) SetAge(val *int) *Manager2WithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *Manager2WithUpdateTracker) SetCurrentMood(val *string) *Manager2WithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *Manager2WithUpdateTracker) SetAddedAt(val *time.Time) *Manager2WithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
