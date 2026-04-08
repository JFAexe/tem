package template

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"

	"github.com/JFAexe/tem/pkg/env"
)

type Dict = map[string]any

func CommonFuncs() template.FuncMap {
	return template.FuncMap{
		"inline":             inline(template.New("")),
		"pwd":                os.Getwd,
		"hostname":           os.Hostname,
		"envs":               env.Environ,
		"env":                env.Get,
		"envOr":              swapArgs(env.Or),
		"envExpand":          env.Expand,
		"envEscape":          env.Escape,
		"envIsSet":           env.IsSet,
		"fromEnv":            env.DecodeString,
		"fileContent":        fileContent,
		"osIsAbs":            filepath.IsAbs,
		"osAbs":              filepath.Abs,
		"osClean":            filepath.Clean,
		"osBase":             filepath.Base,
		"osDir":              filepath.Dir,
		"osExt":              filepath.Ext,
		"osJoin":             filepath.Join,
		"osGlob":             filepath.Glob,
		"isAbs":              path.IsAbs,
		"clean":              path.Clean,
		"base":               path.Base,
		"dir":                path.Dir,
		"ext":                path.Ext,
		"Join":               path.Join,
		"equalFold":          strings.EqualFold,
		"upper":              strings.ToUpper,
		"lower":              strings.ToLower,
		"title":              strings.ToTitle,
		"trimSpace":          strings.TrimSpace,
		"trim":               swapArgs(strings.Trim),
		"trimLeft":           swapArgs(strings.TrimLeft),
		"trimRight":          swapArgs(strings.TrimRight),
		"trimPrefix":         swapArgs(strings.TrimPrefix),
		"trimSuffix":         swapArgs(strings.TrimSuffix),
		"split":              swapArgs(strings.Split),
		"replace":            replace,
		"join":               join,
		"joinList":           swapArgs(strings.Join),
		"truncate":           truncate,
		"repeat":             swapArgs(strings.Repeat),
		"contains":           swapArgs(strings.Contains),
		"hasPrefix":          swapArgs(strings.HasPrefix),
		"hasSuffix":          swapArgs(strings.HasSuffix),
		"indent":             indent,
		"timeNow":            timeNow,
		"timeOffset":         timeOffset,
		"timeTruncate":       timeTruncate,
		"timeUTC":            timeUTC,
		"timeLocal":          timeLocal,
		"timeFormat":         timeFormat,
		"timeFormatTime":     timeFormatTime,
		"timeFormatDate":     timeFormatDate,
		"timeFormatDateTime": timeFormatDateTime,
		"toString":           toString,
		"toMD5":              toMD5,
		"toSHA1":             toSHA1,
		"toSHA3":             toSHA3,
		"toSHA256":           toSHA256,
		"toSHA512":           toSHA512,
		"toJSON":             toJSON,
		"toJSONPretty":       toJSONPretty,
		"fromJSON":           fromJSON,
		"toYAML":             toYAML,
		"fromYAML":           fromYAML,
		"toTOML":             toTOML,
		"fromTOML":           fromTOML,
		"toBase64":           toBase64,
		"fromBase64":         fromBase64,
		"toHex":              toHex,
		"fromHex":            fromHex,
		"xor":                xor,
		"dict":               dict,
		"set":                set,
		"unset":              unset,
		"isSet":              isSet,
		"merge":              merge,
	}
}

func EnvFuncs(envs env.Store) template.FuncMap {
	return template.FuncMap{
		"envs":      envs.Environ,
		"env":       envs.Get,
		"envOr":     swapArgs(envs.Or),
		"envExpand": envs.Expand,
		"envIsSet":  envs.IsSet,
	}
}

func TemplateFuncs(t *template.Template) template.FuncMap {
	return template.FuncMap{
		"inline": inline(t),
	}
}

func inline(t *template.Template) func(value string, data ...any) (string, error) {
	return func(value string, data ...any) (string, error) {
		var (
			buf bytes.Buffer
			ctx = any(data)
		)

		if len(data) == 1 {
			ctx = data[0]
		}

		clone, err := t.Clone()
		if err != nil {
			return "", err
		}

		if clone, err = clone.New("inline").Parse(value); err != nil {
			return "", err
		}

		if err := clone.Execute(&buf, ctx); err != nil {
			return "", err
		}

		return buf.String(), nil
	}
}

func fileContent(filePath string) (string, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func replace(old, new, src string) string {
	return strings.ReplaceAll(src, old, new)
}

func join(separator string, items ...string) string {
	return strings.Join(items, separator)
}

func truncate(size int, str string) string {
	runes := []rune(str)

	if size < 0 && len(runes)+size > 0 {
		return string(runes[len(runes)+size:])
	}

	if size >= 0 && len(runes) > size {
		return string(runes[:size])
	}

	return str
}

func indent(level int, str string) string {
	if level <= 0 || str == "" {
		return str
	}

	var (
		builder strings.Builder

		prefix = strings.Repeat(" ", level)
	)

	for i, s := range strings.Split(str, "\n") {
		if i > 0 {
			builder.WriteByte('\n')
		}

		if strings.TrimSpace(s) != "" {
			builder.WriteString(prefix)
			builder.WriteString(s)
		}
	}

	return builder.String()
}

func timeNow() time.Time {
	return time.Now()
}

func timeOffset(offset string, t time.Time) (time.Time, error) {
	dur, err := time.ParseDuration(offset)
	if err != nil {
		return time.Time{}, err
	}

	return t.Add(dur), nil
}

func timeTruncate(step string, t time.Time) (time.Time, error) {
	dur, err := time.ParseDuration(step)
	if err != nil {
		return time.Time{}, err
	}

	return t.Truncate(dur), nil
}

func timeUTC(t time.Time) time.Time {
	return t.UTC()
}

func timeLocal(t time.Time) time.Time {
	return t.Local()
}

func timeFormat(t time.Time) string {
	return t.Format(time.RFC3339)
}

func timeFormatTime(t time.Time) string {
	return t.Format(time.TimeOnly)
}

func timeFormatDate(t time.Time) string {
	return t.Format(time.DateOnly)
}

func timeFormatDateTime(t time.Time) string {
	return t.Format(time.DateTime)
}

func toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.GoStringer:
		return v.GoString()
	case fmt.Stringer:
		return v.String()
	}

	switch rv := reflect.ValueOf(value); rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return ""
		}

		return toString(rv.Elem().Interface())
	}

	return fmt.Sprint(value)
}

func toMD5(value string) string {
	hash := md5.Sum([]byte(value))

	return hex.EncodeToString(hash[:])
}

func toSHA1(value string) string {
	hash := sha1.Sum([]byte(value))

	return hex.EncodeToString(hash[:])
}

func toSHA3(value string) string {
	hash := sha3.Sum512([]byte(value))

	return hex.EncodeToString(hash[:])
}

func toSHA256(value string) string {
	hash := sha256.Sum256([]byte(value))

	return hex.EncodeToString(hash[:])
}

func toSHA512(value string) string {
	hash := sha512.Sum512([]byte(value))

	return hex.EncodeToString(hash[:])
}

func toJSON(value any) (string, error) {
	out, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode json: %w", err)
	}

	return string(out), nil
}

func toJSONPretty(value any) (string, error) {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to encode json: %w", err)
	}

	return string(out), nil
}

func fromJSON(value string) (map[string]any, error) {
	var out map[string]any

	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}

	return out, nil
}

func toYAML(value any) (string, error) {
	out, err := yaml.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode yaml: %w", err)
	}

	return string(out), nil
}

func fromYAML(value string) (map[string]any, error) {
	var out map[string]any

	if err := yaml.Unmarshal([]byte(value), &out); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %w", err)
	}

	return out, nil
}

func toTOML(value any) (string, error) {
	out, err := toml.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to encode toml: %w", err)
	}

	return string(out), nil
}

func fromTOML(value string) (map[string]any, error) {
	var out map[string]any

	if err := toml.Unmarshal([]byte(value), &out); err != nil {
		return nil, fmt.Errorf("failed to decode toml: %w", err)
	}

	return out, nil
}

func toBase64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func fromBase64(value string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		if raw, err = base64.URLEncoding.DecodeString(value); err != nil {
			return "", err
		}
	}

	return string(raw), nil
}

func toHex(value string) string {
	return hex.EncodeToString([]byte(value))
}

func fromHex(value string) (string, error) {
	raw, err := hex.DecodeString(value)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func xor(key, value string) string {
	data := []byte(value)

	for i := range data {
		data[i] ^= key[i%len(key)]
	}

	return string(data)
}

func dict(kv ...any) (Dict, error) {
	d := make(Dict)

	if len(kv)%2 != 0 {
		return d, fmt.Errorf("amount of arguments for key-values should be even, got %d", len(kv))
	}

	for i := 0; i < len(kv); i += 2 {
		d[toString(kv[i])] = kv[i+1]
	}

	return d, nil
}

func set(key string, value any, d Dict) Dict {
	d[key] = value

	return d
}

func unset(key string, d Dict) Dict {
	delete(d, key)

	return d
}

func isSet(key string, d Dict) bool {
	_, ok := d[key]

	return ok
}

func merge(from, to Dict) Dict {
	maps.Copy(to, from)

	return to
}

func swapArgs[L, R, T any](fn func(L, R) T) func(R, L) T {
	return func(r R, l L) T {
		return fn(l, r)
	}
}
