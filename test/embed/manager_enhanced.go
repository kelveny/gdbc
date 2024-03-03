// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package embed

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
	trackMap map[string]bool
}

func (e *ManagerWithUpdateTracker) ColumnsChanged(tbl ...string) []string {
	cols := []string{}

	for col, _ := range e.trackMap {
		cols = append(cols, col)
	}

	return cols
}

func (e *ManagerWithUpdateTracker) SetTitle(val *string) *ManagerWithUpdateTracker {
	e.Title = val

	if e.trackMap == nil {
		e.trackMap = make(map[string]bool)
	}

	e.trackMap["title"] = true

	return e
}
