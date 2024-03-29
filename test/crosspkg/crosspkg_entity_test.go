package crosspkg

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kelveny/gdbc/pkg/accessor"
	"github.com/kelveny/gdbc/test/embed"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

type CrossPackageEntityTestSuite struct {
	suite.Suite
	Logger *zap.Logger

	Db *sqlx.DB
}

func TestCrossPackageEntity(t *testing.T) {
	suite.Run(t, new(CrossPackageEntityTestSuite))
}

func (s *CrossPackageEntityTestSuite) SetupSuite() {
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

func (s *CrossPackageEntityTestSuite) TearDownSuite() {
	if s.Db != nil {
		s.teardownTestDatabase()

		if err := s.Db.Close(); err != nil {
			s.Logger.Error("close db connection: %w", zap.Error(err))
		}
	}
}

func (s *CrossPackageEntityTestSuite) setupTestDatabase() {
	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS executive;
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

CREATE TABLE IF NOT EXISTS executive (
	id integer primary key,
	term text
);
ALTER TABLE executive ADD CONSTRAINT fk_executive_id FOREIGN KEY (id) REFERENCES manager(id);

	`)
}

func (s *CrossPackageEntityTestSuite) teardownTestDatabase() {
	_ = s.Db.MustExec(`
DROP TABLE IF EXISTS executive;
DROP TABLE IF EXISTS manager;
DROP TABLE IF EXISTS employee;
DROP TABLE IF EXISTS person;
DROP TYPE IF EXISTS mood;
DROP SEQUENCE IF EXISTS person_id_seq;
    `)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive{}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive{}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive{}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive{}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded2() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive2{
		Manager2: embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive2{
		Manager2: embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive2{
		Manager2: embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive2{
		Manager2: embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded3() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive3{
		Manager3: embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive3{
		Manager3: embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		}}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive3{
		Manager3: embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive3{
		Manager3: embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded4() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive4{
		Manager4: embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive4{
		Manager4: embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive4{
		Manager4: embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive4{
		Manager4: embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded5() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive5{
		Manager: &embed.Manager{},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive5{
		Manager: &embed.Manager{},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive5{
		Manager: &embed.Manager{},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive5{
		Manager: &embed.Manager{},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded6() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive6{
		Manager2: &embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive6{
		Manager2: &embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive6{
		Manager2: &embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive6{
		Manager2: &embed.Manager2{
			Employee: &embed.Employee{},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded7() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive7{
		Manager3: &embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive7{
		Manager3: &embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		}}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive7{
		Manager3: &embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive7{
		Manager3: &embed.Manager3{
			Employee2: embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func (s *CrossPackageEntityTestSuite) TestEmbedded8() {
	req := require.New(s.T())

	a := accessor.New(s.Db)

	// test with non-zero/non-nil values
	e := Executive8{
		Manager4: &embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foobar"
	e.LastName = "gdbc"
	e.CurrentMood = toPtr("happy")
	e.Company = toPtr("foo.com")
	e.Title = toPtr("CIO")
	e.Term = toPtr("nonsense")

	err := a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)

	e2 := Executive8{
		Manager4: &embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foobar", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.Equal("happy", *e2.CurrentMood)
	req.Equal("foo.com", *e2.Company)
	req.Equal("CIO", *e2.Title)
	req.Equal("nonsense", *e2.Term)

	// test with nil input at every level
	e = Executive8{
		Manager4: &embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e.FirstName = "foo"
	e.LastName = "gdbc"

	err = a.Create(context.Background(), &e, e.TableName())
	req.NoError(err)
	req.True(e.Id != 0)
	req.Equal("foo", e.FirstName)
	req.Equal("gdbc", e.LastName)
	req.True(e.AddedAt == nil)
	req.True(e.Age == nil)
	req.True(e.Email == nil)
	req.True(e.CurrentMood == nil)
	req.True(e.Company == nil)
	req.True(e.Title == nil)
	req.True(e.Term == nil)

	e2 = Executive8{
		Manager4: &embed.Manager4{
			Employee2: &embed.Employee2{
				Person: &embed.Person{},
			},
		},
	}
	e2.Id = e.Id

	_, err = a.Delete(context.Background(), &e2, e2.TableName())
	req.NoError(err)
	req.Equal("foo", e2.FirstName)
	req.Equal("gdbc", e2.LastName)
	req.True(e2.AddedAt == nil)
	req.True(e2.Age == nil)
	req.True(e2.Email == nil)
	req.True(e2.CurrentMood == nil)
	req.True(e2.Company == nil)
	req.True(e2.Title == nil)
	req.True(e2.Term == nil)
}

func toPtr(s string) *string {
	return &s
}
