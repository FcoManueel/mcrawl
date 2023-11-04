package persist

import (
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersistence(t *testing.T) {
	file := path.Join(t.TempDir(), "filedb")
	t.Log("file", file)
	db := Open(file)

	require.NotNil(t, db.f)
	require.True(t, strings.HasSuffix(db.Filename(), "filedb.json"))

	u1, err := url.Parse("http://test.com")
	require.NoError(t, err)

	u2, err := url.Parse("http://test.com/2")
	require.NoError(t, err)

	u3, err := url.Parse("http://test.com/3")
	require.NoError(t, err)

	// save two results
	err = db.Save(u1, []*url.URL{u2})
	require.NoError(t, err)

	err = db.Save(u2, []*url.URL{u3, u1})
	require.NoError(t, err)

	// load to check correctness
	results, err := db.load()
	require.NoError(t, err)
	require.Len(t, results, 2)

	require.Len(t, results[0].Children, 1)
	require.Equal(t, results[0].Page, u1.String())
	require.Equal(t, results[0].Children[0], u2.String())

	require.Len(t, results[1].Children, 2)
	require.Equal(t, results[1].Page, u2.String())
	require.Equal(t, results[1].Children[0], u3.String())
	require.Equal(t, results[1].Children[1], u1.String())

	// test close
	db.Close()
	require.Nil(t, db.f)
}
