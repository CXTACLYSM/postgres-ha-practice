package slow

import "context"

type Query struct{}

type Handler interface {
	Handle(context.Context, Query) error
}
