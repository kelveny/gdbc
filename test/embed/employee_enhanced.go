// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

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
	trackMap map[string]bool
}

func (e *EmployeeWithUpdateTracker) ColumnsChanged() []string {
	cols := []string{}

	for col, _ := range e.trackMap {
		cols = append(cols, col)
	}

	return cols
}

func (e *EmployeeWithUpdateTracker) SetCompany(val *string) *EmployeeWithUpdateTracker {
	e.Company = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["company"] = true

	return e
}
