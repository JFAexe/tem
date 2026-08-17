package convert

import (
	"cmp"
	"math"
)

const (
	Float32ExactIntMax = 1 << 24
	Float64ExactIntMax = 1 << 53
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Float interface {
	~float32 | ~float64
}

func Clamp[T cmp.Ordered](v, mi, ma T) T {
	return min(max(v, mi), ma)
}

func ClampFloat[F Float, I Int | Uint](v F, mi, ma I) I {
	f := float64(v)

	if f <= float64(mi) {
		return mi
	}

	if f >= float64(ma) {
		return ma
	}

	return I(math.Trunc(f))
}

func SafeInt[T Int](v T) int {
	return int(Clamp(int64(v), math.MinInt, math.MaxInt))
}

func SafeInt8[T Int](v T) int8 {
	return int8(Clamp(int64(v), math.MinInt8, math.MaxInt8))
}

func SafeInt16[T Int](v T) int16 {
	return int16(Clamp(int64(v), math.MinInt16, math.MaxInt16))
}

func SafeInt32[T Int](v T) int32 {
	return int32(Clamp(int64(v), math.MinInt32, math.MaxInt32))
}

func SafeIntToUint8[T Int](v T) uint8 {
	return uint8(Clamp(int64(v), 0, math.MaxUint8))
}

func SafeIntToUint16[T Int](v T) uint16 {
	return uint16(Clamp(int64(v), 0, math.MaxUint16))
}

func SafeIntToUint32[T Int](v T) uint32 {
	return uint32(Clamp(int64(v), 0, math.MaxUint32))
}

func SafeIntToUint64[T Int](v T) uint64 {
	return uint64(max(v, 0))
}

func SafeIntToFloat32[T Int](v T) float32 {
	return float32(Clamp(int64(v), -Float32ExactIntMax, Float32ExactIntMax))
}

func SafeIntToFloat64[T Int](v T) float64 {
	return float64(Clamp(int64(v), -Float64ExactIntMax, Float64ExactIntMax))
}

func SafeUint[T Uint](v T) uint {
	return uint(Clamp(uint64(v), 0, math.MaxUint))
}

func SafeUint8[T Uint](v T) uint8 {
	return uint8(Clamp(uint64(v), 0, math.MaxUint8))
}

func SafeUint16[T Uint](v T) uint16 {
	return uint16(Clamp(uint64(v), 0, math.MaxUint16))
}

func SafeUint32[T Uint](v T) uint32 {
	return uint32(Clamp(uint64(v), 0, math.MaxUint32))
}

func SafeUintToInt8[T Uint](v T) int8 {
	return int8(Clamp(uint64(v), 0, math.MaxInt8))
}

func SafeUintToInt16[T Uint](v T) int16 {
	return int16(Clamp(uint64(v), 0, math.MaxInt16))
}

func SafeUintToInt32[T Uint](v T) int32 {
	return int32(Clamp(uint64(v), 0, math.MaxInt32))
}

func SafeUintToInt64[T Uint](v T) int64 {
	return int64(Clamp(uint64(v), 0, math.MaxInt64))
}

func SafeUintToFloat32[T Uint](v T) float32 {
	return float32(Clamp(uint64(v), 0, Float32ExactIntMax))
}

func SafeUintToFloat64[T Uint](v T) float64 {
	return float64(Clamp(uint64(v), 0, Float64ExactIntMax))
}

func SafeFloat32[T Float](v T) float32 {
	return float32(Clamp(SafeFloat64(v), -math.MaxFloat32, math.MaxFloat32))
}

func SafeFloat64[T Float](v T) float64 {
	f := float64(v)

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}

	return f
}

func SafeFloatToInt8[T Float](v T) int8 {
	return ClampFloat[float64, int8](SafeFloat64(v), math.MinInt8, math.MaxInt8)
}

func SafeFloatToInt16[T Float](v T) int16 {
	return ClampFloat[float64, int16](SafeFloat64(v), math.MinInt16, math.MaxInt16)
}

func SafeFloatToInt32[T Float](v T) int32 {
	return ClampFloat[float64, int32](SafeFloat64(v), math.MinInt32, math.MaxInt32)
}

func SafeFloatToInt64[T Float](v T) int64 {
	return ClampFloat[float64, int64](SafeFloat64(v), math.MinInt64, math.MaxInt64)
}

func SafeFloatToUint8[T Float](v T) uint8 {
	return ClampFloat[float64, uint8](SafeFloat64(v), 0, math.MaxUint8)
}

func SafeFloatToUint16[T Float](v T) uint16 {
	return ClampFloat[float64, uint16](SafeFloat64(v), 0, math.MaxUint16)
}

func SafeFloatToUint32[T Float](v T) uint32 {
	return ClampFloat[float64, uint32](SafeFloat64(v), 0, math.MaxUint32)
}

func SafeFloatToUint64[T Float](v T) uint64 {
	return ClampFloat[float64, uint64](SafeFloat64(v), 0, math.MaxUint64)
}
