package types

import "math"

type Vector struct {
	X, Y float64
}

func (v Vector) Add(u Vector) Vector {
	return Vector{v.X + u.X, v.Y + u.Y}
}

func (v Vector) Mul(s float64) Vector {
	return Vector{v.X * s, v.Y * s}
}

func (v Vector) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

func (v Vector) Norm() Vector {
	l := v.Len()
	if l == 0 {
		return Vector{}
	}
	return Vector{v.X / l, v.Y / l}
}

func (v Vector) Eq(u Vector) bool {
	return v.X == u.X && v.Y == u.Y
}
