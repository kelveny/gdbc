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
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"

	cpy "github.com/barkimedes/go-deepcopy"
)

type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

type UpdateTracker interface {
	ColumnsChanged(tbl ...string) []string
}

type EntityMappingSchema struct {
	TableName string

	// field name -> column name (note, fields in embedded type are not included here)
	Columns map[string]string

	// embedded mappings
	BaseMappings []*EntityMappingSchema

	Entity     any
	EntityType reflect.Type
}

func (m *EntityMappingSchema) Schemas() []*EntityMappingSchema {
	var schemas []*EntityMappingSchema

	for _, mm := range m.BaseMappings {
		schemas = append(schemas, mm.Schemas()...)
	}

	return append(schemas, m)
}

func (m *EntityMappingSchema) Tables() []string {
	var tables []string

	for _, mm := range m.BaseMappings {
		tables = append(tables, mm.Tables()...)
	}
	return append(tables, m.TableName)
}

func (m *EntityMappingSchema) GetColumnSelectString() string {
	var builder strings.Builder
	for i, t := range m.Tables() {
		if i == 0 {
			builder.WriteString(t + ".*")
		} else {
			builder.WriteString(", " + t + ".*")
		}
	}

	return builder.String()
}

func (m *EntityMappingSchema) GetTableJoinString(idColumns ...string) string {
	tables := m.Tables()

	var builder strings.Builder

	for i := 0; i < len(tables)-1; i++ {
		if i > 0 {
			builder.WriteString(" JOIN ")
		}
		builder.WriteString(tables[i+1] + " ON ")
		builder.WriteString(getJoinOnString(tables[i], tables[i+1], idColumns...))
	}

	return builder.String()
}

func getJoinOnString(table1, table2 string, idColumns ...string) string {
	var builder strings.Builder

	for i, id := range idColumns {
		builder.WriteString(fmt.Sprintf("%s.%s=%s.%s", table1, id, table2, id))
		if i < len(idColumns)-1 {
			builder.WriteString(" AND ")
		}
	}
	return builder.String()
}

type noopSqlResult struct {
}

func (r noopSqlResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r noopSqlResult) RowsAffected() (int64, error) {
	return 0, nil
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
	s, err := EntitySchema(entity, reflect.TypeOf(entity), tbl)
	if err != nil {
		return err
	}

	if len(s.BaseMappings) > 0 {
		return a.createComposite(ctx, s, idFields...)
	}

	return a.create(ctx, entity, s, nil, tbl, idFields...)
}

func (a *Accessor) create(
	ctx context.Context,
	entity any, s *EntityMappingSchema,
	baseColValueMap map[string]reflect.Value,
	tbl string,
	idFields ...string,
) error {
	var err error

	idColumns := []string{}
	colValueMap := map[string]reflect.Value{}

	if len(idFields) == 0 {
		idFields = []string{"Id"}

		// check if default primary mapping exists
		idColumns, colValueMap, err = a.getMapping(entity, idFields...)
		if err != nil {
			// default id column does not exist, continue insertion without it
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
	if len(baseColValueMap) == 0 {
		baseColValueMap = colValueMap
	}

	colLookup := map[string]string{}
	for field, col := range s.Columns {
		colLookup[col] = field
	}

	return a.SqlizerGet(ctx, entity, func(builder squirrel.StatementBuilderType) Sqlizer {
		b := builder.Insert(tbl)
		cols, vals := buildCreateMapping(idColumns, colLookup, baseColValueMap, colValueMap)
		return b.Columns(cols...).Values(vals...).Suffix("RETURNING *")
	})
}

func buildCreateMapping(
	idColumns []string,
	colFieldLookup map[string]string,
	baseColValueMap map[string]reflect.Value,
	colValueMap map[string]reflect.Value,
) ([]string, []any) {
	cols := []string{}
	vals := []any{}

	for k, v := range baseColValueMap {
		if stringInSlice(k, idColumns) {
			if !v.IsZero() {
				cols = append(cols, k)
				vals = append(vals, getDriverValue(v))
			}
		}
	}

	for k, v := range colValueMap {
		if !stringInSlice(k, idColumns) {
			if _, ok := colFieldLookup[k]; ok {
				cols = append(cols, k)
				vals = append(vals, getDriverValue(v))
			}
		}
	}

	return cols, vals
}

func createPointerValue(original reflect.Value) reflect.Value {
	originalType := original.Type()
	pointerType := reflect.PtrTo(originalType)

	pointerValue := reflect.New(pointerType.Elem())
	pointerValue.Elem().Set(original)

	return pointerValue
}

func (a *Accessor) createComposite(ctx context.Context, s *EntityMappingSchema, idFields ...string) error {
	var baseColValueMap map[string]reflect.Value

	var err error
	for i, mm := range s.Schemas() {
		if i < len(s.Schemas())-1 {
			var pEntity reflect.Value
			c, err := cpy.Anything(mm.Entity)
			if err != nil {
				return err
			}

			pEntity = createPointerValue(reflect.Indirect(reflect.ValueOf(c)))
			err = a.create(ctx, pEntity.Interface(), mm, baseColValueMap, mm.TableName, idFields...)
			if err != nil {
				return err
			}

			// capture possible returned auto-increment id values
			if i == 0 {
				_, baseColValueMap, err = a.getMapping(pEntity.Interface(), idFields...)
				if err != nil {
					return err
				}
			}
		} else {
			err = a.create(ctx, mm.Entity, mm, baseColValueMap, mm.TableName, idFields...)
		}

		if err != nil {
			return err
		}
	}

	return nil
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

	s, err := EntitySchema(entity, reflect.TypeOf(entity), tbl)
	if err != nil {
		return err
	}

	if len(s.BaseMappings) > 0 {
		return a.readComposite(ctx, s, idFields...)
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
				eq[k] = getDriverValue(v)
			}
		}
		return builder.Select("*").From(tbl).Where(eq)
	})
}

func (a *Accessor) readComposite(ctx context.Context, s *EntityMappingSchema, idFields ...string) error {
	idColumns, colValueMap, err := a.getMapping(s.Entity, idFields...)
	if err != nil {
		return err
	}

	tables := s.Tables()
	return a.SqlizerGet(ctx, s.Entity, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				eq[fmt.Sprintf("%s.%s", tables[0], k)] = getDriverValue(v)
			}
		}
		return builder.
			Select(s.GetColumnSelectString()).
			From(tables[0]).
			Join(s.GetTableJoinString(idColumns...)).
			Where(eq)
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

	s, err := EntitySchema(entity, reflect.TypeOf(entity), tbl)
	if err != nil {
		return nil, err
	}

	if len(s.BaseMappings) > 0 {
		return a.updateComposite(ctx, s, entity.(UpdateTracker), idFields...)
	}

	idColumns, colValueMap, err := a.getMapping(entity, idFields...)
	if err != nil {
		return nil, errors.New("missing ID columns")
	}

	tracker, _ := entity.(UpdateTracker)

	if tracker != nil && len(tracker.ColumnsChanged(s.TableName)) == 0 {
		return noopSqlResult{}, nil
	}

	colFieldLookup := map[string]string{}
	for field, col := range s.Columns {
		colFieldLookup[col] = field
	}

	return a.execUpdate(
		ctx,
		idColumns,
		colFieldLookup,
		colValueMap,
		colValueMap,
		s.TableName,
		tracker,
	)
}

func (a *Accessor) execUpdate(
	ctx context.Context,
	idColumns []string,
	colFieldLookup map[string]string,
	baseColValueMap map[string]reflect.Value,
	colValueMap map[string]reflect.Value,
	tbl string,
	tracker UpdateTracker,
) (sql.Result, error) {
	colValueMap = removeNestedCols(colValueMap)

	return a.SqlizerExec(ctx, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range baseColValueMap {
			if stringInSlice(k, idColumns) {
				eq[k] = getDriverValue(v)
			}
		}

		q := builder.Update(tbl)
		if tracker != nil {
			colsChanged := tracker.ColumnsChanged(tbl)

			for k, v := range colValueMap {
				if !stringInSlice(k, idColumns) && colFieldLookup[k] != "" && stringInSlice(k, colsChanged) {
					q = q.Set(k, getDriverValue(v))
				}
			}
		} else {
			// perform full update
			for k, v := range colValueMap {
				if !stringInSlice(k, idColumns) && colFieldLookup[k] != "" {
					q = q.Set(k, getDriverValue(v))
				}
			}
		}

		return q.Where(eq)
	})
}

func (a *Accessor) updateComposite(
	ctx context.Context,
	s *EntityMappingSchema,
	tracker UpdateTracker,
	idFields ...string,
) (sql.Result, error) {
	var baseColValueMap map[string]reflect.Value
	var idColumns []string

	var result sql.Result

	for i, m := range s.Schemas() {
		if i < len(s.Schemas())-1 {
			var pEntity reflect.Value
			c, err := cpy.Anything(m.Entity)
			if err != nil {
				return nil, err
			}

			pEntity = createPointerValue(reflect.Indirect(reflect.ValueOf(c)))

			// capture possible returned auto-increment id values
			if i == 0 {
				idColumns, baseColValueMap, err = a.getMapping(pEntity.Interface(), idFields...)
				if err != nil {
					return nil, err
				}
			}

			if tracker != nil && len(tracker.ColumnsChanged(m.TableName)) == 0 {
				continue
			}

			colFieldLookup := map[string]string{}
			for field, col := range m.Columns {
				colFieldLookup[col] = field
			}

			_, err = a.execUpdate(
				ctx,
				idColumns,
				colFieldLookup,
				baseColValueMap,
				baseColValueMap,
				m.TableName,
				tracker,
			)
			if err != nil {
				return nil, err
			}
		} else {
			idColumns, colValueMap, err := a.getMapping(m.Entity, idFields...)
			if err != nil {
				return nil, err
			}

			if tracker != nil && len(tracker.ColumnsChanged(m.TableName)) == 0 {
				continue
			}

			colFieldLookup := map[string]string{}
			for field, col := range m.Columns {
				colFieldLookup[col] = field
			}

			result, err = a.execUpdate(
				ctx,
				idColumns,
				colFieldLookup,
				baseColValueMap,
				colValueMap,
				m.TableName,
				tracker,
			)
			if err != nil {
				return nil, err
			}
		}
	}

	return result, nil
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
	// if entity is intended to read back before delete
	if reflect.TypeOf(entity).Kind() == reflect.Ptr {
		_ = a.Read(ctx, entity, tbl, idFields...)
	}

	if len(idFields) == 0 {
		idFields = []string{"Id"}
	}

	s, err := EntitySchema(entity, reflect.TypeOf(entity), tbl)
	if err != nil {
		return nil, err
	}

	if len(s.BaseMappings) > 0 {
		return a.deleteComposite(ctx, s, idFields...)
	}

	idColumns, colValueMap, err := a.getMapping(entity, idFields...)
	if err != nil {
		return nil, errors.New("missing ID columns")
	}

	return a.execDelete(ctx, colValueMap, tbl, idColumns)
}

func (a *Accessor) execDelete(
	ctx context.Context,
	colValueMap map[string]reflect.Value,
	tbl string,
	idColumns []string,
) (sql.Result, error) {
	colValueMap = removeNestedCols(colValueMap)
	return a.SqlizerExec(ctx, func(builder squirrel.StatementBuilderType) Sqlizer {
		eq := squirrel.Eq{}
		for k, v := range colValueMap {
			if stringInSlice(k, idColumns) {
				eq[k] = getDriverValue(v)
			}
		}
		return builder.Delete(tbl).Where(eq)
	})
}

func (a *Accessor) deleteComposite(ctx context.Context, s *EntityMappingSchema, idFields ...string) (sql.Result, error) {
	var err error
	var r sql.Result

	schemas := s.Schemas()
	for i := len(schemas) - 1; i >= 0; i-- {
		m := schemas[i]

		var pEntity reflect.Value
		c, err := cpy.Anything(m.Entity)
		if err != nil {
			return nil, err
		}
		pEntity = createPointerValue(reflect.Indirect(reflect.ValueOf(c)))

		idColumns, colValueMap, err := a.getMapping(pEntity.Interface(), idFields...)
		if err != nil {
			return nil, errors.New("missing ID columns")
		}

		r, err = a.execDelete(ctx, colValueMap, m.TableName, idColumns)
		if err != nil {
			return nil, err
		}
	}

	return r, err
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

	// Note: mapper.FieldMap does not support the case when entity points to an embedded type
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

func entityType(typ reflect.Type) reflect.Type {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if !field.Anonymous {
			_, _, err := fieldMappedColumnWithAttributes(field, "db")
			if err == nil {
				return typ
			}
		}
	}

	if typ.NumField() > 0 {
		field := typ.Field(0)
		if field.Anonymous {
			typ = reflectx.Deref(field.Type)
			if typ.Kind() == reflect.Struct {
				return entityType(typ)
			}
		}
	}

	return nil
}

func EntitySchema(v any, typ reflect.Type, tableName string) (*EntityMappingSchema, error) {
	m := EntityMappingSchema{
		TableName:  tableName,
		Entity:     v,
		EntityType: reflectx.Deref(typ),
		Columns:    map[string]string{},
	}

	typ = reflectx.Deref(typ)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %s should be struct", typ.Name())
	}

	typ = entityType(typ)
	if typ == nil {
		return nil, fmt.Errorf("type %s should be in compliance with entity type", m.EntityType.Name())
	}

	// second round, check annonymous embedded fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Anonymous {
			ft := reflectx.Deref(field.Type)
			if ft.Kind() != reflect.Struct {
				return nil, fmt.Errorf("embedded type %s in type %s should be struct", ft.Name(), typ.Name())
			}

			col, attrs, err := fieldMappedColumnWithAttributes(field, "db")
			if err != nil {
				return nil, err
			}

			if col != "" {
				return nil, fmt.Errorf("embedded type %s in type %s should have empty column name", ft.Name(), typ.Name())
			}

			if baseTable, ok := attrs["table"]; ok {
				if baseTable == tableName {
					return nil, fmt.Errorf("embedded type %s in type %s should not have the same table mapping", ft.Name(), typ.Name())
				}

				fieldValue := reflect.Indirect(reflect.ValueOf(v)).Field(i)
				baseSchema, err := EntitySchema(fieldValue.Interface(), ft, baseTable)

				if err != nil {
					return nil, err
				}

				m.BaseMappings = append(m.BaseMappings, baseSchema)
			} else {
				return nil, fmt.Errorf("embedded type %s in type %s should have table attribute", ft.Name(), typ.Name())
			}

		} else {
			col, _, err := fieldMappedColumnWithAttributes(field, "db")
			if err != nil {
				return nil, err
			}

			if col != "" {
				m.Columns[field.Name] = col
			}
		}
	}

	return &m, nil
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

func fieldMappedColumnWithAttributes(
	field reflect.StructField,
	tagName string,
) (string, map[string]string, error) {
	attrs := map[string]string{}

	if !strings.Contains(string(field.Tag), tagName+":") {
		return "", attrs, fmt.Errorf("%s does not exist", tagName)
	}

	tag := field.Tag.Get(tagName)

	// split the options from the name
	parts := strings.Split(tag, ",")
	mappedName := parts[0]

	if mappedName == "-" {
		mappedName = ""
	}

	if len(parts) > 1 {
		for i := 1; i < len(parts); i++ {
			tokens := strings.Split(strings.Trim(parts[i], " "), "=")
			if len(tokens) == 2 {
				attrs[tokens[0]] = tokens[1]
			} else {
				attrs[tokens[0]] = ""
			}
		}
	}

	return mappedName, attrs, nil
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

func getDriverValue(val reflect.Value) any {
	var v any

	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil
		}

		v = val.Elem().Interface()
		if reflect.ValueOf(v).IsZero() {
			return nil
		}
	} else {
		v = val.Interface()
	}

	if v != nil {
		if valuer, ok := v.(driver.Valuer); ok {
			val, err := valuer.Value()
			if err == nil {
				// deal with special nil value
				if val == "<nil>" {
					return nil
				}
				return val
			}
		}
	}

	return v
}
