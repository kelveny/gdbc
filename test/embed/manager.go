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

//go:generate gdbc -entity Manager2 -table manager

type Manager2 struct {
	*Employee `db:",table=employee"`

	Title *string `db:"title"`
}

//go:generate gdbc -entity Manager3 -table manager

type Manager3 struct {
	Employee2 `db:",table=employee"`

	Title *string `db:"title"`
}

//go:generate gdbc -entity Manager4 -table manager
type Manager4 struct {
	*Employee2 `db:",table=employee"`

	Title *string `db:"title"`
}
