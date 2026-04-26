package functions

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha3"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"hash/crc32"
	"maps"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/goccy/go-yaml"

	"github.com/JFAexe/tem/pkg/env"
)

var hashers = map[string]hash.Hash{
	"crc32":      crc32.NewIEEE(),
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

type Data struct {
	envs env.Store
}

func (*Data) Xor(key string, value any) string {
	data := []byte(ToString(value))

	for i := range data {
		data[i] ^= key[i%len(key)]
	}

	return string(data)
}

func (*Data) Hash(kind string, value any) (string, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))

	hasher, ok := hashers[kind]
	if !ok {
		return "", fmt.Errorf("invalid hash function %#q, supported: %s", kind, strings.Join(slices.Sorted(maps.Keys(hashers)), ", "))
	}

	return hex.EncodeToString(hasher.Sum([]byte(ToString(value)))), nil
}

func (*Data) FromHex(data string) (string, error) {
	raw, err := hex.DecodeString(data)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (*Data) ToHex(value any) string {
	return hex.EncodeToString([]byte(ToString(value)))
}

func (*Data) FromBase64(data string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		if raw, err = base64.URLEncoding.DecodeString(data); err != nil {
			return "", err
		}
	}

	return string(raw), nil
}

func (*Data) ToBase64(value any) string {
	return base64.StdEncoding.EncodeToString([]byte(ToString(value)))
}

func (*Data) ToBase64URL(value any) string {
	return base64.URLEncoding.EncodeToString([]byte(ToString(value)))
}

func (*Data) FromJSON(data string) (any, error) {
	var out any

	if err := json.Unmarshal([]byte(data), &out); err != nil {
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

func (*Data) FromYAML(data string) (any, error) {
	var out any

	if err := yaml.Unmarshal([]byte(data), &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return out, nil
}

func (*Data) ToYAML(value any) (string, error) {
	out, err := yaml.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal yaml: %w", err)
	}

	return string(out), nil
}

func (*Data) FromTOML(data string) (any, error) {
	var out any

	if err := toml.Unmarshal([]byte(data), &out); err != nil {
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

func (*Data) FromDotEnv(data string) (any, error) {
	var out env.Map

	if err := env.Unmarshal([]byte(data), &out, env.WithDecoderExpand(false)); err != nil {
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

func (f *Data) FromDotEnvExpanded(data string) (any, error) {
	var out env.Map

	if err := env.Unmarshal([]byte(data), &out, env.WithDecoderExpand(true), env.WithDecoderLookup(f.envs.Lookup)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	return out, nil
}

func (f *Data) ToDotEnvExpanded(value env.Map) (string, error) {
	out, err := env.Marshal(value, env.WithEncoderExpand(true), env.WithEncoderLookup(f.envs.Lookup))
	if err != nil {
		return "", fmt.Errorf("failed to marshal .env: %w", err)
	}

	return string(out), nil
}
