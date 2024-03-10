package crosspkg

import "github.com/kelveny/gdbc/test/embed"

//go:generate gdbc -entity Executive -table executive
type Executive struct {
	embed.Manager `db:",table=manager"`
	Term          *string `db:"term"`
}
