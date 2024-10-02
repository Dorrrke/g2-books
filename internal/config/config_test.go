package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	type want struct {
		cfg Config
	}
	type test struct {
		name  string
		flags []string
		env   func()
		want  want
	}
	tests := []test{
		{
			name:  "Test ReadConfig() func; Case 1:",
			flags: []string{"test", "-host", "124.123.1.11:8080", "-debug"},
			env:   nil,
			want: want{
				cfg: Config{
					Host:        "124.123.1.11:8080",
					DBDsn:       defaultDBDSN,
					MigratePath: defaultMigratePath,
					Debug:       true,
				},
			},
		},
		{
			name:  "Test ReadConfig() func; Case 2:",
			flags: []string{"test", "-debug"},
			env: func() {
				t.Setenv("SERVER_HOS", "1.1.1.1:1111")
				t.Setenv("DB_DSN", "testDsn")
				t.Setenv("MIGRATE_PATH", "testMigratePath")
			},
			want: want{
				cfg: Config{
					Host:        "1.1.1.1:1111",
					DBDsn:       "testDsn",
					MigratePath: "testMigratePath",
					Debug:       true,
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.flags) != 0 {
				os.Args = tc.flags
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			}
			if tc.env != nil {
				tc.env()
				defer os.Unsetenv("SERVER_HOS")
				defer os.Unsetenv("DB_DSN")
				defer os.Unsetenv("MIGRATE_PATH")
			}
			cfg := ReadConfig()
			assert.Equal(t, tc.want.cfg, cfg)
		})
	}
}
