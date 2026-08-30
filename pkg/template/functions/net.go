package functions

import (
	"fmt"
	"net"

	"github.com/JFAexe/tem/pkg/convert"
)

var ipPredicates = map[string]func(net.IP) bool{
	"unspecified":        net.IP.IsUnspecified,
	"loopback":           net.IP.IsLoopback,
	"private":            net.IP.IsPrivate,
	"multicast":          net.IP.IsMulticast,
	"globalunicast":      net.IP.IsGlobalUnicast,
	"linklocalunicast":   net.IP.IsLinkLocalUnicast,
	"linklocalmulticast": net.IP.IsLinkLocalMulticast,
}

type NetHost struct {
	Host string `json:"host" yaml:"host" toml:"host"`
	Port string `json:"port" yaml:"port" toml:"port"`
}

func (h NetHost) String() string {
	return net.JoinHostPort(h.Host, h.Port)
}

type Net struct{}

func (*Net) JoinHostPort(args ...any) (string, error) {
	var host, port string

	switch len(args) {
	case 0:
		return "", fmt.Errorf("%w: expected host+port, or port, or host, port", ErrValueRequired)
	case 1:
		if h, ok := args[0].(NetHost); ok {
			host = h.Host
			port = h.Port
		} else {
			port = convert.ToString(args[0])
		}
	case 2:
		host = convert.ToString(args[0])
		port = convert.ToString(args[1])
	default:
		return "", fmt.Errorf("%w: expected host+port, or port, or host, port", ErrTooManyArguments)
	}

	return net.JoinHostPort(host, port), nil
}

func (*Net) SplitHostPort(hostport any) (NetHost, error) {
	host, port, err := net.SplitHostPort(convert.ToString(hostport))
	if err != nil {
		return NetHost{}, fmt.Errorf("failed to split host-port: %w", err)
	}

	return NetHost{
		Host: host,
		Port: port,
	}, nil
}

func (*Net) ParseMAC(value any) (net.HardwareAddr, error) {
	mac, err := net.ParseMAC(convert.ToString(value))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC: %w", err)
	}

	return mac, nil
}

func (*Net) ToIPv4(value any) net.IP {
	ip := convert.ToIP(value)

	if v4 := ip.To4(); v4 != nil {
		return v4
	}

	return nil
}

func (*Net) ToIPv6(value any) net.IP {
	ip := convert.ToIP(value)

	if v4 := ip.To4(); v4 == nil {
		return ip.To16()
	}

	return nil
}

func (*Net) IsIP(args ...any) (bool, error) {
	switch len(args) {
	case 0:
		return false, fmt.Errorf("%w: expected ip, or type, ip", ErrValueRequired)
	case 1:
		return convert.ToIP(args[0]) != nil, nil
	case 2:
		kind := normalizeString(args[0])

		pred, ok := ipPredicates[kind]
		if !ok {
			return false, fmt.Errorf("invalid ip kind %#q, supported: %s", kind, joinKeys(ipPredicates))
		}

		ip := convert.ToIP(args[1])
		if ip == nil {
			return false, nil
		}

		return pred(ip), nil
	default:
		return false, fmt.Errorf("%w: expected ip, or type, ip", ErrTooManyArguments)
	}
}

func (*Net) IsIPv4(value any) bool {
	return ipVersion(value) == 4
}

func (*Net) IsIPv6(value any) bool {
	return ipVersion(value) == 6
}

func (*Net) IPVersion(value any) int {
	return ipVersion(value)
}

func (*Net) Equal(other, value any) bool {
	var (
		a = convert.ToIP(value)
		b = convert.ToIP(other)
	)

	return a.Equal(b)
}

func ipVersion(value any) int {
	switch ip := convert.ToIP(value); {
	case ip.To4() != nil:
		return 4
	case ip.To16() != nil:
		return 6
	default:
		return 0
	}
}
