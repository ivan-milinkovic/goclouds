package main

type FlipBook[T any] struct {
	stack         []Matrix2D[T]
	depth         int
	width, height int
}

func NewFlipBook[T any](depth int, w, h int) *FlipBook[T] {
	stack := make([]Matrix2D[T], depth)
	for i := range depth {
		InitMatrix2D(&stack[i], w, h)
	}
	return &FlipBook[T]{
		stack:  stack,
		depth:  depth,
		width:  w,
		height: h,
	}
}
