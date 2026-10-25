package fortigate

import (
	"sort"
	"strings"
	"unicode"
)

// Node is a nested FortiGate config tree (dict-like).
type Node map[string]any

const editOrderKey = "\x00edit-order"

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
	_, errs := parseTree(lines, 0, "", root, false, &errors)
	_ = errs
	return root, errors
}

func parseTree(lines []string, i int, terminator string, result Node, implicitVDOMEnd bool, errors *[]string) (int, Node) {
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
			i, sub = parseTree(lines, i, "end", child, name == "vdom", errors)
			result[name] = sub
		case "edit":
			name := firstToken(rest)
			if result.Map(name) == nil {
				result[editOrderKey] = append(toStringSlice(result[editOrderKey]), name)
			}
			child := result.Map(name)
			if child == nil {
				child = Node{}
			}
			var sub Node
			i, sub = parseTree(lines, i, "next", child, implicitVDOMEnd, errors)
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
			if cmd == "end" && terminator == "next" && implicitVDOMEnd {
				return i - 1, result
			}
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
	i := strings.IndexFunc(s, unicode.IsSpace)
	if i < 0 {
		return s, ""
	}
	return s[:i], strings.TrimSpace(s[i:])
}

func firstToken(s string) string {
	tokens := shellSplit(s)
	if len(tokens) == 0 {
		return ""
	}
	return tokens[0]
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
	escaped, started := false, false
	flush := func() {
		if started {
			out = append(out, cur.String())
			cur.Reset()
			started = false
		}
	}
	for _, r := range s {
		if escaped {
			if r != '\n' {
				cur.WriteRune(r)
			}
			escaped = false
			started = true
			continue
		}
		if r == '\\' && quote != '\'' {
			escaped = true
			started = true
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
		if r == '\'' || r == '"' {
			quote = r
			started = true
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			continue
		}
		cur.WriteRune(r)
		started = true
	}
	if escaped {
		cur.WriteRune('\\')
	}
	flush()
	return out
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

// EditNames retains device evaluation order. Map-only nodes have a deterministic fallback.
func EditNames(n Node) []string {
	var names []string
	seen := map[string]bool{}
	for _, name := range toStringSlice(n[editOrderKey]) {
		if n.Map(name) != nil && !seen[name] {
			names = append(names, name)
			seen[name] = true
		}
	}
	var rest []string
	for name, raw := range n {
		if _, ok := raw.(Node); ok && !seen[name] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	return append(names, rest...)
}
