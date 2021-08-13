package statement

import (
	"fmt"
	"strings"
)

type part struct {
	query  string
	values []interface{}
}

func (p *part) String() (q string, err error) {
	var buf strings.Builder
	if err = p.Build(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (p *part) Build(buf Buffer) (err error) {
	return p.build(buf, false)
}

func (p *part) build(buf Buffer, keyword bool) (err error) {
	if strings.Count(p.query, "?") != len(p.values) {
		return fmt.Errorf("%w: %s, %#v", ErrInvalidArgNumber, p.query, p.values)
	}

	valueIdx := 0
	query := p.query
	for {
		idx := strings.Index(query, "?")
		if idx == -1 {
			buf.WriteString(query)
			break
		}

		buf.WriteString(query[:idx])
		query = query[idx+1:]

		arg := p.values[valueIdx]
		valueIdx++

		switch arg := arg.(type) {
		case Statement:
			err = arg.Build(buf)
		default:
			err = writeValue(buf, arg, keyword)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
