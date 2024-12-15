package internal

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/arhyth/mitch"
	"github.com/rs/zerolog/log"
)

var (
	rgxVerPrefix = regexp.MustCompile(`^[0-9]+`)
)

func ParseMigration(file io.Reader) (*Version, error) {
	var forwardStmts, rollbackStmts []string
	inRollback := false

	buf := new(bytes.Buffer)
	tee := io.TeeReader(file, buf)
	scanner := bufio.NewScanner(tee)
	scanner.Split(SplitSQLStatements)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.Contains(line, "/* rollback") {
			inRollback = true
			spl := strings.Split(line, "/* rollback\n")
			rollbackStmts = append(rollbackStmts, spl[1])
			continue
		}
		if inRollback && strings.HasSuffix(line, "*/") {
			continue
		}

		if inRollback {
			rollbackStmts = append(rollbackStmts, line)
			continue
		}

		forwardStmts = append(forwardStmts, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	ver := &Version{
		ContentHash: HashSum(buf.String()),
		Up:          &SQL{Statements: forwardStmts},
		Down:        &SQL{Statements: rollbackStmts},
	}

	return ver, nil
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

var _ bufio.SplitFunc = SplitSQLStatements

// SplitSQLStatements implements bufio.SplitFunc
// It allows statements to span multiple lines but expects each statement
// to start on a new line.
// It also allows a statement to end with an inline comment.
func SplitSQLStatements(data []byte, atEOF bool) (int, []byte, error) {
	// 1st case: end of file, return last scanned
	if atEOF {
		return len(data), data, bufio.ErrFinalToken
	}

	i := bytes.IndexByte(data, ';')
	inl := -1
	if i != -1 {
		inl = bytes.IndexByte(data[i:], '\n')
	}
	switch {
	// 2nd case: no terminating `;`, reread with more data
	case i == -1:
		return 0, nil, nil
	// 3rd case: happy path, semicolon followed by a newline
	case i > -1 && inl > -1:
		if ith := bytes.IndexByte(data[i+1:i+inl], ';'); ith != -1 {
			return 0, nil, mitch.ErrMultiStatementLine
		}
		return i + inl + 1, data[:i+inl], nil
	// 4th case: semicolon without a new line
	case i > -1 && inl == -1:
		return 0, nil, nil
	default:
		log.Warn().
			Int("semicolon_idx", i).
			Int("newlind_idx", inl).
			Str("buffer", string(data)).
			Msg("unhandled SplitSQLStatements case")
		return 0, nil, nil
	}
}
