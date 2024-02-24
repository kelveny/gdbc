package accessor

import (
	"context"
	"database/sql"
	"reflect"
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
	age integer,
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
INSERT INTO person(first_name, last_name, email, age, added_at) VALUES (?, ?, ?, ?, ?)    
    `), "foo", "test", "foo@test", 20, time.Now().UTC())

	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, age, added_at) VALUES (?, ?, ?, ?, ?)    
    `), "bar", "test", "bar@test", 30, time.Now().UTC())
}

func (s *AccessorTestSuite) teardownTestDatabase() {
	req := require.New(s.T())
	accessor := New(s.Db)

	result, err := accessor.Exec(context.Background(), "delete from person where first_name=?", "foo")
	req.NoError(err)
	affected, err := result.RowsAffected()
	req.NoError(err)
	req.True(affected == 1)

	result, err = accessor.NamedExec(context.Background(), "delete from person where first_name=:first_name",
		map[string]any{
			"first_name": "bar",
		})
	req.NoError(err)

	affected, err = result.RowsAffected()
	req.NoError(err)
	req.True(affected == 1)

	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS person;
    `)

	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS city;
    `)
}

func (s *AccessorTestSuite) TestGet() {
	req := require.New(s.T())

	accessor := New(s.Db)

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		Age       *int      `db:"age"`
		AddedAt   time.Time `db:"added_at"`
	}

	err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
		"foo", "test")

	req.NoError(err)
	req.True(p.FirstName == "foo")
	req.True(p.LastName == "test")
	req.True(p.Email == "foo@test")
	req.True(*p.Age == 20)
	req.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestNamedGet() {
	req := require.New(s.T())

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

	req.NoError(err)
	req.True(p.FirstName == "foo")
	req.True(p.LastName == "test")
	req.True(p.Email == "foo@test")
	req.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestNamedGetByExample() {
	req := require.New(s.T())

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
	req.NoError(err)
	req.True(p.FirstName == "foo")
	req.True(p.LastName == "test")
	req.True(p.Email == "foo@test")
	req.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestSelect() {
	req := require.New(s.T())

	accessor := New(s.Db)

	var p []struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}

	err := accessor.Select(context.Background(), &p, "select * from person where last_name=?", "test")
	req.NoError(err)
	req.True(len(p) == 2)
	req.True(p[0].FirstName == "foo")
	req.True(p[0].LastName == "test")
	req.True(p[0].Email == "foo@test")
	req.True(p[1].FirstName == "bar")
	req.True(p[1].LastName == "test")
	req.True(p[1].Email == "bar@test")

	var pp []*struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	err = accessor.Select(context.Background(), &pp, "select * from person where last_name=?", "test")
	req.NoError(err)
	req.True(len(pp) == 2)
	req.True(pp[0].FirstName == "foo")
	req.True(pp[0].LastName == "test")
	req.True(pp[0].Email == "foo@test")
	req.True(pp[1].FirstName == "bar")
	req.True(pp[1].LastName == "test")
	req.True(pp[1].Email == "bar@test")
}

func (s *AccessorTestSuite) TestNamedSelect() {
	req := require.New(s.T())

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
	req.NoError(err)
	req.True(len(p) == 2)
	req.True(p[0].FirstName == "foo")
	req.True(p[0].LastName == "test")
	req.True(p[0].Email == "foo@test")
	req.True(p[1].FirstName == "bar")
	req.True(p[1].LastName == "test")
	req.True(p[1].Email == "bar@test")

	var pp []*struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
	}
	err = accessor.NamedSelect(context.Background(), &pp, "select * from person where last_name=:last_name order by first_name DESC",
		map[string]any{
			"last_name": "test",
		})
	req.NoError(err)
	req.True(len(pp) == 2)
	req.True(pp[0].FirstName == "foo")
	req.True(pp[0].LastName == "test")
	req.True(pp[0].Email == "foo@test")
	req.True(pp[1].FirstName == "bar")
	req.True(pp[1].LastName == "test")
	req.True(pp[1].Email == "bar@test")
}

func (s *AccessorTestSuite) TestExecTx() {
	req := require.New(s.T())

	err := ExecTx(context.Background(), s.Db, &sql.TxOptions{}, func(ctx context.Context, accessor *Accessor) error {
		var p struct {
			FirstName string    `db:"first_name"`
			LastName  string    `db:"last_name"`
			Email     string    `db:"email"`
			AddedAt   time.Time `db:"added_at"`
		}

		err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
			"foo", "test")

		req.NoError(err)
		req.True(p.FirstName == "foo")
		req.True(p.LastName == "test")
		req.True(p.Email == "foo@test")
		req.True(!p.AddedAt.IsZero())

		return nil
	})

	req.True(err == nil)
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
	req := require.New(s.T())

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

	req.NoError(err)
	req.True(p.FirstName == "foo")
	req.True(p.LastName == "test")
	req.True(p.Email == "foo@test")
	req.True(!p.AddedAt.IsZero())
}

func (s *AccessorTestSuite) TestSqlizerSelect() {
	req := require.New(s.T())

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

	req.NoError(err)
	req.True(len(p) == 2)
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
	req := require.New(s.T())

	var p struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	req.True(Column(p, "FirstName") == "first_name")

	cols := Columns(p)
	req.True(len(cols) == 4)
	req.True(cols[0] == "first_name")
	req.True(cols[1] == "last_name")
	req.True(cols[2] == "email")
	req.True(cols[3] == "added_at")

	var pp *struct {
		FirstName string    `db:"first_name"`
		LastName  string    `db:"last_name"`
		Email     string    `db:"email"`
		AddedAt   time.Time `db:"added_at"`
	}

	req.True(Column(pp, "FirstName") == "first_name")

	cols = Columns(pp)
	req.True(len(cols) == 4)
	req.True(cols[0] == "first_name")
	req.True(cols[1] == "last_name")
	req.True(cols[2] == "email")
	req.True(cols[3] == "added_at")

	var ppp Baz
	req.True(Column(ppp, "FirstName") == "first_name")
	req.True(Column(ppp, "LastName") == "last_name")
	req.True(Column(ppp, "Email") == "email")
	req.True(Column(ppp, "AddedAt") == "added_at")

	cols = Columns(ppp)
	req.True(len(cols) == 4)
	req.True(cols[0] == "first_name")
	req.True(cols[1] == "last_name")
	req.True(cols[2] == "email")
	req.True(cols[3] == "added_at")
}

func (s *AccessorTestSuite) TestDelete() {
	req := require.New(s.T())

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
	req.True(err == nil)
	affected, err := result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)
	req.True(p.Email == "")

	p.FirstName = "bar"
	result, err = accessor.Delete(context.Background(), &p, "person", "FirstName", "LastName")
	req.True(err == nil)
	affected, err = result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)
	req.True(p.Email == "bar@test")

	// restore
	s.setupTestDatabase()
}

type Person struct {
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	AddedAt   time.Time `db:"added_at"`
}

// this type can be generated with go:generate tooling. write the original
// here for self-completeness and use as a template for the tool.
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
	req := require.New(s.T())

	// partial update
	p := &PersonWithUpdateTracker{}
	p.FirstName = "foo"
	p.LastName = "test"

	p.SetEmail("foo@test.change")

	accessor := New(s.Db)
	result, err := accessor.Update(context.Background(), p, "person", "FirstName", "LastName")
	req.True(err == nil)
	affected, err := result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)

	pp := Person{}
	err = accessor.Get(context.Background(), &pp, "select * from person where first_name=? and last_name=?", "foo", "test")
	req.True(err == nil)
	req.True(pp.Email == "foo@test.change")

	// update in full
	ppp := &Person{}
	ppp.FirstName = "foo"
	ppp.LastName = "test"
	ppp.Email = "foo@test.change.again"
	ppp.AddedAt = time.Now().UTC()
	result, err = accessor.Update(context.Background(), ppp, "person", "FirstName", "LastName")
	req.True(err == nil)
	affected, err = result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)

	pp = Person{}
	err = accessor.Get(context.Background(), &pp, "select * from person where first_name=? and last_name=?", "foo", "test")
	req.True(err == nil)
	req.True(pp.Email == "foo@test.change.again")

	// restore
	s.setupTestDatabase()
}

func (s *AccessorTestSuite) TestCreateWithImpicitIdMapping() {
	req := require.New(s.T())

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

	req.True(err == nil)

	// assert on auto-increment of ID property
	req.True(city.Id != 0)

	// read
	city2 := struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		ZipCode string `db:"zip_code"`
	}{
		Id: city.Id,
	}

	err = accessor.Read(context.Background(), &city2, "city")
	req.True(err == nil)
	req.True(city2.Name == "San Jose")
	req.True(city2.ZipCode == "95120")

	// delete
	result, err := accessor.Delete(context.Background(), &city2, "city")
	req.True(err == nil)
	affected, err := result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)
}

func (s *AccessorTestSuite) TestCreateWithoutIdMapping() {
	req := require.New(s.T())

	// create
	p := Person{
		FirstName: "foo",
		LastName:  "bar",
		Email:     "foo.bar@test",
	}
	accessor := New(s.Db)
	err := accessor.Create(context.Background(), &p, "person")

	req.True(err == nil)

	// read
	p2 := Person{}

	p2.FirstName = "foo"
	p2.LastName = "bar"
	err = accessor.Read(context.Background(), &p2, "person", "FirstName", "LastName")
	req.True(err == nil)
	req.True(p2.Email == "foo.bar@test")

	// delete
	result, err := accessor.Delete(context.Background(), &p2, "person", "FirstName", "LastName")
	req.True(err == nil)
	affected, err := result.RowsAffected()
	req.True(err == nil)
	req.True(affected == 1)
}

func (s *AccessorTestSuite) TestNotFoundError() {
	req := require.New(s.T())

	city := struct {
		Id      int    `db:"id"`
		Name    string `db:"name"`
		ZipCode string `db:"zip_code"`
	}{}

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &city, "SELECT * FROM city WHERE id=?", 100)
	req.True(err != nil)
	req.True(err == sql.ErrNoRows)
}

func (s *AccessorTestSuite) TestGetSingleColumn() {
	req := require.New(s.T())

	email := ""

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &email, "SELECT email FROM person WHERE first_name=?", "foo")
	req.True(err == nil)
	req.True(email == "foo@test")
}

func (s *AccessorTestSuite) TestGetSilentDrop() {
	req := require.New(s.T())

	p := struct {
		Email string `db:"email"`
	}{}

	accessor := New(s.Db)
	err := accessor.Get(context.Background(), &p, "SELECT * FROM person WHERE first_name=?", "foo")
	req.True(err == nil)
	req.True(p.Email == "foo@test")
}

/////////////////////////////////////////////////////////////////////////////

type BaseEntity struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type ChildEntity struct {
	BaseEntity `db:",table=base"`
	ChildAttr  string `db:"child_attr"`
}

type GrandChildEntity struct {
	ChildEntity    `db:",table=child"`
	GrandChildAttr string `db:"grand_child_attr"`
}

type ChildEntityWithNonEmptyEmbeddedColumnMapping struct {
	BaseEntity `db:"notempty,table=base"`
	ChildAttr  string `db:"child_attr"`
}

type ChildEntityWithEmbeddedColumnWithoutTableAttribute struct {
	BaseEntity `db:",readonly"`
	ChildAttr  string `db:"child_attr"`
}

type GrandChildEntityWithNonEmptyEmbeddedChildColumnMapping struct {
	ChildEntityWithNonEmptyEmbeddedColumnMapping `db:",table=child"`
	GrandChildAttr                               string `db:"grand_child_attr"`
}

func (s *AccessorTestSuite) TestGetMapping() {
	req := require.New(s.T())

	a := New(s.Db)

	idColumns, colValMap, err := a.getMapping(&BaseEntity{Id: 1, Name: "gdbc"}, "Id")
	req.NoError(err)

	req.Equal([]string{"id"}, idColumns)
	req.Equal(2, len(colValMap))
	req.Equal(1, colValMap["id"].Interface())
	req.Equal("gdbc", colValMap["name"].Interface())

	idColumns, colValMap, err = a.getMapping(
		&ChildEntity{
			BaseEntity{
				Id:   1,
				Name: "gdbc",
			},
			"child",
		},
		"Id",
	)
	req.NoError(err)
	req.Equal([]string{"id"}, idColumns)
	req.Equal(3, len(colValMap))
	req.Equal(1, colValMap["id"].Interface())
	req.Equal("gdbc", colValMap["name"].Interface())
	req.Equal("child", colValMap["child_attr"].Interface())

	idColumns, colValMap, err = a.getMapping(
		&GrandChildEntity{
			ChildEntity{
				BaseEntity{
					Id:   1,
					Name: "gdbc",
				},
				"child",
			},
			"grandchild",
		},
		"Id",
	)
	req.NoError(err)
	req.Equal([]string{"id"}, idColumns)
	req.Equal(4, len(colValMap))
	req.Equal(1, colValMap["id"].Interface())
	req.Equal("gdbc", colValMap["name"].Interface())
	req.Equal("child", colValMap["child_attr"].Interface())
	req.Equal("grandchild", colValMap["grand_child_attr"].Interface())
}

func (s *AccessorTestSuite) TestHappySchema() {
	req := require.New(s.T())

	e := GrandChildEntity{}
	e.Id = 1
	e.Name = "base"
	e.ChildAttr = "child"
	e.GrandChildAttr = "grand_child"

	schema, err := EntitySchema(e, reflect.TypeOf(e), "grand_child")
	req.NoError(err)

	req.EqualValues(schema, &EntityMappingSchema{
		TableName: "grand_child",
		Columns: map[string]string{
			"GrandChildAttr": "grand_child_attr",
		},
		BaseMappings: []*EntityMappingSchema{
			{
				TableName: "child",
				Columns: map[string]string{
					"ChildAttr": "child_attr",
				},
				BaseMappings: []*EntityMappingSchema{
					{
						TableName: "base",
						Columns: map[string]string{
							"Id":   "id",
							"Name": "name",
						},
						Entity:     e.BaseEntity,
						EntityType: reflect.TypeOf(e.BaseEntity),
					},
				},
				Entity:     e.ChildEntity,
				EntityType: reflect.TypeOf(e.ChildEntity),
			},
		},
		Entity:     e,
		EntityType: reflect.TypeOf(e),
	})
}

func (s *AccessorTestSuite) TestFailedSchema() {
	req := require.New(s.T())

	e1 := ChildEntityWithNonEmptyEmbeddedColumnMapping{}

	_, err := EntitySchema(e1, reflect.TypeOf(e1), "child")
	req.Error(err)
	req.Equal(err.Error(), "embedded type BaseEntity in type ChildEntityWithNonEmptyEmbeddedColumnMapping should have empty column name")

	_, err = EntitySchema(ChildEntity{}, reflect.TypeOf(ChildEntity{}), "base")
	req.Error(err)
	req.Equal(err.Error(), "embedded type BaseEntity in type ChildEntity should not have the same table mapping")

	e2 := ChildEntityWithEmbeddedColumnWithoutTableAttribute{}
	_, err = EntitySchema(e2, reflect.TypeOf(e2), "child")
	req.Error(err)
	req.Equal(err.Error(), "embedded type BaseEntity in type ChildEntityWithEmbeddedColumnWithoutTableAttribute should have table attribute")

	e3 := GrandChildEntityWithNonEmptyEmbeddedChildColumnMapping{}
	_, err = EntitySchema(e3, reflect.TypeOf(e3), "grand_child")
	req.Error(err)
	req.Equal(err.Error(), "embedded type BaseEntity in type ChildEntityWithNonEmptyEmbeddedColumnMapping should have empty column name")
}

func (s *AccessorTestSuite) TestSchemaTables() {
	req := require.New(s.T())

	e := GrandChildEntity{}
	e.Id = 100

	schema, err := EntitySchema(e, reflect.TypeOf(e), "grand_child")
	req.NoError(err)

	r := schema.Schemas()
	req.Equal(100, r[0].Entity.(BaseEntity).Id)

	tables := schema.Tables()
	req.EqualValues([]string{"base", "child", "grand_child"}, tables)

	sel := schema.GetColumnSelectString()
	req.Equal("base.*, child.*, grand_child.*", sel)

	join := schema.GetTableJoinString("id")
	req.Equal("child ON base.id=child.id JOIN grand_child ON child.id=grand_child.id", join)

	join = schema.GetTableJoinString("id1", "id2")
	req.Equal("child ON base.id1=child.id1 AND base.id2=child.id2 JOIN grand_child ON child.id1=grand_child.id1 AND child.id2=grand_child.id2", join)
}

func (s *AccessorTestSuite) TestEmbeddedColumnMapping() {
	req := require.New(s.T())

	e := GrandChildEntity{}

	col := Column(e, "Id")
	req.Equal("id", col)

	col = Column(e, "Name")
	req.Equal("name", col)

	col = Column(e, "ChildAttr")
	req.Equal("child_attr", col)

	col = Column(e, "GrandChildAttr")
	req.Equal("grand_child_attr", col)
}
