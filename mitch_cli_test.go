package mitch_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinary(t *testing.T) {
	t.Parallel()
	cli := buildCLI(t)

	t.Run("migrate", func(t *testing.T) {
		t.Parallel()
		total := countSQLFiles(t, "/app/testdata/migrations")

		expectOut := "mitch: successfully migrated database to version: " + strconv.Itoa(total)
		out, err := cli.run("--config", "/app/testdata/mitch.env")
		require.NoError(t, err)
		require.Contains(t, out, expectOut)
	})
	t.Run("rollback", func(tt *testing.T) {
		tt.Skip()
	})
}

type mitchBinary struct {
	binaryPath string
}

func (g mitchBinary) run(params ...string) (string, error) {
	cmd := exec.Command(g.binaryPath, params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run mitch: %v\nout: %v", err, string(out))
	}
	return string(out), nil
}

// buildCLI builds mitch test binary
func buildCLI(t *testing.T) mitchBinary {
	t.Helper()
	binName := "mitch-test"
	dir := t.TempDir()
	output := filepath.Join(dir, binName)
	args := []string{
		"build",
		"-o", output,
	}

	args = append(args, "./cmd")
	build := exec.Command("go", args...)
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build %s binary: %v: %s", binName, err, string(out))
	}
	return mitchBinary{
		binaryPath: output,
	}
}

func countSQLFiles(t *testing.T, dir string) int {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	require.NoError(t, err)
	return len(files)
}
