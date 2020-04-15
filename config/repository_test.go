package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseRepository tests the parse repository function.
func TestParseRepository(t *testing.T) {
	type testcase struct {
		name string
		s    string
		tf   func(t *testing.T, r *Repository, err error)
	}

	tests := []testcase{
		{
			"Valid",
			"host=172.16.1.1 port=5432 username=some password=pass driver_name=postgres protocol=postgresql sslmode=true dbname=some_db max_timeout=5m raw_url=some://raw.url",
			func(t *testing.T, r *Repository, err error) {
				require.NoError(t, err)

				assert.Equal(t, "172.16.1.1", r.Host)
				assert.Equal(t, 5432, r.Port)
				assert.Equal(t, "some", r.Username)
				assert.Equal(t, "pass", r.Password)
				assert.Equal(t, "postgres", r.DriverName)
				assert.Equal(t, "some_db", r.DBName)
				assert.Equal(t, time.Minute*5, *r.MaxTimeout)
				assert.Equal(t, "some://raw.url", r.RawURL)
			},
		},
		{
			"InvalidEqualSymbol",
			"host:172.16.1.1 port=5432",
			func(t *testing.T, r *Repository, err error) {
				require.Error(t, err)
			},
		},
		{
			"InvalidPort",
			"host=172.16.1.1 port=as5432",
			func(t *testing.T, r *Repository, err error) {
				require.Error(t, err)
			},
		},
		{
			"UnknownKey",
			"host=172.16.1.1 unknown=something",
			func(t *testing.T, r *Repository, err error) {
				require.NoError(t, err)
			},
		},
		{
			"InvalidTimeout",
			"host=172.16.1.1 max_timeout=3123dhas",
			func(t *testing.T, r *Repository, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tcase := range tests {
		t.Run(tcase.name, func(t *testing.T) {
			t.Helper()
			r := &Repository{}

			err := r.Parse(tcase.s)
			tcase.tf(t, r, err)
		})
	}
}
