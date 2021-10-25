package statement

import (
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brunotm/norm/internal/buffer"
)

var rfc3339micro = "'2006-01-02T15:04:05.999999Z07:00'"

func writeValue(buf Buffer, arg interface{}, keyword bool) (err error) {
	if v, ok := arg.(driver.Valuer); ok {
		if arg, err = v.Value(); err != nil {
			return err
		}
	}

	switch arg := arg.(type) {
	case nil:
		_, _ = buf.WriteString("null")
	case int:
		_, _ = buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int8:
		_, _ = buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int16:
		_, _ = buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int32:
		_, _ = buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int64:
		_, _ = buf.WriteString(strconv.FormatInt(arg, 10))
	case float32:
		_, _ = buf.WriteString(strconv.FormatFloat(float64(arg), 'f', -1, 64))
	case float64:
		_, _ = buf.WriteString(strconv.FormatFloat(arg, 'f', -1, 64))
	case bool:
		_, _ = buf.WriteString(strconv.FormatBool(arg))
	case []byte:
		_, _ = buf.WriteString(quoteBytes(arg))
	case string:
		if keyword {
			_, _ = buf.WriteString(arg)
		} else {
			_, _ = buf.WriteString(quoteString(arg))
		}
	case time.Time:
		_, _ = buf.WriteString(arg.Format(rfc3339micro))
	case fmt.Stringer:
		_, _ = buf.WriteString(quoteString(arg.String()))
	default:
		return fmt.Errorf("statement: invalid arg type: %T, value: %#v", arg, arg)
	}

	return nil
}

// TODO: consider manually inlining this
func quoteString(str string) string {
	buf := buffer.New()
	defer buf.Release()

	buf.WriteString(`'`)
	buf.WriteString(strings.ReplaceAll(str, "'", "''"))
	buf.WriteString(`'`)
	return buf.String()
}

// TODO: consider manually inlining this
func quoteBytes(b []byte) string {
	buf := buffer.New()
	defer buf.Release()

	buf.WriteString(`'\x`)
	buf.WriteString(hex.EncodeToString(b))
	buf.WriteString(`'`)
	return buf.String()
}
