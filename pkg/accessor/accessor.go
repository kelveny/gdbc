// Accessor feature summary
//
// 1. public accessor methods
//  Get(ctx context.Context, dest any, query string, args ...any) error
//  Select(ctx context.Context, dest any, query string, args ...any) error
//  Exec(ctx context.Context, query string, args ...any) (result sql.Result, outErr error)
//
//  NamedGet(ctx context.Context, dest any, query string, arg any) error
//  NamedSelect(ctx context.Context, dest any, query string, arg any) error
//  NamedExec(ctx context.Context, query string, arg any) (sql.Result, error)
//
//    SqlizerGet(
//        ctx context.Context,
//        dest any,
//        sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
//    ) error
//    SqlizerSelect(
//        ctx context.Context,
//        dest any,
//        sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
//    ) error
//    SqlizerExec(
//        ctx context.Context,
//        sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
//    ) (result sql.Result, outErr error)
//
//  // Entity CRUD
//    Create(ctx context.Context, entity any, tbl string, idFields ...string) error
//  Read(ctx context.Context, entity any, tbl string, idFields ...string) error
//    Update(ctx context.Context, entity any, tbl string, idFields ...string) (sql.Result, error)
//  Delete(ctx context.Context, entity any, tbl string, idFields ...string) (sql.Result, error)
//
// 2. public helper functions
//    ExecTx(
//        ctx context.Context,
//        db *pg.DB,
//        txOps *sql.TxOptions,
//        execFn func(ctx context.Context, accessor *Accessor) error,
//    ) (outErr error)
//
//  Column(v any, fieldName string) string
//  Columns(v any) []string
//
// 3. Accessor itself is not thread-safe, however, its underlying backend musts be thread-safe.
// 4. Accessor assumes manipulation of Dabatabse entity objects, columns of corresponding
//    column mappings should exist in entity type (in Go struct tag "db")
// 5. Delete() and Update() supports default "Id" (column id) mapping
// 6. Accessor bridges github.com/jmoiron/sqlx and github.com/Masterminds/squirrel, allows
//      generic database CRUD and SELECT operations with regular literal query binding, named query binding
//      querying by example and programmatical query binding.
//
package accessor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

type UpdateTracker interface {
	ColumnsChanged() []string
}

type Accessor struct {
	Db sqlx.Ext
}

func New(db sqlx.Ext) *Accessor {
	return &Accessor{
		Db: db,
	}
}

// Get returns a single found entity.
// Note: wildcard column selection is allowed, unmatched column to field mapping will be silently ignored
//
// Usage example:
/*
   var p struct {
       FirstName string    `db:"first_name"`
       LastName  string    `db:"last_name"`
       Email     string    `db:"email"`
       AddedAt   time.Time `db:"added_at"`
   }

   err := accessor.Get(context.Background(), &p, "select * from person where first_name=? and last_name=?",
       "foo", "test")
*/
func (a *Accessor) Get(ctx context.Context, dest any, query string, args ...any) error {
	query = a.Db.Rebind(query)

	if db, ok := a.Db.(*sqlx.DB); ok {
		return db.Unsafe().GetContext(ctx, dest, query, args...)
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		return tx.Unsafe().GetContext(ctx, dest, query, args...)
	}

	return errors.New("invalid accessor backend")
}

// Usage example:
/*
   var p []struct {
       FirstName string `db:"first_name"`
       LastName  string `db:"last_name"`
       Email     string `db:"email"`
   }

   err := accessor.Select(context.Background(), &p, "select * from person where last_name=?", "test")
*/
func (a *Accessor) Select(ctx context.Context, dest any, query string, args ...any) error {
	query = a.Db.Rebind(query)

	if db, ok := a.Db.(*sqlx.DB); ok {
		return db.Unsafe().SelectContext(ctx, dest, query, args...)
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		return tx.Unsafe().SelectContext(ctx, dest, query, args...)
	}

	return errors.New("invalid accessor backend")
}

// Usage example:
/*
   result, err := accessor.Exec(context.Background(), "delete from person where first_name=?", "foo")
*/
func (a *Accessor) Exec(ctx context.Context, query string, args ...any) (result sql.Result, outErr error) {
	query = a.Db.Rebind(query)

	defer func() {
		if c := recover(); c != nil {
			outErr = errors.New("sql execution error: " + query)
		}
	}()

	if db, ok := a.Db.(*sqlx.DB); ok {
		return db.Unsafe().MustExecContext(ctx, query, args...), nil
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		return tx.Unsafe().MustExecContext(ctx, query, args...), nil
	}

	return nil, errors.New("invalid accessor backend")
}

//
// Insert a database row based on entity properties.
//
// Create allows caller to specify hint about its primary key column(s). Only when corresponding
// entity fields have non-zero value(s), will the zon-zero value(s) be used, otherwise, related
// privary column(s) will be inserted with database supplied value(s).
//
// Zero value of ID fields indicates to insert the row using database's auto-increment value,
// returning auto-increment ID value will be backfilled into the entity object
//
// If no idFieds is given, Create will try "Id" -> "id" mapping for primary key,
// if "Id" -> "id" mapping does not exist, Create does not perform special handling of primary
// key column(s) that may have database auto-increment enabled.
//
//
// Usage example
/*
   // create
   city := struct {
       Id      int    `db:"id"`
       Name    string `db:"name"`
       ZipCode string `db:"zip_code"`
   }{
       Name:    "San Jose",
       ZipCode: "95120",
   }

   accessor := New(s.Db.DB)
   err := accessor.Create(context.Background(), &city, "city")
*/
func (a *Accessor) Create(ctx context.Context, entity any, tbl string, idFields ...string) error {
	idColumns := []string{}
	colValueMap := map[string]reflect.Value{}
	var err error

	if len(idFields) == 0 {
		idFields = []string{"Id"}

		// check if default primary mapping exists
		idColumns, colValueMap, err = a.getMapping(entity, idFields...)
		if err != nil {
			idColumns, colValueMap, err = a.getMapping(entity)
			if err != nil {
				return err
			}
		}
	} else {
		idColumns, colValueMap, err = a.getMapping(entity, idFields...)
		if err != nil {
			return err
		}
	}

	colValueMap = removeNestedCols(colValueMap)
	return a.SqlizerGet(ctx, entity, func(builder squirrel.StatementBuilderType) Sqlizer {
		b := builder.Insert(tbl)

		cols := []string{}
		vals := []any{}

		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				if !v.IsZero() {
					cols = append(cols, k)
					vals = append(vals, v.Interface())
				}
			} else {
				cols = append(cols, k)
				vals = append(vals, v.Interface())
			}
		}

		return b.Columns(cols...).Values(vals...).Suffix("RETURNING *")
	})
}

// Usage example
/*
   city2 := struct {
       Id      int    `db:"id"`
       Name    string `db:"name"`
       ZipCode string `db:"zip_code"`
   }{
       Id: city.Id,
   }

   err = accessor.Read(context.Background(), &city2, "city")
*/
func (a *Accessor) Read(ctx context.Context, entity any, tbl string, idFields ...string) error {
	if len(idFields) == 0 {
		idFields = []string{"Id"}
	}

	idColumns, colValueMap, err := a.getMapping(entity, idFields...)
	if err != nil {
		return errors.New("missing ID columns")
	}

	colValueMap = removeNestedCols(colValueMap)
	return a.SqlizerGet(ctx, entity, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				eq[k] = v.Interface()
			}
		}
		return builder.Select("*").From(tbl).Where(eq)
	})
}

// Usage example:
/*
   // partial update
   p := &PersonWithUpdateTracker{}
   p.FirstName = "foo"
   p.LastName = "test"

   // this update will be tracked
   p.SetEmail("foo@test.change")

   accessor := New(s.Db.DB)
   result, err := accessor.Update(context.Background(), p, "person", "FirstName", "LastName")

   // update in full
   ppp := &Person{}
   ppp.FirstName = "foo"
   ppp.LastName = "test"
   ppp.Email = "foo@test.change.again"
   ppp.AddedAt = time.Now().UTC()
   result, err = accessor.Update(context.Background(), ppp, "person", "FirstName", "LastName")
*/
func (a *Accessor) Update(ctx context.Context, entity any, tbl string, idFields ...string) (sql.Result, error) {
	if len(idFields) == 0 {
		idFields = []string{"Id"}
	}

	idColumns, colValueMap, err := a.getMapping(entity, idFields...)
	if err != nil {
		return nil, errors.New("missing ID columns")
	}

	colValueMap = removeNestedCols(colValueMap)
	return a.SqlizerExec(ctx, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				eq[k] = v.Interface()
			}
		}

		q := builder.Update(tbl)
		if tracker, ok := entity.(UpdateTracker); ok {
			colsChanged := tracker.ColumnsChanged()

			for k, v := range colValueMap {
				if !stringInSlice(k, idColumns) && stringInSlice(k, colsChanged) {
					q = q.Set(k, v.Interface())
				}
			}
		} else {
			for k, v := range colValueMap {
				if !stringInSlice(k, idColumns) {
					q = q.Set(k, v.Interface())
				}
			}
		}

		return q.Where(eq)
	})
}

// Usage example:
/*
   p := struct {
       FirstName string    `db:"first_name"`
       LastName  string    `db:"last_name"`
       Email     string    `db:"email"`
       AddedAt   time.Time `db:"added_at"`
   }{
       FirstName: "foo",
       LastName:  "test",
   }

   accessor := New(s.Db.DB)
   result, err := accessor.Delete(context.Background(), p, "person", "FirstName", "LastName")
*/
func (a *Accessor) Delete(ctx context.Context, entity any, tbl string, idFields ...string) (sql.Result, error) {
	if len(idFields) == 0 {
		idFields = []string{"Id"}
	}

	idColumns, colValueMap, err := a.getMapping(entity, idFields...)
	if err != nil {
		return nil, errors.New("missing ID columns")
	}

	colValueMap = removeNestedCols(colValueMap)
	return a.SqlizerExec(ctx, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				eq[k] = v.Interface()
			}
		}
		return builder.Delete(tbl).Where(eq)
	})
}

func (a *Accessor) getMapping(entity any, idFields ...string) (idColumns []string, colValueMap map[string]reflect.Value, outErr error) {
	for _, idField := range idFields {
		col := Column(entity, idField)
		if col != "" {
			idColumns = append(idColumns, col)
		}
	}
	if len(idColumns) != len(idFields) {
		return nil, nil, errors.New("missing ID columns")
	}

	var mapper *reflectx.Mapper
	if db, ok := a.Db.(*sqlx.DB); ok {
		mapper = db.Mapper
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		mapper = tx.Mapper
	}

	// column (tag name) -> reflect.Value mapping
	colValueMap = mapper.FieldMap(reflect.ValueOf(entity))
	return
}

// Usage example:
/*
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
*/
func (a *Accessor) NamedGet(ctx context.Context, dest any, query string, arg any) error {
	var rows *sqlx.Rows
	var err error

	if db, ok := a.Db.(*sqlx.DB); ok {
		rows, err = db.Unsafe().NamedQueryContext(ctx, query, arg)
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		rows, err = tx.Unsafe().NamedQuery(query, arg)
	} else {
		return errors.New("invalid accessor backend")
	}

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		return rows.StructScan(dest)
	}

	return sql.ErrNoRows
}

// Usage example
/*
   var p []struct {
       FirstName string `db:"first_name"`
       LastName  string `db:"last_name"`
       Email     string `db:"email"`
   }

   err := accessor.NamedSelect(context.Background(), &p, "select * from person where last_name=:last_name order by first_name DESC",
       map[string]any{
           "last_name": "test",
       })
*/
func (a *Accessor) NamedSelect(ctx context.Context, dest any, query string, arg any) error {
	var rows *sqlx.Rows
	var err error

	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}
	direct := reflect.Indirect(value)

	slice, err := baseType(value.Type(), reflect.Slice)
	if err != nil {
		return err
	}
	direct.SetLen(0)
	isPtr := slice.Elem().Kind() == reflect.Ptr
	base := reflectx.Deref(slice.Elem())

	if db, ok := a.Db.(*sqlx.DB); ok {
		rows, err = db.Unsafe().NamedQueryContext(ctx, query, arg)
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		rows, err = tx.Unsafe().NamedQuery(query, arg)
	} else {
		return errors.New("invalid accessor backend")
	}

	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		elem := reflect.New(base)
		err := rows.StructScan(elem.Elem().Addr().Interface())
		if err != nil {
			return err
		}

		if isPtr {
			direct.Set(reflect.Append(direct, elem))
		} else {
			direct.Set(reflect.Append(direct, reflect.Indirect(elem)))
		}
	}

	return nil
}

func (a *Accessor) NamedExec(ctx context.Context, query string, arg any) (sql.Result, error) {
	if db, ok := a.Db.(*sqlx.DB); ok {
		return db.Unsafe().NamedExecContext(ctx, query, arg)
	} else if tx, ok := a.Db.(*sqlx.Tx); ok {
		return tx.Unsafe().NamedExecContext(ctx, query, arg)
	} else {
		return nil, errors.New("invalid accessor backend")
	}
}

// Usage example:
/*
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
*/
func (a *Accessor) SqlizerGet(
	ctx context.Context,
	dest any,
	sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
) error {
	q, args, err := sqlizer(squirrel.StatementBuilder).ToSql()

	if err != nil {
		return err
	}

	return a.Get(ctx, dest, a.Db.Rebind(q), args...)
}

// Usage example:
/*
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
*/
func (a *Accessor) SqlizerSelect(
	ctx context.Context,
	dest any,
	sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
) error {
	q, args, err := sqlizer(squirrel.StatementBuilder).ToSql()

	if err != nil {
		return err
	}

	return a.Select(ctx, dest, a.Db.Rebind(q), args...)
}

func (a *Accessor) SqlizerExec(
	ctx context.Context,
	sqlizer func(builder squirrel.StatementBuilderType) Sqlizer,
) (result sql.Result, outErr error) {
	q, args, err := sqlizer(squirrel.StatementBuilder).ToSql()

	if err != nil {
		return nil, err
	}

	return a.Exec(ctx, a.Db.Rebind(q), args...)
}

// ExecTx uses annonymous execution function to achieve crash-safe and implicit transaction commission effect
// Usage example
/*
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
*/
func ExecTx(
	ctx context.Context,
	db *sqlx.DB,
	txOps *sql.TxOptions,
	execFn func(ctx context.Context, accessor *Accessor) error,
) (outErr error) {
	tx, err := db.Unsafe().BeginTxx(ctx, txOps)
	if err != nil {
		return err
	}

	// make execution of transaction be crash-safe
	defer func() {
		if c := recover(); c != nil {
			_ = tx.Rollback()
			outErr = errors.New("panic error in executing transaction")
		}
	}()

	err = execFn(ctx, New(tx))
	if err != nil {
		return tx.Rollback()
	} else {
		return tx.Commit()
	}
}

// Column return the mapped column mapping name in entity type of v
// and specified Go struct field name.
// Optimizing with multi-threaded safe caching if it is needed
func Column(v any, fieldName string) string {
	typ := reflect.TypeOf(v)
	typ = reflectx.Deref(typ)
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	// first round, check direct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Name == fieldName {
			return fieldMappedColumn(field, "db")
		}
	}

	// second round, check annonymous embedded fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous {
			t := reflectx.Deref(field.Type)

			col := Column(reflect.New(t).Interface(), fieldName)
			if col != "" {
				return col
			}
		}
	}
	return ""
}

// Columns return all found column mappings in entity type of v.
// Optimizing with multi-threaded safe caching if it is needed
func Columns(v any) []string {
	typ := reflect.TypeOf(v)
	typ = reflectx.Deref(typ)
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}

	colNames := []string{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous {
			t := reflectx.Deref(field.Type)

			colNames = append(colNames, Columns(reflect.New(t).Interface())...)
		} else {
			if !strings.Contains(string(field.Tag), "db:") {
				continue
			}
			tag := field.Tag.Get("db")

			// split the options from the name
			parts := strings.Split(tag, ",")
			mappedName := parts[0]

			if mappedName == "-" {
				mappedName = ""
			}

			if mappedName != "" {
				colNames = append(colNames, mappedName)
			}
		}
	}

	return dedupe(colNames)
}

/////////////////////////////////////////////////////////////////////////////

func dedupe(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func fieldMappedColumn(
	field reflect.StructField,
	tagName string,
) string {
	if !strings.Contains(string(field.Tag), tagName+":") {
		return ""
	}

	tag := field.Tag.Get(tagName)

	// split the options from the name
	parts := strings.Split(tag, ",")
	mappedName := parts[0]

	if mappedName == "-" {
		mappedName = ""
	}

	return mappedName
}

func baseType(t reflect.Type, expected reflect.Kind) (reflect.Type, error) {
	t = reflectx.Deref(t)
	if t.Kind() != expected {
		return nil, fmt.Errorf("expected %s but got %s", expected, t.Kind())
	}
	return t, nil
}

func stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func removeNestedCols(m map[string]reflect.Value) map[string]reflect.Value {
	if m == nil {
		return nil
	}

	ret := make(map[string]reflect.Value)
	for k, v := range m {
		if !strings.Contains(k, ".") {
			ret[k] = v
		}
	}

	return ret
}
