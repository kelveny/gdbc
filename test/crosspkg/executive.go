package crosspkg

import "github.com/kelveny/gdbc/test/embed"

//go:generate gdbc -entity Executive -table executive
type Executive struct {
	embed.Manager `db:",table=manager"`
	Term          *string `db:"term"`
}

//go:generate gdbc -entity Executive2 -table executive
type Executive2 struct {
	embed.Manager2 `db:",table=manager"`
	Term           *string `db:"term"`
}

//go:generate gdbc -entity Executive3 -table executive
type Executive3 struct {
	embed.Manager3 `db:",table=manager"`
	Term           *string `db:"term"`
}

//go:generate gdbc -entity Executive4 -table executive
type Executive4 struct {
	embed.Manager4 `db:",table=manager"`
	Term           *string `db:"term"`
}

//go:generate gdbc -entity Executive5 -table executive
type Executive5 struct {
	*embed.Manager `db:",table=manager"`
	Term           *string `db:"term"`
}

//go:generate gdbc -entity Executive6 -table executive
type Executive6 struct {
	*embed.Manager2 `db:",table=manager"`
	Term            *string `db:"term"`
}

//go:generate gdbc -entity Executive7 -table executive
type Executive7 struct {
	*embed.Manager3 `db:",table=manager"`
	Term            *string `db:"term"`
}

//go:generate gdbc -entity Executive8 -table executive
type Executive8 struct {
	*embed.Manager4 `db:",table=manager"`
	Term            *string `db:"term"`
}
