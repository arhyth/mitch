package internal

import (
	"bufio"
	"os"
	"strings"
)

// ParseMigrationFile
func ParseMigrationFile(filePath string) (*Version, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "--") && !strings.HasPrefix(trimmed, "/*") {
			forwardBuilder.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Version{
		Up: &Migration{
			SQL: forwardBuilder.String(),
		},
		Down: &Migration{
			SQL: rollbackBuilder.String(),
		},
	}, nil
}
