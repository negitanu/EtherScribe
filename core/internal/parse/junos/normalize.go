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
	var chunk strings.Builder
	var quote byte
	escaped, comment, lineComment := false, false, false
	flatten := func(parts []string) []string {
		var path []string
		for _, frame := range stack {
			path = append(path, frame...)
		}
		return append(path, parts...)
	}
	emit := func(delimiter byte) {
		parts := shellSplit(strings.TrimSpace(chunk.String()))
		chunk.Reset()
		off := len(parts) > 0 && parts[0] == "inactive:"
		if off {
			parts = parts[1:]
		}
		path := flatten(parts)
		if len(parts) > 0 {
			out = append(out, Statement{Op: OpSet, Path: path})
			if off {
				out = append(out, Statement{Op: OpDeactivate, Path: append([]string{}, path...)})
			}
		}
		switch delimiter {
		case '{':
			if len(parts) == 0 {
				errs = append(errs, "brace mismatch: missing hierarchy")
			}
			stack = append(stack, parts)
		case '}':
			if len(parts) > 0 {
				errs = append(errs, "brace syntax: missing semicolon")
			}
			if len(stack) == 0 {
				errs = append(errs, "brace mismatch: extra }")
			} else {
				stack = stack[:len(stack)-1]
			}
		}
	}
	for i := 0; i < len(content); i++ {
		c := content[i]
		if lineComment {
			if c == '\n' {
				lineComment = false
				chunk.WriteByte(' ')
			}
			continue
		}
		if comment {
			if c == '*' && i+1 < len(content) && content[i+1] == '/' {
				comment = false
				i++
				chunk.WriteByte(' ')
			}
			continue
		}
		if quote != 0 {
			chunk.WriteByte(c)
			if escaped {
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == quote {
				quote = 0
			}
			continue
		}
		if c == '#' {
			lineComment = true
			continue
		}
		if c == '/' && i+1 < len(content) && content[i+1] == '*' {
			comment = true
			i++
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
			chunk.WriteByte(c)
		case '{', '}', ';':
			emit(c)
		default:
			chunk.WriteByte(c)
		}
	}
	if strings.TrimSpace(chunk.String()) != "" {
		errs = append(errs, "brace syntax: unterminated statement")
	}
	if quote != 0 {
		errs = append(errs, "brace syntax: unclosed quote")
	}
	if comment {
		errs = append(errs, "brace syntax: unclosed comment")
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
	escaped := false
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if quote != 0 {
			if escaped {
				cur.WriteRune(r)
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
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
