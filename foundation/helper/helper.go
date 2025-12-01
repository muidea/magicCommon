package helper

import "context"

func GetValueFromContext[T any](ctx context.Context, key any) (ret T, ok bool) {
	val := ctx.Value(key)
	if val == nil {
		return
	}

	ret, ok = val.(T)
	return
}
