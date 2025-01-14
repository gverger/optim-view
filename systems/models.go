package systems

type ShapeTransform struct {
	Id int
	X  float32
	Y  float32
}

type DisplayableNode struct {
	Id   uint64
	Text string

	Transform []ShapeTransform
}
