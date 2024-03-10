package cmd

import (
	"os"
	"strings"
	"testing"

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
	tables := e.FlattenFieldSpecs(m, d, "executive", fieldSpecs)
	req.Equal([]string{
		"executive", "manager", "employee", "person",
	}, tables)

	req.Equal(4, len(fieldSpecs))
	req.Equal(1, len(fieldSpecs["executive"]))
	req.Equal(1, len(fieldSpecs["manager"]))
	req.Equal(1, len(fieldSpecs["employee"]))
	req.Equal(7, len(fieldSpecs["person"]))
}
