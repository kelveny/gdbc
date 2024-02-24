package accessor

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Columns(t *testing.T) {
	req := require.New(t)

	p := struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}{
		FirstName: "foo",
		LastName:  "test",
	}

	cols := Columns(p)

	req.True(len(cols) == 2)

	colMap := map[string]bool{
		"first_name": true,
		"last_name":  true,
	}

	for _, col := range cols {
		_, ok := colMap[col]
		req.Equal(true, ok, "%s should appear", col)
	}
}

func Test_ColumnWithAttributes(t *testing.T) {
	req := require.New(t)

	p := struct {
		FirstName string `db:"first_name, table=person, readonly"`
		LastName  string `db:"last_name"`
	}{
		FirstName: "foo",
		LastName:  "test",
	}

	typ := reflect.TypeOf(p)

	field, ok := typ.FieldByName("FirstName")
	req.True(ok)

	// happy cases
	col, attrs, err := fieldMappedColumnWithAttributes(field, "db")
	req.NoError(err)
	req.True(col == "first_name")

	req.True(len(attrs) == 2)
	req.True(attrs["table"] == "person")

	val, ok := attrs["readonly"]
	req.True(ok)
	req.True(val == "")

	// failure cases
	_, ok = attrs["non-exist"]
	req.True(!ok)

	col, attrs, err = fieldMappedColumnWithAttributes(field, "non-exist")
	req.True(err != nil)
	req.True(col == "")
	req.True(len(attrs) == 0)
}
