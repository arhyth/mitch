package internal

import (
	"bufio"
	"io/fs"
	"strings"
)

// ParseMigration
func ParseMigration(file fs.File) (*Version, error) {
	var forwardBuilder strings.Builder
	var rollbackBuilder strings.Builder
	inRollback := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "/* rollback") {
			inRollback = true
			continue
		}
		if inRollback && strings.HasSuffix(trimmed, "*/") {
			inRollback = false
			continue
		}

		if inRollback {
			rollbackBuilder.WriteString(line + "\n")
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "--") && trimmed != "*/" {
			forwardBuilder.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Version{
		Up: &Migration{
			SQL: strings.TrimSpace(forwardBuilder.String()),
		},
		Down: &Migration{
			SQL: strings.TrimSpace(rollbackBuilder.String()),
		},
	}, nil
}
