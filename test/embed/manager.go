package embed

//go:generate gdbc -entity Manager -table manager
type Manager struct {
	Employee `db:",table=employee"`

	Title *string `db:"title"`
}

type ManagerWrapper struct {
	Manager
}

type ManagerWrapperWrapper struct {
	ManagerWrapper
}
