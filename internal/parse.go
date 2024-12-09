package internal

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"regexp"
	"strconv"
	"strings"

	"github.com/arhyth/mitch"
)

var (
	rgxVerPrefix = regexp.MustCompile(`^[0-9]+`)
)

func ParseMigration(file fs.File) (*Version, error) {
	var forwardBuilder strings.Builder
	var rollbackBuilder strings.Builder
	inRollback := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "/* rollback") {
			inRollback = true
			continue
		}
		if inRollback && strings.HasSuffix(line, "*/") {
			inRollback = false
			continue
		}

		if inRollback {
			rollbackBuilder.WriteString(line + "\n")
		} else if line != "" && !strings.HasPrefix(line, "--") && line != "*/" {
			forwardBuilder.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sum := HashSum(forwardBuilder.String(), rollbackBuilder.String())
	return &Version{
		ContentHash: sum,
		Up: &SQL{
			Statements: strings.TrimSpace(forwardBuilder.String()),
		},
		Down: &SQL{
			Statements: strings.TrimSpace(rollbackBuilder.String()),
		},
	}, nil
}

func HashSum(content ...string) string {
	hash := sha256.New()
	for _, c := range content {
		hash.Write([]byte(c))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func ParseVersion(fname string) (int64, error) {
	verst := rgxVerPrefix.FindString(fname)
	if verst == "" {
		return 0, mitch.ErrFileVersionPrefix
	}
	n, err := strconv.ParseInt(verst, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse version from migration file: %s: %w", fname, err)
	}
	if n < 1 {
		return 0, mitch.ErrVersionZero
	}
	return n, nil
}
