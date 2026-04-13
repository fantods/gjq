package output

import (
	"encoding/json"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/fantods/gjq/internal/query"
)

const (
	ansiReset       = "\x1b[0m"
	ansiDimRed      = "\x1b[2;31m"
	ansiBoldYellow  = "\x1b[1;33m"
	ansiYellow      = "\x1b[33m"
	ansiGreen       = "\x1b[0;32m"
	ansiCyan        = "\x1b[0;36m"
	ansiBoldMagenta = "\x1b[1;35m"
)

var indentBuf []byte

func init() {
	indentBuf = make([]byte, 256)
	for i := range indentBuf {
		indentBuf[i] = ' '
	}
}

func Depth(value interface{}) int {
	switch v := value.(type) {
	case nil, bool, int, float64, string:
		return 1
	case []interface{}:
		maxD := 0
		for _, elem := range v {
			if d := Depth(elem); d > maxD {
				maxD = d
			}
		}
		return 1 + maxD
	case map[string]interface{}:
		maxD := 0
		for _, val := range v {
			if d := Depth(val); d > maxD {
				maxD = d
			}
		}
		return 1 + maxD
	default:
		return 1
	}
}

func WriteResult(w io.Writer, value interface{}, path []query.PathType, pretty, showPath, colorize bool) error {
	var result error
	func() {
		if showPath && len(path) > 0 {
			parts := make([]string, len(path))
			for i, pt := range path {
				parts[i] = pt.String()
			}
			pathStr := strings.Join(parts, ".")
			if colorize {
				io.WriteString(w, ansiBoldMagenta)
				io.WriteString(w, pathStr)
				io.WriteString(w, ansiReset)
				io.WriteString(w, ":\n")
			} else {
				io.WriteString(w, pathStr)
				io.WriteString(w, ":\n")
			}
		}
		if err := writeColoredJSON(w, value, 0, pretty, colorize); err != nil {
			result = err
			return
		}
		io.WriteString(w, "\n")
	}()

	if result != nil {
		if isBrokenPipe(result) {
			return nil
		}
	}
	return result
}

func writeColoredJSON(w io.Writer, value interface{}, indent int, pretty, colorize bool) error {
	nextIndent := indent + 2

	switch v := value.(type) {
	case nil:
		writeColorString(w, "null", ansiDimRed, colorize)
	case bool:
		writeColorBytes(w, strconv.AppendBool(nil, v), ansiBoldYellow, colorize)
	case int:
		writeColorBytes(w, strconv.AppendInt(nil, int64(v), 10), ansiYellow, colorize)
	case float64:
		writeColorString(w, formatFloat(v), ansiYellow, colorize)
	case string:
		quoted, _ := json.Marshal(v)
		writeColorBytes(w, quoted, ansiGreen, colorize)
	case []interface{}:
		io.WriteString(w, "[")
		for i, elem := range v {
			if pretty {
				writeIndent(w, nextIndent)
			}
			if err := writeColoredJSON(w, elem, nextIndent, pretty, colorize); err != nil {
				return err
			}
			if i < len(v)-1 {
				io.WriteString(w, ",")
			}
		}
		if pretty && len(v) > 0 {
			writeIndent(w, indent)
		}
		io.WriteString(w, "]")
	case map[string]interface{}:
		io.WriteString(w, "{")
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, key := range keys {
			if pretty {
				writeIndent(w, nextIndent)
			}
			quotedKey, _ := json.Marshal(key)
			writeColorBytes(w, quotedKey, ansiCyan, colorize)
			if pretty {
				io.WriteString(w, ": ")
			} else {
				io.WriteString(w, ":")
			}
			if err := writeColoredJSON(w, v[key], nextIndent, pretty, colorize); err != nil {
				return err
			}
			if i < len(keys)-1 {
				io.WriteString(w, ",")
			}
		}
		if pretty && len(keys) > 0 {
			writeIndent(w, indent)
		}
		io.WriteString(w, "}")
	}
	return nil
}

func writeIndent(w io.Writer, n int) {
	io.WriteString(w, "\n")
	for n > 0 {
		chunk := n
		if chunk > len(indentBuf) {
			chunk = len(indentBuf)
		}
		w.Write(indentBuf[:chunk])
		n -= chunk
	}
}

func writeColorString(w io.Writer, text, code string, colorize bool) {
	if colorize {
		io.WriteString(w, code)
		io.WriteString(w, text)
		io.WriteString(w, ansiReset)
	} else {
		io.WriteString(w, text)
	}
}

func writeColorBytes(w io.Writer, text []byte, code string, colorize bool) {
	if colorize {
		io.WriteString(w, code)
		w.Write(text)
		io.WriteString(w, ansiReset)
	} else {
		w.Write(text)
	}
}

func formatFloat(f float64) string {
	if f == float64(int(f)) && !math.IsInf(f, 0) && !math.IsNaN(f) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func isBrokenPipe(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "broken pipe")
}
