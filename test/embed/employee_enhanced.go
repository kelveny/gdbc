// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

import "time"

type EmployeeEntityFields struct {
	Company string
}

type EmployeeTableColumns struct {
	Company string
}

func (e *Employee) TableName() string {
	return "employee"
}

func (e *Employee) EntityFields() *EmployeeEntityFields {
	return &EmployeeEntityFields{
		Company: "Company",
	}
}

func (e *Employee) TableColumns() *EmployeeTableColumns {
	return &EmployeeTableColumns{
		Company: "company",
	}
}

type EmployeeWithUpdateTracker struct {
	Employee
	trackMap map[string]map[string]bool
}

func (e *EmployeeWithUpdateTracker) registerChange(tbl string, col string) {
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

func (e *EmployeeWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	if tbl == nil {
		tbl = []string{"employee"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

	return cols
}

func (e *EmployeeWithUpdateTracker) SetFirstName(val string) *EmployeeWithUpdateTracker {
	e.FirstName = val
	e.registerChange("person", "first_name")
	return e
}

func (e *EmployeeWithUpdateTracker) SetLastName(val string) *EmployeeWithUpdateTracker {
	e.LastName = val
	e.registerChange("person", "last_name")
	return e
}

func (e *EmployeeWithUpdateTracker) SetEmail(val *string) *EmployeeWithUpdateTracker {
	e.Email = val
	e.registerChange("person", "email")
	return e
}

func (e *EmployeeWithUpdateTracker) SetAge(val *int) *EmployeeWithUpdateTracker {
	e.Age = val
	e.registerChange("person", "age")
	return e
}

func (e *EmployeeWithUpdateTracker) SetCurrentMood(val *string) *EmployeeWithUpdateTracker {
	e.CurrentMood = val
	e.registerChange("person", "current_mood")
	return e
}

func (e *EmployeeWithUpdateTracker) SetAddedAt(val *time.Time) *EmployeeWithUpdateTracker {
	e.AddedAt = val
	e.registerChange("person", "added_at")
	return e
}

func (e *EmployeeWithUpdateTracker) SetCompany(val *string) *EmployeeWithUpdateTracker {
	e.Company = val
	e.registerChange("employee", "company")

	return e
}
