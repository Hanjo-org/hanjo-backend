package ptr

import "reflect"

// Ptr returns a pointer to the passed value. Useful for literal constants.
func Ptr[T any](v T) *T {
	return &v
}

// ValOrDefault returns the dereferenced value if non-nil, otherwise def.
func ValOrDefault[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// Val returns the dereferenced value if non-nil, otherwise zero value of T.
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// EmptySlice ensures that if a slice is nil, an empty allocated slice []T{} is returned.
func EmptySlice[T any](s []T) []T {
	if s == nil {
		return make([]T, 0)
	}
	return s
}

// EnsureNonNilSlice inspects data via reflection and replaces nil slices with empty slices.
// Useful for ensuring JSON responses render [] instead of null.
func EnsureNonNilSlice(data any) any {
	if data == nil {
		return nil
	}
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice && val.IsNil() {
		return reflect.MakeSlice(val.Type(), 0, 0).Interface()
	}
	return data
}

// Equal returns true if both pointers are nil or point to equal values.
func Equal[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
