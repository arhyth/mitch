package mitch_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

func TestBinaryForward(t *testing.T) {
	cli := buildCLI(t)

	t.Run("ok", func(tt *testing.T) {
		reqrd := require.New(tt)
		dbname := randomName()
		_, err := cli.run(
			"testhelper",
			"--db-name",
			dbname,
			"--tempdir",
			cli.envPath,
			"create",
		)
		reqrd.NoError(err)

		expectOut := "mitch: successfully migrated database to version: 9"
		out, err := cli.run("--env", cli.envPath)
		reqrd.NoError(err)
		reqrd.Contains(out, expectOut)
		tt.Cleanup(func() {
			cli.run(
				"testhelper",
				"--db-name",
				dbname,
				"drop",
			)
		})
	})
}

func TestBinaryRollback(t *testing.T) {
	t.Run("ok", func(tt *testing.T) {
		cli := buildCLI(tt)
		reqrd := require.New(tt)

		// setup and migrate forward
		dbname := randomName()
		_, err := cli.run(
			"testhelper",
			"--db-name",
			dbname,
			"--tempdir",
			cli.envPath,
			"create",
		)
		reqrd.NoError(err)
		_, err = cli.run("--env", cli.envPath)
		reqrd.NoError(err)

		// rollback
		out, err := cli.run(
			"--env",
			cli.envPath,
			"--rollback",
			"002_add_new_field_norollback.sql",
		)
		expectOut := "mitch: successfully rolled database back to version: 1"
		reqrd.NoError(err)
		reqrd.Contains(out, expectOut)
	})
}

type testenv struct {
	binaryPath string
	envPath    string
}

func (e testenv) run(params ...string) (string, error) {
	cmd := exec.Command(e.binaryPath, params...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run mitch: %v\nout: %v", err, string(out))
	}
	return string(out), nil
}

func buildCLI(t *testing.T) testenv {
	t.Helper()
	binName := "mitch"
	dir := t.TempDir()
	binOut := filepath.Join(dir, binName)
	args := []string{
		"build",
		"-o", binOut,
		"./cmd",
	}

	build := exec.Command("go", args...)
	out, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build %s test binary: %v: %s", binName, err, string(out))
	}

	return testenv{
		binaryPath: binOut,
		envPath:    filepath.Join(dir, "test.env"),
	}
}

func randomName() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(uint64(time.Now().UnixNano()))
	result := make([]byte, 8)

	for i := 0; i < 8; i++ {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
