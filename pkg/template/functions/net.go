package functions

import (
	"fmt"
	"net"

	"github.com/JFAexe/tem/pkg/convert"
)

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
		if h, ok := args[0].(*NetHost); ok && h != nil {
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

func (*Net) SplitHostPort(hostport any) (*NetHost, error) {
	host, port, err := net.SplitHostPort(convert.ToString(hostport))
	if err != nil {
		return nil, fmt.Errorf("failed to split host-port: %w", err)
	}

	return &NetHost{
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
