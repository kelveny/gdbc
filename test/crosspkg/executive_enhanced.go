// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package crosspkg

import "time"

type ExecutiveEntityFields struct {
	Term string
}

type ExecutiveTableColumns struct {
	Term string
}

func (e *Executive) TableName() string {
	return "executive"
}

func (e *Executive) EntityFields() *ExecutiveEntityFields {
	return &ExecutiveEntityFields{
		Term: "Term",
	}
}

func (e *Executive) TableColumns() *ExecutiveTableColumns {
	return &ExecutiveTableColumns{
		Term: "term",
	}
}

type ExecutiveWithUpdateTracker struct {
	Executive
	trackMap map[string]map[string]bool
}

func (e *ExecutiveWithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *ExecutiveWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
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

func (e *ExecutiveWithUpdateTracker) SetTerm(val *string) *ExecutiveWithUpdateTracker {
	e.Term = val
	e.registerChange("executive", "term")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetTitle(val *string) *ExecutiveWithUpdateTracker {
	e.Title = val
	e.registerChange("manager", "title")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetCompany(val *string) *ExecutiveWithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetId(val int) *ExecutiveWithUpdateTracker {
	e.Id = val
	e.registerChange("person", "id")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetFirstName(val string) *ExecutiveWithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetLastName(val string) *ExecutiveWithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetEmail(val *string) *ExecutiveWithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetAge(val *int) *ExecutiveWithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetCurrentMood(val *string) *ExecutiveWithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *ExecutiveWithUpdateTracker) SetAddedAt(val *time.Time) *ExecutiveWithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}
