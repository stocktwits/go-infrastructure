package stmocks

import "context"

type key int

var mockKey key

func NewMockContext(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, mockKey, value)
}

func FromMockContext(ctx context.Context) (string, bool) {
	err, ok := ctx.Value(mockKey).(string)
	return err, ok
}
