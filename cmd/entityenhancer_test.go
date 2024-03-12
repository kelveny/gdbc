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

	m := &ModuleSpec{}
	err := m.Init()
	req.NoError(err)

	d, err := m.ResolveModulePath("github.com/kelveny/gdbc/test/crosspkg")
	req.NoError(err)

	err = scanDir(m, d, scanPredicate, buildEntityRegistry)
	req.NoError(err)

	e := lookupEntitySpec(d, "Executive")
	req.NotNil(e)

	fieldSpecs := map[string][]EntityFieldSpec{}
	importSpecs := map[string][]gogen.ImportSpec{}
	tables := e.FlattenFieldSpecs(m, d, "executive", fieldSpecs, importSpecs)
	req.Equal([]string{
		"executive", "manager", "employee", "person",
	}, tables)

	req.Equal(4, len(fieldSpecs))
	req.Equal(1, len(fieldSpecs["executive"]))
	req.Equal(1, len(fieldSpecs["manager"]))
	req.Equal(1, len(fieldSpecs["employee"]))
	req.Equal(7, len(fieldSpecs["person"]))

	req.Equal(4, len(importSpecs))
	req.Equal([]gogen.ImportSpec{
		{
			Name: "",
			Path: "github.com/kelveny/gdbc/test/embed",
		},
	}, importSpecs["executive"])
	req.Equal(0, len(importSpecs["manager"]))
	req.Equal(0, len(importSpecs["employee"]))
	req.Equal([]gogen.ImportSpec{
		{
			Name: "",
			Path: "time",
		},
	}, importSpecs["person"])

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
	}, imports)
}
