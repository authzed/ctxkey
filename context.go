// Package ctxkey provides generic helpers for storing / retrieving values
// from a `context.Context`.
package ctxkey

import (
	"context"
	"fmt"
)

// SettableContext is an interface for a context that can have a value set.
type SettableContext[V any] interface {
	WithValue(ctx context.Context, val V) context.Context
}

// Key is a type that is used as a key in a context.Context for a
// specific type of value V. It mimics the context.Context interface
type Key[V any] struct{}

// New creates a new Key
func New[V any]() *Key[V] {
	return &Key[V]{}
}

// WithValue adds a value to the context for this key.
func (k *Key[V]) WithValue(ctx context.Context, val V) context.Context {
	return context.WithValue(ctx, k, val)
}

// Value retrieves the value from the context for this key. It returns the value
// and a boolean indicating if the value was found.
func (k *Key[V]) Value(ctx context.Context) (V, bool) {
	v, ok := ctx.Value(k).(V)
	return v, ok
}

// MustValue retrieves the value from the context for this key. It panics if the
// value is not found.
func (k *Key[V]) MustValue(ctx context.Context) V {
	v, ok := k.Value(ctx)
	if !ok {
		panic(fmt.Sprintf("could not find value for key %T in context", k))
	}
	return v
}

// DefaultingKey is a type that is used as a key in a context.Context for
// a specific type of value, but returns the default value for V if unset.
type DefaultingKey[V comparable] struct {
	defaultValue V
}

// NewWithDefault creates a new DefaultingKey with the given default value
func NewWithDefault[V comparable](defaultValue V) *DefaultingKey[V] {
	return &DefaultingKey[V]{
		defaultValue: defaultValue,
	}
}

// WithValue adds a value to the context for this key.
func (k *DefaultingKey[V]) WithValue(ctx context.Context, val V) context.Context {
	return context.WithValue(ctx, k, val)
}

// Value retrieves the value from the context for this key. If the value is not
// found, it returns the default value.
func (k *DefaultingKey[V]) Value(ctx context.Context) V {
	v, ok := ctx.Value(k).(V)
	if !ok {
		return k.defaultValue
	}
	return v
}

// MustNonEmptyValue retrieves the value from the context for this key. If the value is
// empty, it panics.
func (k *DefaultingKey[V]) MustNonEmptyValue(ctx context.Context) V {
	v := k.Value(ctx)
	var empty V
	if v == empty {
		panic(fmt.Sprintf("could not find non-nil value for key %T in context", k))
	}
	return v
}

type Box[V any] struct {
	value V
}

// BoxedKey is a type that is used as a key in a
// context.Context that points to a handle containing the desired value.
// This allows a handler higher up in the chain to carve out a spot to be
// filled in by other handlers.
// It can also be used to hold non-comparable objects by wrapping them with a
// pointer.
type BoxedKey[V any] struct {
	defaultValue V
}

// NewBoxedWithDefault creates a new BoxedKey with a default value
func NewBoxedWithDefault[V any](defaultValue V) *BoxedKey[V] {
	return &BoxedKey[V]{
		defaultValue: defaultValue,
	}
}

// WithValue adds a value to the context for this key.
func (k *BoxedKey[V]) WithValue(ctx context.Context, val V) context.Context {
	handle, ok := ctx.Value(k).(*Box[V])
	if ok {
		handle.value = val
		return ctx
	}
	return context.WithValue(ctx, k, &Box[V]{value: val})
}

// WithBox adds a box to the context for this key.
func (k *BoxedKey[V]) WithBox(ctx context.Context) context.Context {
	return context.WithValue(ctx, k, &Box[V]{value: k.defaultValue})
}

// Value retrieves the value from the context for this key. If the value is not
// found, it returns the default value.
func (k *BoxedKey[V]) Value(ctx context.Context) V {
	handle, ok := ctx.Value(k).(*Box[V])
	if !ok {
		return k.defaultValue
	}

	return handle.value
}
