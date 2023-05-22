package accessor

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

type AccessorTestSuite struct {
	suite.Suite
	Logger *zap.Logger

	Db *sqlx.DB
}

func TestAccessor(t *testing.T) {
	suite.Run(t, new(AccessorTestSuite))
}

func (s *AccessorTestSuite) SetupSuite() {
	req := require.New(s.T())

	s.Logger = zap.NewExample()

	var db *sqlx.DB
	var err error

	if db, err = sqlx.Open("sqlite3", ":memory:"); err != nil {
		req.FailNow("setup suite failed: %w", err)
	}

	s.Db = db
	s.setupTestDatabase()
}

func (s *AccessorTestSuite) TearDownSuite() {
	if s.Db != nil {
		s.teardownTestDatabase()

		if err := s.Db.Close(); err != nil {
			s.Logger.Error("close db connection: %w", zap.Error(err))
		}
	}
}

func (s *AccessorTestSuite) setupTestDatabase() {
	_ = s.Db.MustExec(`
CREATE TABLE IF NOT EXISTS person (
    first_name text,
    last_name text,
    email text,
    added_at timestamp
);
    `)

	_ = s.Db.MustExec(`
CREATE TABLE IF NOT EXISTS city (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name text,
    zip_code text
);
    `)

	_ = s.Db.MustExec(`DELETE FROM person`)
	_ = s.Db.MustExec(`DELETE FROM city`)

	// we need explicit binding for ? in query when using sqlx functions
	r := s.Db.Rebind
	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, added_at) VALUES (?, ?, ?, ?)    
    `), "foo", "test", "foo@test", time.Now().UTC())

	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, added_at) VALUES (?, ?, ?, ?)    
    `), "bar", "test", "bar@test", time.Now().UTC())
}

func (s *AccessorTestSuite) teardownTestDatabase() {
	assert := require.New(s.T())
	accessor := New(s.Db)

	result, err := accessor.Exec(context.Background(), "delete from person where first_name=?", "foo")
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	result, err = accessor.NamedExec(context.Background(), "delete from person where first_name=:first_name",
		map[string]any{
			"first_name": "bar",
		})
	assert.True(err == nil)
	affected, err = result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS person;
    `)

	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS city;
    `)
}

func (s *AccessorTestSuite) TestGet() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
		"foo", "test")

	assert.True(err == nil)
	assert.True(p.FirstName == "foo")
	assert.True(p.LastName == "test")
	assert.True(p.Email == "foo@test")
	assert.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestNamedGet() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	err := accessor.NamedGet(context.Background(), &p, "select * from person where first_name=:first_name and last_name=:last_name",
		map[string]any{
			"first_name": "foo",
			"last_name":  "test",
		})

	assert.True(err == nil)
	assert.True(p.FirstName == "foo")
	assert.True(p.LastName == "test")
	assert.True(p.Email == "foo@test")
	assert.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestNamedGetByExample() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p, example struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	example.FirstName = "foo"
	example.LastName = "test"
	err := accessor.NamedGet(context.Background(), &p, "select * from person where first_name=:first_name and last_name=:last_name", &example)
	assert.True(err == nil)
	assert.True(p.FirstName == "foo")
	assert.True(p.LastName == "test")
	assert.True(p.Email == "foo@test")
	assert.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestSelect() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p []struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}

	err := accessor.Select(context.Background(), &p, "select * from person where last_name=?", "test")
	assert.True(err == nil)
	assert.True(len(p) == 2)
	assert.True(p[0].FirstName == "foo")
	assert.True(p[0].LastName == "test")
	assert.True(p[0].Email == "foo@test")
	assert.True(p[1].FirstName == "bar")
	assert.True(p[1].LastName == "test")
	assert.True(p[1].Email == "bar@test")

	var pp []*struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	err = accessor.Select(context.Background(), &pp, "select * from person where last_name=?", "test")
	assert.True(err == nil)
	assert.True(len(pp) == 2)
	assert.True(pp[0].FirstName == "foo")
	assert.True(pp[0].LastName == "test")
	assert.True(pp[0].Email == "foo@test")
	assert.True(pp[1].FirstName == "bar")
	assert.True(pp[1].LastName == "test")
	assert.True(pp[1].Email == "bar@test")
}

func (s *AccessorTestSuite) TestNamedSelect() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p []struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}

	err := accessor.NamedSelect(context.Background(), &p, "select * from person where last_name=:last_name order by first_name DESC",
		map[string]any{
			"last_name": "test",
		})
	assert.True(err == nil)
	assert.True(len(p) == 2)
	assert.True(p[0].FirstName == "foo")
	assert.True(p[0].LastName == "test")
	assert.True(p[0].Email == "foo@test")
	assert.True(p[1].FirstName == "bar")
	assert.True(p[1].LastName == "test")
	assert.True(p[1].Email == "bar@test")

	var pp []*struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	err = accessor.NamedSelect(context.Background(), &pp, "select * from person where last_name=:last_name order by first_name DESC",
		map[string]any{
			"last_name": "test",
		})
	assert.True(err == nil)
	assert.True(len(pp) == 2)
	assert.True(pp[0].FirstName == "foo")
	assert.True(pp[0].LastName == "test")
	assert.True(pp[0].Email == "foo@test")
	assert.True(pp[1].FirstName == "bar")
	assert.True(pp[1].LastName == "test")
	assert.True(pp[1].Email == "bar@test")
}

func (s *AccessorTestSuite) TestExecTx() {
	assert := require.New(s.T())

	err := ExecTx(context.Background(), s.Db, &sql.TxOptions{}, func(ctx context.Context, accessor *Accessor) error {
		var p struct {
			FirstName string    `db:"first_name"`
			LastName  string    `db:"last_name"`
			Email     string    `db:"email"`
			AddedAt   time.Time `db:"added_at"`
		}

		err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
			"foo", "test")

		assert.True(err == nil)
		assert.True(p.FirstName == "foo")
		assert.True(p.LastName == "test")
		assert.True(p.Email == "foo@test")
		assert.True(!p.AddedAt.IsZero())

		return nil
	})

	assert.True(err == nil)
}

func (s *AccessorTestSuite) TestExecTxWithPanic() {
	assert := require.New(s.T())

	err := ExecTx(context.Background(), s.Db, &sql.TxOptions{}, func(ctx context.Context, accessor *Accessor) error {
		panic("panic")
	})

	assert.True(err != nil)
	assert.True(err.Error() == "panic error in executing transaction")
}

func (s *AccessorTestSuite) TestSqlizerGet() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	err := accessor.SqlizerGet(context.Background(), &p, func(builder squirrel.StatementBuilderType) Sqlizer {
		return builder.Select("*").From("person").Where(squirrel.Eq{
			Column(p, "FirstName"): "foo",
			Column(p, "LastName"):  "test",
		})
	})

	assert.True(err == nil)
	assert.True(p.FirstName == "foo")
	assert.True(p.LastName == "test")
	assert.True(p.Email == "foo@test")
	assert.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestSqlizerSelect() {
	assert := require.New(s.T())

	accessor := New(s.Db)

	var p []struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	err := accessor.SqlizerSelect(context.Background(), &p, func(builder squirrel.StatementBuilderType) Sqlizer {
		return builder.Select("*").From("person").Where(squirrel.Eq{
			Column(p, "LastName"): "test",
		})
	})

	assert.True(err == nil)
	assert.True(len(p) == 2)
}

type Foo struct {
	FirstName string `db:"first_name"`
}

type Bar struct {
	Foo
	LastName string `db:"last_name"`
}

type Baz struct {
	*Bar
	Email   string    `db:"email"`
	AddedAt time.Time `db:"added_at"`
}

func (s *AccessorTestSuite) TestColumnMapping() {
	assert := require.New(s.T())

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	assert.True(Column(p, "FirstName") == "first_name")

	cols := Columns(p)
	assert.True(len(cols) == 4)
	assert.True(cols[0] == "first_name")
	assert.True(cols[1] == "last_name")
	assert.True(cols[2] == "email")
	assert.True(cols[3] == "added_at")

	var pp *struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	assert.True(Column(pp, "FirstName") == "first_name")

	cols = Columns(pp)
	assert.True(len(cols) == 4)
	assert.True(cols[0] == "first_name")
	assert.True(cols[1] == "last_name")
	assert.True(cols[2] == "email")
	assert.True(cols[3] == "added_at")

	var ppp Baz
	assert.True(Column(ppp, "FirstName") == "first_name")
	assert.True(Column(ppp, "LastName") == "last_name")
	assert.True(Column(ppp, "Email") == "email")
	assert.True(Column(ppp, "AddedAt") == "added_at")

	cols = Columns(ppp)
	assert.True(len(cols) == 4)
	assert.True(cols[0] == "first_name")
	assert.True(cols[1] == "last_name")
	assert.True(cols[2] == "email")
	assert.True(cols[3] == "added_at")
}

func (s *AccessorTestSuite) TestDelete() {
	assert := require.New(s.T())

	p := struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}{
		FirstName: "foo",
		LastName:  "test",
	}

	accessor := New(s.Db)
	result, err := accessor.Delete(context.Background(), p, "person", "FirstName", "LastName")
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	p.FirstName = "bar"
	result, err = accessor.Delete(context.Background(), &p, "person", "FirstName", "LastName")
	assert.True(err == nil)
	affected, err = result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	// restore
	s.setupTestDatabase()
}

type Person struct {
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	AddedAt   time.Time `db:"added_at"`
}

// this type will be generated with go:generate later
type PersonWithUpdateTracker struct {
	Person
	trackMap map[string]bool
}

func (p *PersonWithUpdateTracker) SetEmail(email string) *PersonWithUpdateTracker {
	p.Email = email

	col := Column(p, "Email")
	if col != "" {
		if p.trackMap == nil {
			p.trackMap = make(map[string]bool)
		}
		p.trackMap[col] = true
	}
	return p
}

func (p *PersonWithUpdateTracker) ColumnsChanged() []string {
	var colsChanged []string

	if p.trackMap != nil {
		for k := range p.trackMap {
			colsChanged = append(colsChanged, k)
		}
	}

	return colsChanged
}

func (s *AccessorTestSuite) TestUpdate() {
	assert := require.New(s.T())

	// partial update
	p := &PersonWithUpdateTracker{}
	p.FirstName = "foo"
	p.LastName = "test"

	p.SetEmail("foo@test.change")

	accessor := New(s.Db)
	result, err := accessor.Update(context.Background(), p, "person", "FirstName", "LastName")
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	pp := Person{}
	err = accessor.Get(context.Background(), &pp, "select * from person where first_name=? and last_name=?", "foo", "test")
	assert.True(err == nil)
	assert.True(pp.Email == "foo@test.change")

	// update in full
	ppp := &Person{}
	ppp.FirstName = "foo"
	ppp.LastName = "test"
	ppp.Email = "foo@test.change.again"
	ppp.AddedAt = time.Now().UTC()
	result, err = accessor.Update(context.Background(), ppp, "person", "FirstName", "LastName")
	assert.True(err == nil)
	affected, err = result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	pp = Person{}
	err = accessor.Get(context.Background(), &pp, "select * from person where first_name=? and last_name=?", "foo", "test")
	assert.True(err == nil)
	assert.True(pp.Email == "foo@test.change.again")

	// restore
	s.setupTestDatabase()
}

func (s *AccessorTestSuite) TestCreate() {
	assert := require.New(s.T())

	// create
	city := struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		ZipCode string `db:"zip_code"`
	}{
		Name:    "San Jose",
		ZipCode: "95120",
	}

	accessor := New(s.Db)
	err := accessor.Create(context.Background(), &city, "city")

	assert.True(err == nil)
	assert.True(city.Id != 0)

	// read
	city2 := struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		ZipCode string `db:"zip_code"`
	}{
		Id: city.Id,
	}

	err = accessor.Read(context.Background(), &city2, "city")
	assert.True(err == nil)
	assert.True(city2.Name == "San Jose")
	assert.True(city2.ZipCode == "95120")

	// delete
	result, err := accessor.Delete(context.Background(), &city2, "city")
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)
}

func (s *AccessorTestSuite) TestNotFoundError() {
	assert := require.New(s.T())

	city := struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		ZipCode string `db:"zip_code"`
	}{}

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &city, "SELECT * FROM city WHERE id=?", 100)
	assert.True(err != nil)
	assert.True(err == sql.ErrNoRows)
}

func (s *AccessorTestSuite) TestGetSingleColumn() {
	assert := require.New(s.T())

	email := ""

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &email, "SELECT email FROM person WHERE first_name=?", "foo")
	assert.True(err == nil)
	assert.True(email == "foo@test")
}

func (s *AccessorTestSuite) TestGetSilentDrop() {
	assert := require.New(s.T())

	p := struct {
		Email string `db:"email"`
	}{}

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &p, "SELECT * FROM person WHERE first_name=?", "foo")
	assert.True(err == nil)
	assert.True(p.Email == "foo@test")
}
