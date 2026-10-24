// Package redaction ports WallScribe/utils/redaction.py.
package redaction

import "regexp"

const Redacted = "<redacted>"

var secretFields = map[string]struct{}{
	"password":    {},
	"community":   {},
	"secret":      {},
	"private_key": {},
	"psk_value":   {},
}

// ENC pattern from WallScribe: ENC followed by a non-space token.
var encPattern = regexp.MustCompile(`ENC\s+\S+`)

func IsSecretField(name string) bool {
	if _, ok := secretFields[name]; ok {
		return true
	}
	n := len(name)
	if n >= 9 && name[n-9:] == "_password" {
		return true
	}
	if n >= 7 && name[n-7:] == "_secret" {
		return true
	}
	return false
}

func ScrubSecrets(value string) string {
	return encPattern.ReplaceAllString(value, Redacted)
}

func SecretPresence(value string) string {
	if value == "" {
		return ""
	}
	return "設定あり"
}
