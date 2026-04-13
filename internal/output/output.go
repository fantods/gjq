package output

import (
	"encoding/json"
	"fmt"
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
	_, _ = fmt.Fprintf(w, "")
	result := func() error {
		if showPath && len(path) > 0 {
			parts := make([]string, len(path))
			for i, pt := range path {
				parts[i] = pt.String()
			}
			pathStr := strings.Join(parts, ".")
			if colorize {
				fmt.Fprintf(w, "%s%s%s:\n", ansiBoldMagenta, pathStr, ansiReset)
			} else {
				fmt.Fprintf(w, "%s:\n", pathStr)
			}
		}
		if err := writeColoredJSON(w, value, 0, pretty, colorize); err != nil {
			return err
		}
		fmt.Fprintln(w)
		return nil
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
		writeColor(w, "null", ansiDimRed, colorize)
	case bool:
		writeColor(w, fmt.Sprintf("%t", v), ansiBoldYellow, colorize)
	case int:
		writeColor(w, fmt.Sprintf("%d", v), ansiYellow, colorize)
	case float64:
		writeColor(w, formatFloat(v), ansiYellow, colorize)
	case string:
		quoted, _ := json.Marshal(v)
		writeColor(w, string(quoted), ansiGreen, colorize)
	case []interface{}:
		if _, err := fmt.Fprint(w, "["); err != nil {
			return err
		}
		for i, elem := range v {
			if pretty {
				if _, err := fmt.Fprintf(w, "\n%*s", nextIndent, ""); err != nil {
					return err
				}
			}
			if err := writeColoredJSON(w, elem, nextIndent, pretty, colorize); err != nil {
				return err
			}
			if i < len(v)-1 {
				if _, err := fmt.Fprint(w, ","); err != nil {
					return err
				}
			}
		}
		if pretty && len(v) > 0 {
			if _, err := fmt.Fprintf(w, "\n%*s", indent, ""); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, "]"); err != nil {
			return err
		}
	case map[string]interface{}:
		if _, err := fmt.Fprint(w, "{"); err != nil {
			return err
		}
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, key := range keys {
			if pretty {
				if _, err := fmt.Fprintf(w, "\n%*s", nextIndent, ""); err != nil {
					return err
				}
			}
			quotedKey, _ := json.Marshal(key)
			writeColor(w, string(quotedKey), ansiCyan, colorize)
			if pretty {
				if _, err := fmt.Fprint(w, ": "); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprint(w, ":"); err != nil {
					return err
				}
			}
			if err := writeColoredJSON(w, v[key], nextIndent, pretty, colorize); err != nil {
				return err
			}
			if i < len(keys)-1 {
				if _, err := fmt.Fprint(w, ","); err != nil {
					return err
				}
			}
		}
		if pretty && len(keys) > 0 {
			if _, err := fmt.Fprintf(w, "\n%*s", indent, ""); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(w, "}"); err != nil {
			return err
		}
	}
	return nil
}

func writeColor(w io.Writer, text, code string, colorize bool) {
	if colorize {
		fmt.Fprintf(w, "%s%s%s", code, text, ansiReset)
	} else {
		fmt.Fprint(w, text)
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
