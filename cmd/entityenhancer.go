package cmd

import (
	"bufio"
	"bytes"
	"errors"
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
	trackMap map[string]map[string]bool
}

func (e *{{ .Entity }}WithUpdateTracker) registerChange(tbl string, col string) {
	if e.trackMap == nil {
		e.trackMap = make(map[string]map[string]bool)
	}

	if m, ok := e.trackMap[tbl]; ok {
		m[col] = true
	} else {
		m = make(map[string]bool)
		e.trackMap[tbl] = m

		m[col] = true
	}
}

func (e *{{ .Entity }}WithUpdateTracker) ColumnsChanged(tbl ...string) []string {
    cols := []string{}

	if tbl == nil {
		tbl = []string{"{{ .Table }}"}
	}

	if e.trackMap != nil {
		m := e.trackMap[tbl[0]]
		for col := range m {
			cols = append(cols, col)
		}
	}

    return cols
}

{{- with $root := . }}

{{ range $index, $f := .Fields }}
func (e *{{ $root.Entity }}WithUpdateTracker) Set{{ $f.Name }}(val {{ $f.TypeDecl }}) *{{ $root.Entity }}WithUpdateTracker {
    e.{{ $f.Name }} = val
	e.registerChange("{{ $root.Table }}", "{{ $f.Column }}")
    return e
}
{{ end }}

{{ range $i, $base := .BaseFields }}
{{ range $j, $f := $base.Fields }}

func (e *{{ $root.Entity }}WithUpdateTracker) Set{{ $f.Name }}(val {{ $f.TypeDecl }}) *{{ $root.Entity }}WithUpdateTracker {
	e.{{ $f.Name }} = val
	e.registerChange("{{ $base.Table }}", "{{ $f.Column }}")
	return e
}

{{ end }}
{{ end }}

{{ end }}

`
)

/////////////////////////////////////////////////////////////////////////////
type ModuleSpec struct {
	RootModulePath string
	RootDir        string
}

func getModulePath(gomodPath string) (string, error) {
	file, err := os.Open(gomodPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			return modulePath, nil
		}
	}

	return "", errors.New("invalid go.mod")
}

func (m *ModuleSpec) Init() error {
	root, err := filepath.Abs("")
	if err != nil {
		return err
	}

	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}

		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}

	m.RootDir = root
	m.RootModulePath, err = getModulePath(filepath.Join(m.RootDir, "go.mod"))
	return err
}

func (m *ModuleSpec) ResolveModulePath(modulePath string) (string, error) {
	if strings.HasPrefix(modulePath, m.RootModulePath) {
		return filepath.Join(m.RootDir, strings.TrimPrefix(modulePath, m.RootModulePath)), nil
	}

	return "", errors.New("external module is not supported")
}

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

	PkgHostDir string             // package full path in host in which the entity type is defined
	Imports    []gogen.ImportSpec // imports from the file in which the entity type is defined
}

var entityRegistry map[string]*EntitySpec

func registerEntitySpec(entitySpec *EntitySpec) {
	if entityRegistry == nil {
		entityRegistry = make(map[string]*EntitySpec)
	}

	key := getEntityRegistrationKey(entitySpec.PkgHostDir, entitySpec.Name)
	if _, ok := entityRegistry[key]; !ok {
		entityRegistry[key] = entitySpec
	}
}

func lookupEntitySpec(entityHostPath string, entityName string) *EntitySpec {
	if entityRegistry != nil {
		return entityRegistry[getEntityRegistrationKey(entityHostPath, entityName)]
	}

	return nil
}

func getEntityRegistrationKey(entityHostPath string, entityName string) string {
	return strings.Join([]string{
		entityName,
		entityHostPath,
	}, "@")
}

func resolveImportHostDir(modSpec *ModuleSpec, imports []gogen.ImportSpec, alias string) (string, error) {
	modPath := ""
	for _, importSpec := range imports {
		if importSpec.Name == alias || importSpec.Name == "" && filepath.Base(importSpec.Path) == alias {
			modPath = importSpec.Path
			break
		}
	}

	return modSpec.ResolveModulePath(modPath)
}

func (es *EntitySpec) FlattenFieldSpecs(
	modSpec *ModuleSpec,
	pkgDir string,
	tbl string,
	fieldSpecs map[string][]EntityFieldSpec,
	importSpecs map[string][]gogen.ImportSpec,
) (tables []string) {
	tables = append(tables, tbl)
	fieldSpecs[tbl] = es.FieldSpecs
	importSpecs[tbl] = es.Imports

	for _, field := range es.TypeSpec.Fields.List {
		if field.Tag != nil {
			if strings.HasPrefix(field.Tag.Value, "`db:") {
				tagValue := strings.Split(strings.Trim(field.Tag.Value, "`"), ":")
				attrs := strings.Split(strings.Trim(tagValue[1], "\""), ",")
				if len(attrs) > 1 {
					for _, v := range attrs[1:] {
						kv := strings.Split(v, "=")
						if len(kv) > 1 {
							if strings.Trim(kv[0], " ") == "table" {
								tbl := strings.Trim(kv[1], " ")

								name := gosyntax.ExprDeclString(es.TokenFset, field.Type)
								name = strings.Trim(name, "*") // remove pointer declaration from name

								tokens := strings.Split(name, ".")
								hostDir := pkgDir
								var err error
								if len(tokens) > 1 {
									name = tokens[1]
									hostDir, err = resolveImportHostDir(modSpec, es.Imports, tokens[0])
									if err != nil {
										logger.Log(logger.ERROR, "Can not resolve import of package %s in %s", tokens[0], pkgDir)
									}
								}

								baseSpec := lookupEntitySpec(hostDir, name)
								if baseSpec == nil {
									// perform lazy registration for cross-package code generation
									scanDir(modSpec, hostDir, scanPredicate, buildEntityRegistry)
								}

								baseSpec = lookupEntitySpec(hostDir, name)
								if baseSpec != nil {
									tables = append(tables, baseSpec.FlattenFieldSpecs(modSpec, baseSpec.PkgHostDir, tbl, fieldSpecs, importSpecs)...)
								} else {
									logger.Log(logger.ERROR, "Can not find entity %s in %s", name, hostDir)
								}
							}
						}
					}
				}
			}
		}
	}
	return
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
	modSpec *ModuleSpec,
	pkgDir string,
	predicate func(fi os.FileInfo) bool,
	do func(modSpec *ModuleSpec, pkgDir string, fi os.FileInfo),
) error {
	p, err := filepath.Abs(pkgDir)
	if err != nil {
		return err
	}
	pkgDir = p

	if dir, err := os.Stat(pkgDir); err == nil && dir.IsDir() {
		fileInfos, err := ioutil.ReadDir(pkgDir)
		if err != nil {
			return err
		}

		for _, fileInfo := range fileInfos {
			if predicate(fileInfo) {
				do(modSpec, pkgDir, fileInfo)
			}
		}
	}
	return nil
}

func buildEntityRegistry(modSpec *ModuleSpec, pkgDir string, fi os.FileInfo) {
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

	scanToBuildEntityRegistry(modSpec, pkgDir, fi, fset, file)
}

func generateWithConfig(config *Config) func(*ModuleSpec, string, os.FileInfo) {
	return func(modSpec *ModuleSpec, pkgDir string, fi os.FileInfo) {
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
		scanToEnhanceEntities(modSpec, pkgDir, fi, fset, file, config)
		logger.Log(logger.PROMPT, "Done entity enhancement for %s \n", fi.Name())
	}
}

func scanToBuildEntityRegistry(
	_ *ModuleSpec,
	path string,
	_ os.FileInfo,
	fset *token.FileSet,
	file *ast.File,
) {
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
								PkgHostDir: path,
								Imports:    gogen.GetFileImports(file),
							}

							registerEntitySpec(&entitySpec)
						}
					}
				}
			}
		}
	}
}

func scanToEnhanceEntities(
	modSpec *ModuleSpec,
	pkgDir string,
	fi os.FileInfo,
	fset *token.FileSet,
	file *ast.File,
	config *Config,
) {
	for _, d := range file.Decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if tspec, ok := spec.(*ast.TypeSpec); ok {
					if entity, ok := tspec.Type.(*ast.StructType); ok {
						option := config.GetEntityOption(tspec.Name.Name)

						if option != nil {
							enhanceEntity(modSpec, pkgDir, fi, fset, file, entity, option)
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

func isImportSpecInSlice(slice []gogen.ImportSpec, spec gogen.ImportSpec) bool {
	for _, specInSlice := range slice {
		if specInSlice.Name == spec.Name && specInSlice.Path == spec.Path {
			return true
		}
	}

	return false
}

func mergeBaseImports(
	imports []gogen.ImportSpec,
	tables []string,
	flattenImports map[string][]gogen.ImportSpec,
) []gogen.ImportSpec {
	for i, tbl := range tables {
		// process imports from base entities only
		if i > 0 {
			specs := flattenImports[tbl]
			for _, specInBase := range specs {
				if !isImportSpecInSlice(imports, specInBase) {
					imports = append(imports, specInBase)
				}
			}
		}
	}

	return imports
}

func generate(
	writer io.Writer,
	pkgDir string,
	file *ast.File,
	_ *ast.StructType,
	tables []string,
	fields map[string][]EntityFieldSpec,
	flattenImports map[string][]gogen.ImportSpec,
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

	if len(tables) > 1 {
		imports = mergeBaseImports(imports, tables, flattenImports)
	}
	gogen.WriteImportDecls(writer, imports)

	baseFields := []struct {
		Table  string
		Fields []EntityFieldSpec
	}{}

	if len(tables) > 1 {
		for _, tbl := range tables[1:] {
			baseFields = append(baseFields, struct {
				Table  string
				Fields []EntityFieldSpec
			}{
				Table:  tbl,
				Fields: fields[tbl],
			})
		}
	}

	// generate code
	binding := struct {
		Entity     string
		Table      string
		Fields     []EntityFieldSpec
		BaseFields []struct {
			Table  string
			Fields []EntityFieldSpec
		}
	}{
		Entity:     option.Entity,
		Table:      option.Table,
		Fields:     fields[option.Table],
		BaseFields: baseFields,
	}
	t := template.Must(template.New("EntityEnhancer").
		Parse(entityenhancerTemplate))
	return t.Execute(writer, binding)
}

func enhanceEntity(
	modSpec *ModuleSpec,
	pkgDir string,
	fi os.FileInfo,
	fset *token.FileSet,
	file *ast.File,
	entity *ast.StructType,
	option *Option,
) {
	entitySpec := lookupEntitySpec(pkgDir, option.Entity)
	flattenFields := map[string][]EntityFieldSpec{}
	flattenImports := map[string][]gogen.ImportSpec{}
	tables := entitySpec.FlattenFieldSpecs(modSpec, pkgDir, option.Table, flattenFields, flattenImports)

	if len(flattenFields) > 0 {
		var outputFileName string

		// first pass to generate in memory
		var buf bytes.Buffer
		if err := generate(&buf, pkgDir, file, entity, tables, flattenFields, flattenImports, option, false); err != nil {
			logger.Log(logger.ERROR, "Code generation error %s\n", err)
			return
		}

		file, err := parser.ParseFile(fset, "", buf.Bytes(), parser.ParseComments)
		if err != nil {
			logger.Log(logger.ERROR, "Code generation error %s\n", err)
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

		if err := generate(output, pkgDir, file, entity, tables, flattenFields, flattenImports, option, true); err != nil {
			logger.Log(logger.ERROR, "Code generation error in cleaning imports%s\n", err)

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

func Execute() {
	entityName := flag.String("entity", "", "name of the entity type")
	tableName := flag.String("table", "", "name of the database table that entity type is associated")
	path := flag.String("path", "", "path to scan")

	flag.Parse()

	if *entityName == "" || *tableName == "" {
		logger.Log(logger.ERROR, "Need valid entity and table names\n")
		os.Exit(1)
	}

	modSpec := &ModuleSpec{}
	_ = modSpec.Init()

	config := &Config{
		Entityenhancer: []*Option{
			{
				Entity: *entityName,
				Table:  *tableName,
			},
		},
	}

	scanDir(modSpec, *path, scanPredicate, buildEntityRegistry)

	// TODO we can scan from the entity registry to save some time
	scanDir(modSpec, *path, scanPredicate, generateWithConfig(config))
}

/////////////////////////////////////////////////////////////////////////////
