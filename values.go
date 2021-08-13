package statement

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func writeValue(buf Buffer, arg interface{}, keyword bool) (err error) {
	switch arg := arg.(type) {
	case nil:
		buf.WriteString("null")
	case int:
		buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int8:
		buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int16:
		buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int32:
		buf.WriteString(strconv.FormatInt(int64(arg), 10))
	case int64:
		buf.WriteString(strconv.FormatInt(arg, 10))
	case float32:
		buf.WriteString(strconv.FormatFloat(float64(arg), 'f', -1, 64))
	case float64:
		buf.WriteString(strconv.FormatFloat(arg, 'f', -1, 64))
	case bool:
		buf.WriteString(strconv.FormatBool(arg))
	case []byte:
		buf.WriteString(quoteBytes(arg))
	case string:
		if keyword {
			buf.WriteString(arg)
		} else {
			buf.WriteString(quoteString(arg))
		}
	case time.Time:
		buf.WriteString(arg.Truncate(time.Microsecond).Format("'2006-01-02 15:04:05.999999999Z07:00:00'"))
	case fmt.Stringer:
		buf.WriteString(quoteString(arg.String()))
	default:
		return fmt.Errorf("invalid arg type: %T, value: %#v", arg, arg)
	}

	return nil
}

// TODO: consider manually inlining this
func quoteString(str string) string {
	return "'" + strings.ReplaceAll(str, "'", "''") + "'"
}

// TODO: consider manually inlining this
func quoteBytes(buf []byte) string {
	return `'\x` + hex.EncodeToString(buf) + "'"
}
