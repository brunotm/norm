package migrate

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidNoTx = fmt.Errorf("migrate: migrations that disable transactions must have only one statement")
	noTXRegexp     = regexp.MustCompile(`--\s+migrate:\s+NoTransaction`)
)

func parseStatement(data []byte) (s Statements, err error) {
	s = Statements{}

	var stmt string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "--") {
			if noTXRegexp.MatchString(line) {
				s.NoTx = true
			}
			continue
		}

		if stmt != "" {
			stmt += " "
		}

		if line[len(line)-1] == ';' {
			stmt += line[:len(line)-1]
			s.Statements = append(s.Statements, stmt)
			stmt = ""
			continue
		}

		stmt += line
	}

	if stmt != "" {
		s.Statements = append(s.Statements, stmt)
	}

	if s.NoTx && len(s.Statements) > 1 {
		return s, ErrInvalidNoTx
	}

	return s, nil
}
