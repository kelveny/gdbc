package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/kelveny/mockcompose/pkg/gogen"
	"github.com/stretchr/testify/require"
)

func TestResolveModulePath(t *testing.T) {
	req := require.New(t)

	m := &ModuleSpec{}
	err := m.Init()
	req.NoError(err)

	p, err := m.ResolveModulePath("github.com/kelveny/gdbc/cmd")
	req.NoError(err)
	req.True(strings.HasPrefix(p, m.RootDir))
	req.Equal("/cmd", strings.TrimPrefix(p, m.RootDir))

	fi, err := os.Stat(p)
	req.NoError(err)
	req.True(fi.IsDir())
}

func TestFlattenFieldSpecs(t *testing.T) {
	req := require.New(t)

	entities := []string{
		"Executive",
		"Executive2",
		"Executive3",
		"Executive4",
		"Executive5",
		"Executive6",
		"Executive7",
		"Executive8",
	}

	for _, entity := range entities {
		m := &ModuleSpec{}
		err := m.Init()
		req.NoError(err, entity)

		d, err := m.ResolveModulePath("github.com/kelveny/gdbc/test/crosspkg")
		req.NoError(err, entity)

		err = scanDir(m, d, scanPredicate, buildEntityRegistry)
		req.NoError(err, entity)

		e := lookupEntitySpec(d, entity)
		req.NotNil(e, entity)

		fieldSpecs := map[string][]EntityFieldSpec{}
		importSpecs := map[string][]gogen.ImportSpec{}
		tables := e.FlattenFieldSpecs(m, d, "executive", fieldSpecs, importSpecs)
		req.Equal([]string{
			"executive", "manager", "employee", "person",
		}, tables)

		req.Equal(4, len(fieldSpecs), entity)
		req.Equal(1, len(fieldSpecs["executive"]), entity)
		req.Equal(1, len(fieldSpecs["manager"]), entity)
		req.Equal(1, len(fieldSpecs["employee"]), entity)
		req.Equal(7, len(fieldSpecs["person"]), entity)

		req.Equal(4, len(importSpecs), entity)
		req.Equal([]gogen.ImportSpec{
			{
				Name: "",
				Path: "github.com/kelveny/gdbc/test/embed",
			},
		}, importSpecs["executive"], entity)
		req.Equal(0, len(importSpecs["manager"]))
		req.Equal(0, len(importSpecs["employee"]))
		req.Equal([]gogen.ImportSpec{
			{
				Name: "",
				Path: "time",
			},
		}, importSpecs["person"], entity)

		imports := mergeBaseImports(e.Imports, tables, importSpecs)
		req.Equal(2, len(imports))
		req.Equal([]gogen.ImportSpec{
			{
				Name: "",
				Path: "github.com/kelveny/gdbc/test/embed",
			},
			{
				Name: "",
				Path: "time",
			},
		}, imports, entity)
	}
}
