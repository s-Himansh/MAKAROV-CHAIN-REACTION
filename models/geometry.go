package models

type GeometryType int

const (
	Slab GeometryType = iota
	Sphere
)

type Geometry struct {
	Thickness float64      // cm (slab thickness or sphere radius)
	Type      GeometryType
}
