package migrate

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	noTXRegexp = regexp.MustCompile(`--\s+migrate:\s+NoTransaction`)
)

func parseStatement(data []byte) (s Statements, err error) {
	s = Statements{}

	var stmt string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "--") {
			if noTXRegexp.MatchString(line) {
				s.NoTx = true
			}
			continue
		}

		if line[len(line)-1] == ';' {
			if stmt != "" {
				stmt += " "
			}
			stmt += line[:len(line)-1]
			s.Statements = append(s.Statements, stmt)
			stmt = ""
		}
	}

	if s.NoTx && len(s.Statements) > 1 {
		return s, fmt.Errorf("migrate: migrations that disable transactions must have only one statement")
	}

	return s, nil
}
