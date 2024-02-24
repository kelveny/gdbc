package enhancer

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kelveny/gdbc/pkg/accessor"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/mattn/go-sqlite3"
)

type TestSuite struct {
	suite.Suite
	Logger *zap.Logger

	Db *sqlx.DB
}

func TestEnhancer(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) SetupSuite() {
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

func (s *TestSuite) TearDownSuite() {
	if s.Db != nil {
		s.teardownTestDatabase()

		if err := s.Db.Close(); err != nil {
			s.Logger.Error("close db connection: %w", zap.Error(err))
		}
	}
}

func (s *TestSuite) setupTestDatabase() {
	_ = s.Db.MustExec(`
CREATE TABLE IF NOT EXISTS person (
	first_name text,
	last_name text,
	email text,
	added_at timestamp
);
	`)

	_ = s.Db.MustExec(`DELETE FROM person`)

	// we need explicit binding for ? in query when using sqlx functions
	r := s.Db.Rebind
	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, added_at) VALUES (?, ?, ?, ?)	
	`), "foo", "test", "foo@test", time.Now().UTC())

	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, added_at) VALUES (?, ?, ?, ?)	
	`), "bar", "test", "bar@test", time.Now().UTC())

	_ = s.Db.MustExec(r(`
INSERT INTO person(first_name, last_name, email, added_at) VALUES (?, ?, ?, ?)	
	`), "baz", "test", nil, time.Now().UTC())

}

func (s *TestSuite) teardownTestDatabase() {
	assert := require.New(s.T())
	accessor := accessor.New(s.Db)

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
}

func (s *TestSuite) TestGeneratedTypes() {
	assert := require.New(s.T())

	accessor := accessor.New(s.Db)
	p := Person{
		FirstName: "foo",
	}

	err := accessor.Read(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
	assert.True(err == nil)
	assert.True(p.LastName == "test")
	assert.True(p.Email == "foo@test")

	p = Person{
		FirstName: "baz",
	}
	err = accessor.Read(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
	assert.True(err == nil)
	assert.True(p.LastName == "test")
	assert.True(p.Email == "")

}

func (s *TestSuite) TestPartialUpdate() {
	assert := require.New(s.T())

	accessor := accessor.New(s.Db)
	p := PersonWithUpdateTracker{}
	p.FirstName = "foo"
	p.SetEmail("foo-changed@test")

	result, err := accessor.Update(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)

	pp := Person{
		FirstName: "foo",
	}

	err = accessor.Read(context.Background(), &pp, pp.TableName(), pp.EntityFields().FirstName)
	assert.True(err == nil)
	assert.True(pp.FirstName == "foo")
	assert.True(pp.LastName == "test")
	assert.True(pp.Email == "foo-changed@test")
}

func (s *TestSuite) TestCreateReadDelete() {
	assert := require.New(s.T())

	accessor := accessor.New(s.Db)
	p := Person{
		FirstName: "foofoo",
		LastName:  "test",
		Email:     "foofoo@test",
	}

	err := accessor.Create(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
	assert.True(err == nil)

	pp := Person{
		FirstName: "foofoo",
	}

	err = accessor.Read(context.Background(), &pp, pp.TableName(), pp.EntityFields().FirstName)
	assert.True(err == nil)
	assert.True(pp.FirstName == "foofoo")
	assert.True(pp.LastName == "test")
	assert.True(pp.Email == "foofoo@test")

	result, err := accessor.Delete(context.Background(), &pp, pp.TableName(), pp.EntityFields().FirstName)
	assert.True(err == nil)
	affected, err := result.RowsAffected()
	assert.True(err == nil)
	assert.True(affected == 1)
}
