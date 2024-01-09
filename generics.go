package siocore

import "golang.org/x/exp/constraints"

type Numbers interface {
	constraints.Integer | constraints.Float | constraints.Complex
}
