package functions

import (
	"net"

	"github.com/JFAexe/tem/pkg/convert"
)

type IP struct{}

func (*IP) Equal(other, value any) bool {
	var (
		a = convert.ToIP(value)
		b = convert.ToIP(other)
	)

	return (a == nil && b == nil) || a.Equal(b)
}

func (*IP) IPVersion(value any) int {
	switch ip := convert.ToIP(value); {
	case ip.To4() != nil:
		return 4
	case ip.To16() != nil:
		return 6
	default:
		return 0
	}
}

func (*IP) ToIPv4(value any) net.IP {
	ip := convert.ToIP(value)

	if v4 := ip.To4(); v4 != nil {
		return v4
	}

	return nil
}

func (*IP) ToIPv6(value any) net.IP {
	ip := convert.ToIP(value)

	if v4 := ip.To4(); v4 == nil {
		return ip.To16()
	}

	return nil
}

func (f *IP) IsIPv4(value any) bool {
	return f.IPVersion(value) == 4
}

func (f *IP) IsIPv6(value any) bool {
	return f.IPVersion(value) == 6
}

func (*IP) IsUnspecified(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsUnspecified()
}

func (*IP) IsLoopback(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsLoopback()
}

func (*IP) IsPrivate(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsPrivate()
}

func (*IP) IsMulticast(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsMulticast()
}

func (*IP) IsGlobalUnicast(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsGlobalUnicast()
}

func (*IP) IsLinkLocalUnicast(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsLinkLocalUnicast()
}

func (*IP) IsLinkLocalMulticast(value any) bool {
	ip := convert.ToIP(value)

	return ip != nil && ip.IsLinkLocalMulticast()
}
