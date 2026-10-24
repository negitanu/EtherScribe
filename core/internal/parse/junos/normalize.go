package junos

import (
	"strings"
	"unicode"
)

// Op is the statement operation.
type Op string

const (
	OpSet        Op = "set"
	OpDeactivate Op = "deactivate"
	OpDelete     Op = "delete"
)

// Statement is a flattened Junos configuration statement.
type Statement struct {
	Op    Op
	Path  []string
	Value string
	Raw   string
}

// Normalize detects brace vs set and returns a statement stream.
func Normalize(content string) ([]Statement, []string) {
	trimmed := strings.TrimSpace(content)
	if looksLikeSet(trimmed) {
		return ParseSet(content)
	}
	return ParseBrace(content)
}

func looksLikeSet(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "set ") || strings.HasPrefix(line, "deactivate ") || strings.HasPrefix(line, "delete ") {
			return true
		}
		if strings.Contains(line, "{") {
			return false
		}
	}
	return strings.Contains(content, "\nset ")
}

// ParseSet parses `display set` style lines.
func ParseSet(content string) ([]Statement, []string) {
	var out []Statement
	var errs []string
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		op := OpSet
		rest := line
		switch {
		case strings.HasPrefix(line, "set "):
			rest = strings.TrimSpace(line[4:])
		case strings.HasPrefix(line, "deactivate "):
			op = OpDeactivate
			rest = strings.TrimSpace(line[len("deactivate "):])
		case strings.HasPrefix(line, "delete "):
			op = OpDelete
			rest = strings.TrimSpace(line[len("delete "):])
		default:
			continue
		}
		tokens := shellSplit(rest)
		if len(tokens) == 0 {
			continue
		}
		out = append(out, Statement{Op: op, Path: tokens, Raw: line})
	}
	return out, errs
}

// ParseBrace flattens curly-brace Junos config into set-like path tokens.
// Each `{` pushes one frame (possibly multiple tokens, e.g. "unit 0"); each `}` pops one frame.
func ParseBrace(content string) ([]Statement, []string) {
	var out []Statement
	var errs []string
	var stack [][]string
	flat := func() []string {
		var p []string
		for _, fr := range stack {
			p = append(p, fr...)
		}
		return p
	}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/*") {
			continue
		}
		if i := strings.Index(line, "/*"); i >= 0 {
			line = strings.TrimSpace(line[:i])
			if line == "" {
				continue
			}
		}
		// Close braces (possibly multiple on one line)
		for strings.Contains(line, "}") {
			before, after, _ := strings.Cut(line, "}")
			before = strings.TrimSpace(before)
			if before != "" {
				// rare: `foo; }` style — emit before closing
				if strings.HasSuffix(before, ";") {
					before = strings.TrimSpace(strings.TrimSuffix(before, ";"))
				}
				if before != "" && !strings.HasSuffix(before, "{") {
					parts := shellSplit(before)
					path := append(flat(), parts...)
					out = append(out, Statement{Op: OpSet, Path: path, Raw: raw})
				}
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			} else {
				errs = append(errs, "brace mismatch: extra }")
			}
			line = strings.TrimSpace(after)
		}
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, "{") {
			name := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			parts := shellSplit(name)
			if len(parts) == 0 {
				continue
			}
			stack = append(stack, parts)
			continue
		}
		if strings.HasSuffix(line, ";") {
			line = strings.TrimSpace(strings.TrimSuffix(line, ";"))
		}
		parts := shellSplit(line)
		if len(parts) == 0 {
			continue
		}
		path := append(flat(), parts...)
		out = append(out, Statement{Op: OpSet, Path: path, Raw: raw})
	}
	if len(stack) > 0 {
		errs = append(errs, "brace mismatch: unclosed {")
	}
	return out, errs
}

func shellSplit(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		switch r {
		case '"', '\'':
			quote = r
		default:
			if unicode.IsSpace(r) {
				flush()
			} else {
				cur.WriteRune(r)
			}
		}
	}
	flush()
	return out
}

// PathHas reports whether stmt path starts with prefix tokens.
func PathHas(st Statement, prefix ...string) bool {
	if len(st.Path) < len(prefix) {
		return false
	}
	for i, p := range prefix {
		if st.Path[i] != p {
			return false
		}
	}
	return true
}

// PathValue returns tokens after prefix as a single string.
func PathValue(st Statement, prefixLen int) string {
	if len(st.Path) <= prefixLen {
		return ""
	}
	return strings.Join(st.Path[prefixLen:], " ")
}
