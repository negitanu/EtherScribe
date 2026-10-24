package fortigate

import (
	"strings"
	"unicode"
)

// Node is a nested FortiGate config tree (dict-like).
type Node map[string]any

func (n Node) Map(key string) Node {
	v, ok := n[key]
	if !ok || v == nil {
		return nil
	}
	m, ok := v.(Node)
	if !ok {
		return nil
	}
	return m
}

func (n Node) String(key string) string {
	v, ok := n[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []string:
		if len(t) == 1 {
			return t[0]
		}
		return strings.Join(t, " ")
	default:
		return ""
	}
}

func (n Node) List(key string) []string {
	v, ok := n[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return t
	default:
		return nil
	}
}

// ParseLines builds a config tree from logical lines (WallScribe _parse_config_tree).
func ParseLines(lines []string) (Node, []string) {
	var errors []string
	root := Node{}
	_, errs := parseTree(lines, 0, "", root, &errors)
	_ = errs
	return root, errors
}

func parseTree(lines []string, i int, terminator string, result Node, errors *[]string) (int, Node) {
	for i < len(lines) {
		line := lines[i]
		i++
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		cmd := parts[0]
		rest := ""
		if len(parts) > 1 {
			// Preserve original remainder after first token
			idx := strings.IndexFunc(line, unicode.IsSpace)
			if idx >= 0 {
				rest = strings.TrimSpace(line[idx:])
			}
		}

		switch cmd {
		case "config":
			name := rest
			child := result.Map(name)
			if child == nil {
				child = Node{}
				result[name] = child
			}
			var sub Node
			i, sub = parseTree(lines, i, "end", child, errors)
			result[name] = sub
		case "edit":
			name := unquote(firstToken(rest))
			child := Node{}
			var sub Node
			i, sub = parseTree(lines, i, "next", child, errors)
			sub["_name"] = name
			result[name] = sub
		case "set":
			key, val := splitKeyValue(rest)
			if key != "" {
				result[key] = parseValue(val)
			}
		case "unset":
			key := firstToken(rest)
			delete(result, key)
		case "append", "unselect":
			key, val := splitKeyValue(rest)
			if key == "" {
				*errors = append(*errors, "FortiGateリスト操作に値がありません")
				continue
			}
			existing := toStringSlice(result[key])
			values := toStringSlice(parseValue(val))
			if cmd == "append" {
				result[key] = uniqueAppend(existing, values)
			} else {
				result[key] = filterOut(existing, values)
			}
		case "next", "end":
			if cmd != terminator {
				*errors = append(*errors, "FortiGate設定のブロック終端が不正です: "+cmd)
			}
			return i, result
		default:
			*errors = append(*errors, "FortiGate設定の未対応コマンド")
		}
	}
	if terminator != "" {
		*errors = append(*errors, "FortiGate設定のブロック終端がありません: "+terminator)
	}
	return i, result
}

func splitKeyValue(s string) (string, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", ""
	}
	parts := strings.SplitN(s, " ", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.TrimSpace(parts[1])
}

func firstToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return strings.Fields(s)[0]
}

func parseValue(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	tokens := shellSplit(value)
	if len(tokens) == 0 {
		return ""
	}
	if len(tokens) == 1 {
		return tokens[0]
	}
	return tokens
}

// shellSplit is a minimal shlex-like splitter for FortiGate values.
func shellSplit(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if cur.Len() > 0 || quote != 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case ' ', '\t':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	_ = flush
	return out
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func toStringSlice(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return t
	default:
		return nil
	}
}

func uniqueAppend(a, b []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, x := range append(a, b...) {
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

func filterOut(a, b []string) []string {
	drop := map[string]struct{}{}
	for _, x := range b {
		drop[x] = struct{}{}
	}
	var out []string
	for _, x := range a {
		if _, ok := drop[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// SeparateLogicalLines ports quote-aware line joining from _separate_config.
func SeparateLogicalLines(content string) ([]string, []string) {
	var errors []string
	var lines []string
	var pending []string
	var quote rune
	escaped := false

	for _, raw := range strings.Split(content, "\n") {
		stripped := strings.TrimSpace(raw)
		if len(pending) == 0 && strings.HasPrefix(stripped, "#") {
			lines = append(lines, stripped)
			continue
		}
		fragment := stripped
		if len(pending) > 0 {
			fragment = raw
		}
		scan := fragment
		if len(pending) > 0 {
			scan = "\n" + fragment
		}
		pending = append(pending, fragment)
		for _, char := range scan {
			if escaped {
				escaped = false
				continue
			}
			if quote == '\'' {
				if char == '\'' {
					quote = 0
				}
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if quote != 0 {
				if char == quote {
					quote = 0
				}
				continue
			}
			if char == '\'' || char == '"' {
				quote = char
			}
		}
		if quote != 0 || escaped {
			continue
		}
		complete := strings.TrimSpace(strings.Join(pending, "\n"))
		if complete != "" {
			lines = append(lines, complete)
		}
		pending = nil
	}
	if len(pending) > 0 {
		errors = append(errors, "FortiGate設定に閉じていない引用符があります")
	}
	return lines, errors
}
