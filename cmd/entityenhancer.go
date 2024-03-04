package cmd

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/kelveny/gdbc/pkg/logger"
	"github.com/kelveny/mockcompose/pkg/gofile"
	"github.com/kelveny/mockcompose/pkg/gogen"
	"github.com/kelveny/mockcompose/pkg/gosyntax"
)

var SemVer = "v0.0.0-devel"

func GetSemverInfo() string {
	return SemVer
}

/////////////////////////////////////////////////////////////////////////////

const (
	header = `// CODE GENERATED AUTOMATICALLY WITH github.com/kelveny/gdbc entity enhancer
// THIS FILE SHOULD NOT BE EDITED BY HAND
package %s

`
	entityenhancerTemplate = `
type {{ .Entity }}EntityFields struct {
    {{- range $index, $f := .Fields }}
    {{ $f.Name }} string
    {{- end }}
}

type {{ .Entity }}TableColumns struct {
    {{- range $index, $f := .Fields }}
    {{ $f.Name }} string
    {{- end }}
}

func (e *{{ .Entity }}) TableName() string {
    return "{{ .Table }}"
}

func (e *{{ .Entity }}) EntityFields() *{{ .Entity }}EntityFields {
    return &{{ .Entity }}EntityFields {
        {{- range $index, $f := .Fields }}
        {{ $f.Name }}: "{{ $f.Name }}",
        {{- end }}    
    }
}

func (e *{{ .Entity }}) TableColumns() *{{ .Entity }}TableColumns {
    return &{{ .Entity }}TableColumns {
        {{- range $index, $f := .Fields }}
        {{ $f.Name }}: "{{ $f.Column }}",
        {{- end }}    
    }
}

type {{ .Entity }}WithUpdateTracker struct {
    {{ .Entity }}
    trackMap map[string]bool
}

func (e *{{ .Entity }}WithUpdateTracker) ColumnsChanged() []string {
    cols := []string{}

    for col, _ := range e.trackMap {
        cols = append(cols, col)
    }

    return cols
}

{{- with $root := . }}
{{ range $index, $f := .Fields }}
func (e *{{ $root.Entity }}WithUpdateTracker) Set{{ $f.Name }}(val {{ $f.TypeDecl }}) *{{ $root.Entity }}WithUpdateTracker {
    e.{{ $f.Name }} = val

    if e.trackMap == nil {
        e.trackMap = make(map[string]bool)
    }

    e.trackMap["{{ $f.Column }}"] = true

    return e
}
{{ end }}
{{ end }}
`
)

/////////////////////////////////////////////////////////////////////////////

type EntityFieldSpec struct {
	Name     string
	Column   string
	TypeDecl string
}

type EntitySpec struct {
	Name string

	TokenFset  *token.FileSet // original token fset
	TypeSpec   *ast.StructType
	FieldSpecs []EntityFieldSpec
}

var entityRegistry map[string]*EntitySpec

func registerEntitySpec(entitySpec *EntitySpec) {
	if entityRegistry == nil {
		entityRegistry = make(map[string]*EntitySpec)
	}

	if _, ok := entityRegistry[entitySpec.Name]; !ok {
		entityRegistry[entitySpec.Name] = entitySpec
	}
}

func lookupEntitySpec(entityName string) *EntitySpec {
	if entityRegistry != nil {
		return entityRegistry[entityName]
	}

	return nil
}

func (es *EntitySpec) FlattenFieldSpecs(tbl string, fieldSpecs map[string][]EntityFieldSpec) {
	fieldSpecs[tbl] = es.FieldSpecs

	for _, field := range es.TypeSpec.Fields.List {
		if field.Tag != nil {
			if strings.HasPrefix(field.Tag.Value, "`db:") {
				tagValue := strings.Split(field.Tag.Value, ":")
				attrs := strings.Split(tagValue[1], ",")
				if len(attrs) > 1 {
					for _, v := range attrs[1:] {
						kv := strings.Split(v, "=")
						if len(kv) > 1 {
							if strings.Trim(kv[0], " ") == "table" {
								tbl := strings.Trim(kv[1], " ")

								name := gosyntax.ExprDeclString(es.TokenFset, field.Type)
								name = strings.Trim(name, "*") // remove pointer declaration from name

								baseSpec := lookupEntitySpec(name)
								if baseSpec != nil {
									baseSpec.FlattenFieldSpecs(tbl, fieldSpecs)
								}
							}
						}
					}
				}
			}
		}
	}
}

/////////////////////////////////////////////////////////////////////////////

// must be public for it to be used in loading YAML configuration
type Option struct {
	Entity string `yaml:"entity"`
	Table  string `yaml:"table"`
}

type Config struct {
	Entityenhancer []*Option `yaml:"entityenhancer,flow"`
}

func (c *Config) GetEntityOption(name string) *Option {
	if c == nil {
		return nil
	}

	for _, option := range c.Entityenhancer {
		if option.Entity == name {
			return option
		}
	}

	return nil
}

/////////////////////////////////////////////////////////////////////////////

func scanPredicate(fi os.FileInfo) bool {
	if strings.HasSuffix(fi.Name(), ".go") &&
		!strings.HasSuffix(fi.Name(), "_enhanced.go") {

		return true
	}

	return false
}

func scanDir(
	pkgDir string,
	predicate func(fi os.FileInfo) bool,
	do func(pkgDir string, fi os.FileInfo),
) error {
	if pkgDir == "" {
		p, err := filepath.Abs("")
		if err != nil {
			return err
		}
		pkgDir = p
	}

	if dir, err := os.Stat(pkgDir); err == nil && dir.IsDir() {
		fileInfos, err := ioutil.ReadDir(pkgDir)
		if err != nil {
			return err
		}

		for _, fileInfo := range fileInfos {
			if predicate(fileInfo) {
				do(pkgDir, fileInfo)
			}
		}
	}
	return nil
}

func buildEntityRegistry(pkgDir string, fi os.FileInfo) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset,
		filepath.Join(pkgDir, fi.Name()),
		nil,
		parser.ParseComments)

	if err != nil {
		logger.Log(logger.ERROR, "Error in parsing %s, error: %s\n",
			filepath.Join(pkgDir, fi.Name()), err,
		)
		return
	}

	scanToBuildEntityRegistry(pkgDir, fi, fset, file)
}

func generateWithConfig(config *Config) func(string, os.FileInfo) {
	return func(pkgDir string, fi os.FileInfo) {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(
			fset,
			filepath.Join(pkgDir, fi.Name()),
			nil,
			parser.ParseComments)

		if err != nil {
			logger.Log(logger.ERROR, "Error in parsing %s, error: %s\n",
				filepath.Join(pkgDir, fi.Name()), err,
			)
			return
		}

		logger.Log(logger.PROMPT, "Scan %s... \n", fi.Name())
		scanToEnhanceEntities(pkgDir, fi, fset, file, config)
		logger.Log(logger.PROMPT, "Done entity enhancement for %s \n", fi.Name())
	}
}

func scanToBuildEntityRegistry(_ string, _ os.FileInfo, fset *token.FileSet, file *ast.File) {
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if tspec, ok := spec.(*ast.TypeSpec); ok {
					if entity, ok := tspec.Type.(*ast.StructType); ok {
						if isEntityStruct(entity) {
							entitySpec := EntitySpec{
								Name:       tspec.Name.Name,
								TokenFset:  fset,
								TypeSpec:   entity,
								FieldSpecs: getEntityFields(fset, entity),
							}

							registerEntitySpec(&entitySpec)
						}
					}
				}
			}
		}
	}
}

func scanToEnhanceEntities(pkgDir string, fi os.FileInfo, fset *token.FileSet, file *ast.File, config *Config) {
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if tspec, ok := spec.(*ast.TypeSpec); ok {
					if entity, ok := tspec.Type.(*ast.StructType); ok {
						option := config.GetEntityOption(tspec.Name.Name)

						if option != nil {
							enhanceEntity(pkgDir, fi, fset, file, entity, option)
						}
					}
				}
			}
		}
	}
}

func pkgNameFromDir(pkgDir string) string {
	return strings.ToLower(filepath.Base(pkgDir))
}

func columnFromTag(tag string) string {
	if strings.HasPrefix(tag, "`db:") {
		s, err := strconv.Unquote(tag)
		if err != nil {
			return ""
		}

		// trim prefix
		s = strings.TrimPrefix(s, "db:")

		s, err = strconv.Unquote(s)
		if err != nil {
			return ""
		}

		tokens := strings.Split(s, ",")

		if tokens[0] != "-" {
			return strings.Trim(tokens[0], " ")
		}
	}

	return ""
}

func isEntityStruct(st *ast.StructType) bool {
	for _, field := range st.Fields.List {
		if field.Tag != nil {
			if strings.HasPrefix(field.Tag.Value, "`db:") {
				return true
			}
		}
	}
	return false
}

func getEntityFields(fset *token.FileSet, entity *ast.StructType) []EntityFieldSpec {
	fields := []EntityFieldSpec{}

	for _, field := range entity.Fields.List {
		if field.Tag != nil {
			col := columnFromTag(field.Tag.Value)
			if col != "" {
				fields = append(fields, EntityFieldSpec{
					Name:     field.Names[0].Name,
					Column:   col,
					TypeDecl: gosyntax.ExprDeclString(fset, field.Type),
				})
			}
		}
	}

	return fields
}

func generate(
	writer io.Writer,
	pkgDir string,
	file *ast.File,
	entity *ast.StructType,
	fields []EntityFieldSpec,
	option *Option,
	cleanImports bool,
) error {
	// write package line
	_, err := writer.Write([]byte(fmt.Sprintf(header, pkgNameFromDir(pkgDir))))
	if err != nil {
		return nil
	}

	imports := gogen.GetFileImports(file)
	if cleanImports {
		imports = gogen.CleanImports(file, nil)
	}
	gogen.WriteImportDecls(writer, imports)

	// generate code
	binding := struct {
		Entity string
		Table  string
		Fields []EntityFieldSpec
	}{
		Entity: option.Entity,
		Table:  option.Table,
		Fields: fields,
	}
	t := template.Must(template.New("EntityEnhancer").
		Parse(entityenhancerTemplate))
	return t.Execute(writer, binding)
}

func enhanceEntity(pkgDir string, fi os.FileInfo, fset *token.FileSet, file *ast.File, entity *ast.StructType, option *Option) {
	fields := getEntityFields(fset, entity)
	if len(fields) > 0 {
		var outputFileName string

		// first pass to generate in memory
		var buf bytes.Buffer
		if err := generate(&buf, pkgDir, file, entity, fields, option, false); err != nil {
			return
		}

		file, err := parser.ParseFile(fset, "", buf.Bytes(), parser.ParseComments)
		if err != nil {
			return
		}

		// second pass to generate output file with imports cleaned and formatted
		if strings.HasSuffix(fi.Name(), "_test.go") {
			outputFileName = fmt.Sprintf("%s_enhanced_test.go", strings.ToLower(option.Entity))
		} else {
			outputFileName = fmt.Sprintf("%s_enhanced.go", strings.ToLower(option.Entity))
		}
		output, err := os.OpenFile(
			filepath.Join(pkgDir, outputFileName),
			os.O_CREATE|os.O_RDWR,
			0644)
		if err != nil {
			logger.Log(logger.ERROR, "Error in creating %s, error: %s\n",
				outputFileName, err,
			)

			return
		}

		if err := generate(output, pkgDir, file, entity, fields, option, true); err != nil {
			output.Close()
			return
		}

		offset, err := output.Seek(0, io.SeekCurrent)
		if err != nil {
			logger.Log(logger.ERROR, "Error in file operation on %s, error: %s\n", outputFileName, err)
		} else {
			fi, _ := output.Stat()
			if offset > 0 && offset < fi.Size() {
				output.Truncate(offset)
			}
		}
		output.Close()

		gofile.FormatGoFile(filepath.Join(pkgDir, outputFileName))
	}
}

/////////////////////////////////////////////////////////////////////////////

func usage() {
	logger.Log(logger.PROMPT, `Usage: %s [-help] [options]

entityenhancer generates entity type meta types and enhanced entity type for tracking updates.
`, os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func Execute() {
	entityName := flag.String("entity", "", "name of the entity type")
	tableName := flag.String("table", "", "name of the database table that entity type is associated")
	flag.Parse()

	if *entityName == "" || *tableName == "" {
		logger.Log(logger.ERROR, "Need valid entity and table names\n")
		os.Exit(1)
	}

	config := &Config{
		Entityenhancer: []*Option{
			{
				Entity: *entityName,
				Table:  *tableName,
			},
		},
	}

	scanDir("", scanPredicate, buildEntityRegistry)

	// TODO we can scan from the entity registry now
	scanDir("", scanPredicate, generateWithConfig(config))
}

/////////////////////////////////////////////////////////////////////////////
