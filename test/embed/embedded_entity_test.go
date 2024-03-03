package embed

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kelveny/gdbc/pkg/accessor"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

type AccessorEmbeddedEntityTestSuite struct {
	suite.Suite
	Logger *zap.Logger

	Db *sqlx.DB
}

func TestEmbeddedEntity(t *testing.T) {
	suite.Run(t, new(AccessorEmbeddedEntityTestSuite))
}

func (s *AccessorEmbeddedEntityTestSuite) SetupSuite() {
	req := require.New(s.T())

	s.Logger = zap.NewExample()

	var db *sqlx.DB
	var err error

	if db, err = sqlx.Open("postgres", "postgres://postgres:postgres@localhost:5432/gdbc_test?sslmode=disable"); err != nil {
		req.FailNow("setup suite failed: %w", err)
	}

	s.Db = db
	s.setupTestDatabase()
}

func (s *AccessorEmbeddedEntityTestSuite) TearDownSuite() {
	if s.Db != nil {
		s.teardownTestDatabase()

		if err := s.Db.Close(); err != nil {
			s.Logger.Error("close db connection: %w", zap.Error(err))
		}
	}
}

func (s *AccessorEmbeddedEntityTestSuite) setupTestDatabase() {
	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS manager;
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS person;
DROP TYPE IF EXISTS mood;
DROP SEQUENCE IF EXISTS person_id_seq;
	
CREATE TYPE mood AS ENUM ('happy', 'sad', 'angry', 'calm');

CREATE TABLE IF NOT EXISTS person (
	id serial primary key,
    first_name text,
    last_name text,
    email text,
	age integer,
	current_mood mood, 
    added_at timestamp
);

CREATE SEQUENCE IF NOT EXISTS person_id_seq START WITH 1 INCREMENT BY 1 CYCLE;

CREATE TABLE IF NOT EXISTS employee (
	id integer primary key,
	company text
);
ALTER TABLE employee ADD CONSTRAINT fk_employee_id FOREIGN KEY (id) REFERENCES person(id);

CREATE TABLE IF NOT EXISTS manager (
	id integer primary key,
	title text
);
ALTER TABLE manager ADD CONSTRAINT fk_manager_id FOREIGN KEY (id) REFERENCES employee(id);
	`)
}

func (s *AccessorEmbeddedEntityTestSuite) teardownTestDatabase() {
	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS manager;
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS person;
DROP TYPE IF EXISTS mood;
DROP SEQUENCE IF EXISTS person_id_seq;
    `)
}

func (s *AccessorEmbeddedEntityTestSuite) TestEmbeddedGetFromBase() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// create
	p := Person{
		FirstName: "foo",
		LastName:  "gdbc",
	}
	err := a.Create(context.Background(), &p, "person")
	req.NoError(err)
	req.True(p.Id != 0)

	e := Employee{}
	e.FirstName = "bar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("angry")
	e.Company = toPtr("foo.com")

	err = a.Create(context.Background(), &e, "employee")
	req.NoError(err)
	req.True(e.Id != 0)

	m := Manager{}
	m.FirstName = "foobar"
	m.LastName = "gdbc"
	m.CurrentMood = toPtr("happy")
	m.Company = toPtr("foo.com")
	m.Title = toPtr("CIO")
	m.Id = 1000

	err = a.Create(context.Background(), &m, "manager")
	req.NoError(err)
	req.True(m.Id == 1000)

	// Read
	p2 := Person{}
	p2.FirstName = "foo"
	err = a.Read(context.Background(), &p2, "person", "FirstName")
	req.NoError(err)
	req.True(p2.Id != 0)
	req.Equal("foo", p2.FirstName)
	req.Equal("gdbc", p2.LastName)

	e2 := Employee{}
	e2.Id = 2

	// composite entity CRUD can only be read via true ID fields
	err = a.Read(context.Background(), &e2, "employee", "Id")
	req.NoError(err)
	req.Equal("bar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("angry", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)

	m2 := Manager{}
	m2.Id = 1000

	// composite entity CRUD can only be read via true ID fields
	err = a.Read(context.Background(), &m2, "manager")
	req.NoError(err)
	req.Equal("foobar", m2.FirstName)
	req.Equal("gdbc", m2.LastName)
	req.Equal("happy", *m2.CurrentMood)
	req.Equal("foo.com", *m2.Company)
	req.Equal("CIO", *m2.Title)

	// Negative reads
	p2.Id = 0
	err = a.Read(context.Background(), &p2, "person")
	req.Error(err)

	e2.Id = 0
	err = a.Read(context.Background(), &e2, "employee")
	req.Error(err)

	m2.Id = 0
	err = a.Read(context.Background(), &m2, "manager")
	req.Error(err)

	// delete
	p3 := Person{}
	p3.Id = 1
	_, err = a.Delete(context.Background(), &p3, "person")
	req.NoError(err)
	req.True(p3.Id == 1)
	req.Equal("foo", p3.FirstName)
	req.Equal("gdbc", p3.LastName)

	e3 := Employee{}
	e3.Id = 2
	_, err = a.Delete(context.Background(), &e3, "employee")
	req.NoError(err)
	req.True(e3.Id == 2)
	req.Equal("bar", e3.FirstName)
	req.Equal("gdbc", e3.LastName)
	req.Equal("angry", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)

	m3 := Manager{}
	m3.Id = 1000
	_, err = a.Delete(context.Background(), &m3, "manager")
	req.NoError(err)
	req.True(m3.Id == 1000)
	req.Equal("foobar", m3.FirstName)
	req.Equal("gdbc", m3.LastName)
	req.Equal("happy", *m3.CurrentMood)
	req.Equal("foo.com", *m3.Company)
	req.Equal("CIO", *m3.Title)

	// negative deletes
	p3.Id = 0
	err = a.Read(context.Background(), &p3, "person")
	req.Error(err)

	e3.Id = 0
	err = a.Read(context.Background(), &e3, "employee")
	req.Error(err)

	m3.Id = 0
	err = a.Read(context.Background(), &m3, "manager")
	req.Error(err)
}

func toPtr(s string) *string {
	return &s
}