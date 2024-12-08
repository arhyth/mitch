package internal

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// ParseMigration
func ParseMigration(file fs.File) (*Version, error) {
	var forwardBuilder strings.Builder
	var rollbackBuilder strings.Builder
	inRollback := false

	// duplicate read for hashing SQL contents
	buf := new(bytes.Buffer)
	trdr := io.TeeReader(file, buf)
	scanner := bufio.NewScanner(trdr)
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

	hash := sha256.Sum256(buf.Bytes())
	return &Version{
		ContentHash: fmt.Sprintf("%x", hash[:]),
		Up: &SQL{
			Statements: strings.TrimSpace(forwardBuilder.String()),
		},
		Down: &SQL{
			Statements: strings.TrimSpace(rollbackBuilder.String()),
		},
	}, nil
}
