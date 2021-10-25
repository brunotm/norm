package statement

import (
	"fmt"
	"strings"

	"github.com/brunotm/norm/internal/buffer"
)

// Part is a query fragment that satisfies the statement.Statement interface
type Part struct {
	Query  string
	Values []interface{}
}

// String builds the part and returns the resulting query.
func (p *Part) String() (q string, err error) {
	buf := buffer.New()
	defer buf.Release()

	if err = p.Build(buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Build builds the part into the given buffer.
func (p *Part) Build(buf Buffer) (err error) {
	return p.build(buf, false)
}

func (p *Part) build(buf Buffer, keyword bool) (err error) {
	if strings.Count(p.Query, "?") != len(p.Values) {
		return fmt.Errorf("%w: %s, %#v", ErrInvalidArgNumber, p.Query, p.Values)
	}

	valueIdx := 0
	query := p.Query
	for {
		idx := strings.Index(query, "?")
		if idx == -1 {
			_, _ = buf.WriteString(query)
			break
		}

		_, _ = buf.WriteString(query[:idx])
		query = query[idx+1:]

		arg := p.Values[valueIdx]
		valueIdx++

		switch arg := arg.(type) {
		case Statement:
			_, _ = buf.WriteString("(")
			err = arg.Build(buf)
			_, _ = buf.WriteString(")")
		default:
			err = writeValue(buf, arg, keyword)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
