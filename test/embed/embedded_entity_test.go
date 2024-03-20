package embed

import (
	"context"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/kelveny/gdbc/pkg/accessor"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

type EmbeddedEntityTestSuite struct {
	suite.Suite
	Logger *zap.Logger

	Db *sqlx.DB
}

func TestEmbeddedEntity(t *testing.T) {
	suite.Run(t, new(EmbeddedEntityTestSuite))
}

func (s *EmbeddedEntityTestSuite) SetupSuite() {
	req := require.New(s.T())

	s.Logger = zap.NewExample()

	var db *sqlx.DB
	var err error

	// prerequisite
	// 	. have a local Postgres instance running locally at port 5432
	//  . have a empty database named "gdbc_test" created in the Postgres instance
	if db, err = sqlx.Open("postgres", "postgres://postgres:postgres@localhost:5432/gdbc_test?sslmode=disable"); err != nil {
		req.FailNow("setup suite failed: %w", err)
	}

	s.Db = db
	s.setupTestDatabase()
}

func (s *EmbeddedEntityTestSuite) TearDownSuite() {
	if s.Db != nil {
		s.teardownTestDatabase()

		if err := s.Db.Close(); err != nil {
			s.Logger.Error("close db connection: %w", zap.Error(err))
		}
	}
}

func (s *EmbeddedEntityTestSuite) setupTestDatabase() {
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

func (s *EmbeddedEntityTestSuite) teardownTestDatabase() {
	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS manager;
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS person;
DROP TYPE IF EXISTS mood;
DROP SEQUENCE IF EXISTS person_id_seq;
    `)
}

func (s *EmbeddedEntityTestSuite) TestEmbedded() {
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

	// nested embedding
	mp := ManagerWrapper{}
	mp.FirstName = "foobar"
	mp.LastName = "gdbc"
	mp.CurrentMood = toPtr("happy")
	mp.Company = toPtr("foo.com")
	mp.Title = toPtr("CIO")
	mp.Id = 1001

	err = a.Create(context.Background(), &mp, "manager")
	req.NoError(err)
	req.True(mp.Id == 1001)

	_, err = a.Delete(context.Background(), &mp, "manager")
	req.NoError(err)

	// multiple level of nested embedding
	mpp := ManagerWrapperWrapper{}
	mpp.FirstName = "foobar"
	mpp.LastName = "gdbc"
	mpp.CurrentMood = toPtr("happy")
	mpp.Company = toPtr("foo.com")
	mpp.Title = toPtr("CIO")
	mpp.Id = 1002

	err = a.Create(context.Background(), &mpp, "manager")
	req.NoError(err)
	req.True(mpp.Id == 1002)

	_, err = a.Delete(context.Background(), &mpp, "manager")
	req.NoError(err)

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

	m2p := ManagerWrapper{}
	m2p.Id = 1000

	// composite entity CRUD can only be read via true ID fields
	err = a.Read(context.Background(), &m2p, "manager")
	req.NoError(err)
	req.Equal("foobar", m2p.FirstName)
	req.Equal("gdbc", m2p.LastName)
	req.Equal("happy", *m2p.CurrentMood)
	req.Equal("foo.com", *m2p.Company)
	req.Equal("CIO", *m2p.Title)

	m2pp := ManagerWrapperWrapper{}
	m2pp.Id = 1000

	// composite entity CRUD can only be read via true ID fields
	err = a.Read(context.Background(), &m2pp, "manager")
	req.NoError(err)
	req.Equal("foobar", m2pp.FirstName)
	req.Equal("gdbc", m2pp.LastName)
	req.Equal("happy", *m2pp.CurrentMood)
	req.Equal("foo.com", *m2pp.Company)
	req.Equal("CIO", *m2pp.Title)

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

	// update & read-back
	e4 := EmployeeWithUpdateTracker{}
	e4.Id = 2

	e4.SetCompany(toPtr("bar.com"))
	e4.SetCurrentMood(toPtr("sad"))

	_, err = a.Update(context.Background(), &e4, "employee")
	req.NoError(err)

	e = Employee{}
	e.Id = 2
	err = a.Read(context.Background(), &e, "employee")
	req.NoError(err)
	req.Equal("bar.com", *e.Company)
	req.Equal("sad", *e.CurrentMood)

	m4 := ManagerWithUpdateTracker{}
	m4.Id = 1000
	m4.SetCompany(toPtr("bar.com"))
	m4.SetCurrentMood(toPtr("sad"))
	m4.SetTitle(toPtr("CEO"))

	_, err = a.Update(context.Background(), &m4, "manager")
	req.NoError(err)
	m = Manager{}
	m.Id = 1000
	err = a.Read(context.Background(), &m, "manager")
	req.NoError(err)
	req.Equal("bar.com", *m.Company)
	req.Equal("sad", *m.CurrentMood)
	req.Equal("CEO", *m.Title)

	m4_2 := Manager2WithUpdateTracker{
		Manager2: Manager2{
			Employee: &Employee{},
		},
	}
	m4_2.Id = 1000
	m4_2.SetCompany(toPtr("bar2.com"))
	m4_2.SetCurrentMood(toPtr("happy"))
	m4_2.SetTitle(toPtr("CFO"))

	_, err = a.Update(context.Background(), &m4_2, "manager")
	req.NoError(err)
	m = Manager{}
	m.Id = 1000
	err = a.Read(context.Background(), &m, "manager")
	req.NoError(err)
	req.Equal("bar2.com", *m.Company)
	req.Equal("happy", *m.CurrentMood)
	req.Equal("CFO", *m.Title)

	m4_3 := Manager3WithUpdateTracker{
		Manager3: Manager3{
			Employee2: Employee2{
				Person: &Person{},
			},
		},
	}
	m4_3.Id = 1000
	m4_3.SetCompany(toPtr("bar3.com"))
	m4_3.SetCurrentMood(toPtr("sad"))
	m4_3.SetTitle(toPtr("COO"))

	_, err = a.Update(context.Background(), &m4_3, "manager")
	req.NoError(err)
	m = Manager{}
	m.Id = 1000
	err = a.Read(context.Background(), &m, "manager")
	req.NoError(err)
	req.Equal("bar3.com", *m.Company)
	req.Equal("sad", *m.CurrentMood)
	req.Equal("COO", *m.Title)

	m4_4 := Manager4WithUpdateTracker{
		Manager4: Manager4{
			Employee2: &Employee2{
				Person: &Person{},
			},
		},
	}
	m4_4.Id = 1000
	m4_4.SetCompany(toPtr("bar.com"))
	m4_4.SetCurrentMood(toPtr("sad"))
	m4_4.SetTitle(toPtr("CEO"))

	_, err = a.Update(context.Background(), &m4_4, "manager")
	req.NoError(err)
	m = Manager{}
	m.Id = 1000
	err = a.Read(context.Background(), &m, "manager")
	req.NoError(err)
	req.Equal("bar.com", *m.Company)
	req.Equal("sad", *m.CurrentMood)
	req.Equal("CEO", *m.Title)

	// EntityGet & EntitySelect
	m = Manager{}
	err = a.EntityGet(
		context.Background(),
		&m,
		m.TableName(),
		func(builder squirrel.SelectBuilder) accessor.Sqlizer {
			return builder.Where(squirrel.Eq{
				m.Employee.TableName() + "." + m.Employee.EntityFields().Company: "bar.com",
			})
		},
	)
	req.NoError(err)
	req.Equal("bar.com", *m.Company)
	req.Equal("sad", *m.CurrentMood)
	req.Equal("CEO", *m.Title)
	req.Equal(1000, m.Id)

	mgrList := []Manager{}
	err = a.EntitySelect(
		context.Background(),
		&mgrList,
		m.TableName(),
		func(builder squirrel.SelectBuilder) accessor.Sqlizer {
			return builder.Where(squirrel.Eq{
				m.Employee.TableName() + "." + m.Employee.EntityFields().Company: "bar.com",
			})
		},
	)
	req.NoError(err)
	req.Equal(1, len(mgrList))

	req.Equal("bar.com", *mgrList[0].Company)
	req.Equal("sad", *mgrList[0].CurrentMood)
	req.Equal("CEO", *mgrList[0].Title)
	req.Equal(1000, mgrList[0].Id)

	mgrList2 := []*Manager{}
	err = a.EntitySelect(
		context.Background(),
		&mgrList2,
		m.TableName(),
		func(builder squirrel.SelectBuilder) accessor.Sqlizer {
			return builder.Where(squirrel.Eq{
				m.Employee.TableName() + "." + m.Employee.EntityFields().Company: "bar.com",
			})
		},
	)
	req.NoError(err)
	req.Equal(1, len(mgrList2))

	req.Equal("bar.com", *mgrList2[0].Company)
	req.Equal("sad", *mgrList2[0].CurrentMood)
	req.Equal("CEO", *mgrList2[0].Title)
	req.Equal(1000, mgrList2[0].Id)

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
	req.Equal("sad", *e3.CurrentMood)
	req.Equal("bar.com", *e3.Company)

	m3 := Manager{}
	m3.Id = 1000
	_, err = a.Delete(context.Background(), &m3, "manager")
	req.NoError(err)
	req.True(m3.Id == 1000)
	req.Equal("foobar", m3.FirstName)
	req.Equal("gdbc", m3.LastName)
	req.Equal("sad", *m3.CurrentMood)
	req.Equal("bar.com", *m3.Company)
	req.Equal("CEO", *m3.Title)

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
