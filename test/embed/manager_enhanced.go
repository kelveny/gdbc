// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import (
	"time"
)

type ManagerEntityFields struct {
	Title string
}

type ManagerTableColumns struct {
	Title string
}

func (e *Manager) TableName() string {
	return "manager"
}

func (e *Manager) EntityFields() *ManagerEntityFields {
	return &ManagerEntityFields{
		Title: "Title",
	}
}

func (e *Manager) TableColumns() *ManagerTableColumns {
	return &ManagerTableColumns{
		Title: "title",
	}
}

type ManagerWithUpdateTracker struct {
	Manager
	trackMap map[string]map[string]bool
}

func (e *ManagerWithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *ManagerWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *ManagerWithUpdateTracker) SetTitle(val *string) *ManagerWithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *ManagerWithUpdateTracker) SetCompany(val *string) *ManagerWithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *ManagerWithUpdateTracker) SetId(val int) *ManagerWithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *ManagerWithUpdateTracker) SetFirstName(val string) *ManagerWithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *ManagerWithUpdateTracker) SetLastName(val string) *ManagerWithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *ManagerWithUpdateTracker) SetEmail(val *string) *ManagerWithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *ManagerWithUpdateTracker) SetAge(val *int) *ManagerWithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *ManagerWithUpdateTracker) SetCurrentMood(val *string) *ManagerWithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *ManagerWithUpdateTracker) SetAddedAt(val *time.Time) *ManagerWithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
