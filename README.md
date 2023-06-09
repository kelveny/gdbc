# GDBC - Simple Database Accessor for Go

The goal of `gdbc` is not to provide yet another [ORM](https://en.wikipedia.org/wiki/Object%E2%80%93relational_mapping) solution in Go, it is to combine the power of two popular tools ([squirrel](https://github.com/Masterminds/squirrel) and [sqlx](https://github.com/jmoiron/sqlx)), give developers a simple and straight forward way of accessing database in Go. `gdbc` has following design goals:

- Minimal pre-setup
- Not intrusive, you don't need a framework to use it
- Safe and implicit transaction management
- Out-of-box CURD operations for database entity objects
- Flexible in various use cases

Please note, almost all `gdbc` functions take `context.Context` as function input parameters, for contexts that are cancellable, be aware of that it may impact query results. You might want to use isolated context for database operations. We made this trade-off for simplifying `gdbc` interface.

Transaction support is implicit. The caller only needs to start the transaction flow with a static function, any operations carried out within the anonymous function are safe in case of a system crash. If the function executes without errors, the transaction will be finalized and committed. However, if an error or `panic` occurs during the function execution, the transaction will be reverted or rolled back.

`gdbc` is also equipped with its own `go:generate` tooling, `gdbc` entity enhancer can generate compile-time type-safe meta types to help developer write queries, it can also generate corresponding enhanced entity type for partial updates.

You are not restricted to utlize generated compile-time type-safe meta types with `gdbc` built-in database accessors. You can employ them in any context where compile-time type-safe mapping is applicable. For instance, you can combine them with other Go database accessing libraries.

## Install

```bash
go install github.com/kelveny/gdbc
```

## Usage examples

### 1. Enhance entity type

```go
//go:generate gdbc -entity Person -table person
type Person struct {
    FirstName string `db:"first_name"`
    LastName  string `db:"last_name"`
    Email     string `db:"email"`
}
```

It has `go generate` generated code as:

```go
// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package enhancer

type PersonEntityFields struct {
    FirstName string
    LastName  string
    Email     string
}

type PersonTableColumns struct {
    FirstName string
    LastName  string
    Email     string
}

func (e *Person) TableName() string {
    return "person"
}

func (e *Person) EntityFields() *PersonEntityFields {
    return &PersonEntityFields{
        FirstName: "FirstName",
        LastName:  "LastName",
        Email:     "Email",
    }
}

func (e *Person) TableColumns() *PersonTableColumns {
    return &PersonTableColumns{
        FirstName: "first_name",
        LastName:  "last_name",
        Email:     "email",
    }
}

type PersonWithUpdateTracker struct {
    Person
    trackMap map[string]bool
}

func (e *PersonWithUpdateTracker) ColumnsChanged() []string {
    cols := []string{}

    for col, _ := range e.trackMap {
        cols = append(cols, col)
    }

    return cols
}

func (e *PersonWithUpdateTracker) SetFirstName(val string) *PersonWithUpdateTracker {
    e.FirstName = val

    if e.trackMap == nil {
        e.trackMap = make(map[string]bool)
    }

    e.trackMap["first_name"] = true

    return e
}

func (e *PersonWithUpdateTracker) SetLastName(val string) *PersonWithUpdateTracker {
    e.LastName = val

    if e.trackMap == nil {
        e.trackMap = make(map[string]bool)
    }

    e.trackMap["last_name"] = true

    return e
}

func (e *PersonWithUpdateTracker) SetEmail(val string) *PersonWithUpdateTracker {
    e.Email = val

    if e.trackMap == nil {
        e.trackMap = make(map[string]bool)
    }

    e.trackMap["email"] = true

    return e
}
```

### 2. Create

```go
    accessor := accessor.New(s.Db)
    p := Person{
        FirstName: "foofoo",
        LastName:  "test",
        Email:     "foofoo@test",
    }

    // assume that FirstName is primary key
    err := accessor.Create(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
```

### 3. Read

```go
    pp := Person{
        FirstName: "foofoo",
    }

    // assume that FirstName is primary key
    err = accessor.Read(context.Background(), &pp, pp.TableName(), pp.EntityFields().FirstName)
```

### 4. Partial Update

```go
    accessor := accessor.New(s.Db)
    p := PersonWithUpdateTracker{}
    p.FirstName = "foo"
    p.SetEmail("foo-changed@test")

    // assume that FirstName is primary key
    result, err := accessor.Update(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
```

### 5. Full Update

```go
   accessor := accessor.New(s.Db)
    p := Person{}
    p.FirstName = "foo"
    p.LastName = "test"
    p.SetEmail("foo-changed@test")

    // assume that FirstName is primary key
    result, err := accessor.Update(context.Background(), &p, p.TableName(), p.EntityFields().FirstName)
```

### 6. Delete

```go
    pp := Person{
        FirstName: "foofoo",
    }
    result, err := accessor.Delete(context.Background(), &pp, pp.TableName(), pp.EntityFields().FirstName)
```

Note, you can use either a value type or a pointer type for `entity` parameter of `Delete` method, when it is a pointer type as above example, `Delete` method also returns the read entity.

### 7. Get a single entity by direct SQL query and ad-hoc entity type

```go
   var p struct {
       FirstName string    `db:"first_name"`
       LastName  string    `db:"last_name"`
       Email     string    `db:"email"`
       AddedAt   time.Time `db:"added_at"`
   }

   err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
       "foo", "test")
```

### 8. Select a list of entities by direct SQL query and ad-hoc entity type

```go
   var p []struct {
       FirstName string `db:"first_name"`
       LastName  string `db:"last_name"`
       Email     string `db:"email"`
   }

   err := accessor.Select(context.Background(), &p, "select * from person where last_name=?", "test")
```

### 9. Generic execution

```go
   result, err := accessor.Exec(context.Background(), "delete from person where first_name=?", "foo")
```

### 10. Named Get query to find a single entity

```go
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
```

### 11. Qeury by example

```go
   // query by example
   var p, example struct {
       FirstName string    `db:"first_name"`
       LastName  string    `db:"last_name"`
       Email     string    `db:"email"`
       AddedAt   time.Time `db:"added_at"`
   }

   example.FirstName = "foo"
   example.LastName = "test"
   err := accessor.NamedGet(context.Background(), &p, "select * from person where first_name=:first_name and last_name=:last_name", &example)
```

### 12. Named Select query

```go
   var p []struct {
       FirstName string `db:"first_name"`
       LastName  string `db:"last_name"`
       Email     string `db:"email"`
   }

   err := accessor.NamedSelect(context.Background(), &p, "select * from person where last_name=:last_name order by first_name DESC",
       map[string]any{
           "last_name": "test",
       })
```

### 13. Get a single entity with query builder

```go
   p := Person{}
   err := accessor.SqlizerGet(context.Background(), &p, func(builder squirrel.StatementBuilderType) Sqlizer {
       return builder.Select("*").From(p.TableName()).Where(squirrel.Eq{
           p.TableColumns().FirstName: "foo",
           p.TableColumns().LastName:  "test",
       })
   })
```

### 14. Select a list of entities with query builder

```go
   var p []Person

   err := accessor.SqlizerSelect(context.Background(), &p, func(builder squirrel.StatementBuilderType) Sqlizer {
       return builder.Select("*").From(Person{}.TableName()).Where(squirrel.Eq{
           Person{}.TableColumns().LastName: "test",
       })
   })
```

### 15. Implicit crash-safe transaction

```go
   err := ExecTx(context.Background(), s.Db, &sql.TxOptions{}, func(ctx context.Context, accessor *Accessor) error {
       var p struct {
           FirstName string    `db:"first_name"`
           LastName  string    `db:"last_name"`
           Email     string    `db:"email"`
           AddedAt   time.Time `db:"added_at"`
       }

       err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
           "foo", "test")

       return nil
   })
```

## Notes about entity CRUD

Although [sqlx](https://github.com/jmoiron/sqlx) supports complex (nested) mapping operations, entity CRUD methods(`Create`, `Read`, `Update`, `Delete`) utilize one-level only mapping operations. This is a design choice to make entity level CRUD operations be generic. For entity types that have complex fields, you will need to employ [sql.Scanner](https://pkg.go.dev/database/sql#Scanner) and [driver.Valur](https://cs.opensource.google/go/go/+/refs/tags/go1.20.5:src/database/sql/driver/types.go;l=39) facilities to map between complex field types and driver supported value types. Following example illustrates such a mapping for `NULL-able` primitive value types in [Postgres](https://www.postgresql.org/).

```go
import (
    "database/sql"
    "database/sql/driver"
    "errors"

    "github.com/jackc/pgtype"
)

type NullPrimitive[T any] struct {
    TypeValue T
    Valid     bool
}

// Scan implements the sql.Scanner interface.
func (e *NullPrimitive[T]) Scan(value interface{}) error {
    if e == nil || value == nil {
        return nil
    }

    ci := pgtype.NewConnInfo()
    if dt, ok := ci.DataTypeForValue(e.TypeValue); ok {
        err := dt.Value.(sql.Scanner).Scan(value)
        if err == nil {
            err = dt.Value.AssignTo(&e.TypeValue)
        }
        if err == nil {
            e.Valid = true
        }
        return err
    } else {
        return errors.New("unregistered primitive value type")
    }
}

// Value implements the driver.Valuer interface.
func (e NullPrimitive[T]) Value() (driver.Value, error) {
    if !e.Valid {
        return nil, nil
    }

    ci := pgtype.NewConnInfo()

    if dt, ok := ci.DataTypeForValue(e.TypeValue); ok {
        v := pgtype.NewValue(dt.Value)

        err := v.Set(e.TypeValue)
        if err != nil {
            return nil, err
        }

        return v, nil
    } else {
        return nil, errors.New("unregistered primitive value type")
    }
}
```

For complex field types with nested structure, Go [json](https://pkg.go.dev/encoding/json) marshaler can be your friend to bridge basic driver supported types and complex field types under [sql.Scanner](https://pkg.go.dev/database/sql#Scanner)/[driver.Valur](https://cs.opensource.google/go/go/+/refs/tags/go1.20.5:src/database/sql/driver/types.go;l=39) framework.
