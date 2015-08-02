package main

import (
	"github.com/barnex/fmath"
)

type Vec2 [2]float32
type Vec3 [3]float32

func (v Vec3) Translate(v2 Vec3) Vec3 {
	return Vec3{v[0] + v2[0], v[1] + v2[1], v[2] + v2[2]}
}

func (v Vec3) Scale(t float32) Vec3 {
	return Vec3{v[0] * t, v[1] * t, v[2] * t}
}

func (v Vec2) Translate(v2 Vec2) Vec2 {
	return Vec2{v[0] + v2[0], v[1] + v2[1]}
}

func (v Vec2) Scale(t float32) Vec2 {
	return Vec2{v[0] * t, v[1] * t}
}

type BoundingBox struct {
	min Vec3
	max Vec3
}

func (b BoundingBox) Intersects(b2 BoundingBox) bool {
	return (fmath.Abs(b.min[0] - b2.min[0]) * 2 < (b.max[0] - b.min[0] + b2.max[0] - b2.min[0]) &&
	  fmath.Abs(b.min[1] - b2.min[1]) * 2 < (b.max[1] - b.min[1] + b2.max[1] - b2.min[1]) &&
	  fmath.Abs(b.min[2] - b2.min[2]) * 2 < (b.max[2] - b.min[2] + b2.max[2] - b2.min[2]))
}

func (b BoundingBox) Translate(v Vec3) BoundingBox {
	return BoundingBox{b.min.Translate(v), b.max.Translate(v)}
}
