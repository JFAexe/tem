package functions

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"maps"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"

	"github.com/JFAexe/tem/pkg/convert"
	"github.com/JFAexe/tem/pkg/env"
)

var hashers = map[string]hash.Hash{
	"crc32":      crc32.NewIEEE(),
	"crc64":      crc64.New(crc64.MakeTable(crc64.ECMA)),
	"crc64-iso":  crc64.New(crc64.MakeTable(crc64.ISO)),
	"md5":        md5.New(),
	"sha1":       sha1.New(),
	"sha256-224": sha256.New224(),
	"sha256":     sha256.New(),
	"sha3-224":   sha3.New224(),
	"sha3-256":   sha3.New256(),
	"sha3-384":   sha3.New384(),
	"sha3-512":   sha3.New512(),
	"sha512-224": sha512.New512_224(),
	"sha512-256": sha512.New512_256(),
	"sha512-384": sha512.New384(),
	"sha512":     sha512.New(),
}

type Data struct{}

func (*Data) Xor(key, value any) string {
	var (
		d = convert.ToByteSlice(value)
		k = convert.ToByteSlice(key)
	)

	for i := range d {
		d[i] ^= k[i%len(k)]
	}

	return string(d)
}

func (*Data) Hash(kind, value any) (string, error) {
	k := strings.ToLower(strings.TrimSpace(convert.ToString(kind)))

	hasher, ok := hashers[k]
	if !ok {
		return "", fmt.Errorf("invalid hash function %#q, supported: %s", k, strings.Join(slices.Sorted(maps.Keys(hashers)), ", "))
	}

	return hex.EncodeToString(hasher.Sum(convert.ToByteSlice(value))), nil
}

func (*Data) FromHex(data any) (string, error) {
	raw, err := hex.DecodeString(convert.ToString(data))
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (*Data) ToHex(value any) string {
	return hex.EncodeToString(convert.ToByteSlice(value))
}

func (*Data) FromBase32(data any) (string, error) {
	str := convert.ToString(data)

	raw, err := base32.StdEncoding.DecodeString(str)
	if err != nil {
		if raw, err = base32.HexEncoding.DecodeString(str); err != nil {
			return "", err
		}
	}

	return string(raw), nil
}

func (*Data) ToBase32(value any) string {
	return base32.StdEncoding.EncodeToString(convert.ToByteSlice(value))
}

func (*Data) ToBase32HEX(value any) string {
	return base32.HexEncoding.EncodeToString(convert.ToByteSlice(value))
}

func (*Data) FromBase64(data any) (string, error) {
	str := convert.ToString(data)

	raw, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		if raw, err = base64.URLEncoding.DecodeString(str); err != nil {
			return "", err
		}
	}

	return string(raw), nil
}

func (*Data) ToBase64(value any) string {
	return base64.StdEncoding.EncodeToString(convert.ToByteSlice(value))
}

func (*Data) ToBase64URL(value any) string {
	return base64.URLEncoding.EncodeToString(convert.ToByteSlice(value))
}

func (*Data) FromJSON(data any) (any, error) {
	var out any

	if err := json.Unmarshal(convert.ToByteSlice(data), &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return out, nil
}

func (*Data) ToJSON(value any) (string, error) {
	out, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}

	return string(out), nil
}

func (*Data) ToJSONPretty(value any) (string, error) {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}

	return string(out), nil
}

func (*Data) FromYAML(data any) (any, error) {
	var out any

	if err := yaml.UnmarshalWithOptions(convert.ToByteSlice(data), &out, yaml.AllowDuplicateMapKey(), yaml.CustomUnmarshaler(yamlBinaryUnmarshaler)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return out, nil
}

func (*Data) ToYAML(value any) (string, error) {
	out, err := yaml.MarshalWithOptions(value, yaml.UseLiteralStyleIfMultiline(true), yaml.CustomMarshaler(yamlBinaryMarshaler))
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	return string(out), nil
}

func (*Data) ToYAMLFlow(value any) (string, error) {
	out, err := yaml.MarshalWithOptions(value, yaml.Flow(true), yaml.UseLiteralStyleIfMultiline(true), yaml.CustomMarshaler(yamlBinaryMarshaler))
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	return string(out), nil
}

func (*Data) FromTOML(data any) (any, error) {
	var out any

	if err := toml.Unmarshal(convert.ToByteSlice(data), &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal toml: %w", err)
	}

	return out, nil
}

func (*Data) ToTOML(value any) (string, error) {
	out, err := toml.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal toml: %w", err)
	}

	return string(out), nil
}

func (*Data) FromDotEnv(data any) (env.Map, error) {
	var out env.Map

	if err := env.Unmarshal(convert.ToByteSlice(data), &out, env.WithDecoderExpand(false)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return out, nil
}

func (*Data) ToDotEnv(value env.Map) (string, error) {
	out, err := env.Marshal(value, env.WithEncoderExpand(false))
	if err != nil {
		return "", fmt.Errorf("failed to marshal .env: %w", err)
	}

	return string(out), nil
}

func (*Data) FromDotEnvExpanded(data any) (env.Map, error) {
	var out env.Map

	if err := env.Unmarshal(convert.ToByteSlice(data), &out, env.WithDecoderExpand(true), env.WithDecoderLookup(env.Lookup)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return out, nil
}

func (*Data) ToDotEnvExpanded(value env.Map) (string, error) {
	out, err := env.Marshal(value, env.WithEncoderExpand(true), env.WithEncoderLookup(env.Lookup))
	if err != nil {
		return "", fmt.Errorf("failed to marshal .env: %w", err)
	}

	return string(out), nil
}

func (*Data) FromCSV(delim, data any) ([]map[string]string, error) {
	r := csv.NewReader(strings.NewReader(convert.ToString(data)))
	r.Comma = convert.ToRune(delim)
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal csv: %w", err)
	}

	if len(records) == 0 {
		return make([]map[string]string, 0), nil
	}

	var (
		headers = records[0]
		result  = make([]map[string]string, 0, len(records))
	)

	for _, row := range records[1:] {
		record := make(map[string]string)

		for i, val := range row {
			if i < len(headers) {
				record[headers[i]] = val
			}
		}

		result = append(result, record)
	}

	return result, nil
}

func (*Data) ToCSV(delim, data any) (string, error) {
	var b bytes.Buffer

	w := csv.NewWriter(&b)
	w.Comma = convert.ToRune(delim)

	d := convert.ToSlice(data, convert.ToStringStringMap)

	if len(d) == 0 {
		return "", nil
	}

	var (
		headers = slices.Collect(maps.Keys(d[0]))
		records = make([][]string, 0, len(d)+1)
	)

	records = append(records, headers)

	for _, row := range d {
		record := make([]string, len(headers))

		for i, h := range headers {
			record[i] = row[h]
		}

		records = append(records, record)
	}

	if err := w.WriteAll(records); err != nil {
		return "", fmt.Errorf("failed to marshal csv: %w", err)
	}

	return b.String(), nil
}

func yamlBinaryMarshaler(b []byte) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`!!binary "`)
	buf.Write(base64.StdEncoding.AppendEncode(nil, b))
	buf.WriteByte('"')

	return buf.Bytes(), nil
}

func yamlBinaryUnmarshaler(d *[]byte, r []byte) error {
	var s string

	if err := yaml.Unmarshal(r, &s); err != nil {
		return err
	}

	*d = []byte(s)

	return nil
}
