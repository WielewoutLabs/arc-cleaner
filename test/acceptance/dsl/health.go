package dsl

type Health interface {
	Live() error
	Ready() error
}
