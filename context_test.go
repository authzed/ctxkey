package ctxkey

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

// ContextHandler is the interface for a "chunk" of work.
type ContextHandler interface {
	Handle(context.Context)
}

// ContextHandlerFunc is a function type that implements ContextHandler
type ContextHandlerFunc func(ctx context.Context)

func (f ContextHandlerFunc) Handle(ctx context.Context) {
	f(ctx)
}

// Handler wraps a ContextHandler and attaches an id to it.
type Handler struct {
	ContextHandler
	id string
}

// NewHandlerFromFunc creates a new Handler from a ContextHandlerFunc.
func NewHandlerFromFunc(h ContextHandlerFunc, id string) Handler {
	return Handler{ContextHandler: h, id: id}
}

func ExampleNew() {
	type ExpensiveComputation struct {
		result string
	}
	CtxExpensiveObject := New[*ExpensiveComputation]()

	useHandler := NewHandlerFromFunc(func(ctx context.Context) {
		// fetch the computed value after the computation
		fmt.Println(CtxExpensiveObject.MustValue(ctx).result)
	}, "use")

	// the compute handler performs some computation that we wish to re-use
	computeHandler := NewHandlerFromFunc(func(ctx context.Context) {
		myComputedExpensiveObject := ExpensiveComputation{result: "computed"}
		ctx = CtxExpensiveObject.Set(ctx, &myComputedExpensiveObject)
		useHandler.Handle(ctx)
	}, "compute")

	ctx := context.Background()
	computeHandler.Handle(ctx)
	// Output: computed
}

func ExampleNewWithDefault() {
	type ExpensiveComputation struct {
		result string
	}
	CtxExpensiveObject := NewWithDefault[*ExpensiveComputation](&ExpensiveComputation{result: "pending"})

	useHandler := NewHandlerFromFunc(func(ctx context.Context) {
		// fetch the computed value after the computation
		fmt.Println(CtxExpensiveObject.MustNonEmptyValue(ctx).result)
	}, "use")

	// the compute handler performs some computation that we wish to re-use
	computeHandler := NewHandlerFromFunc(func(ctx context.Context) {
		// fetch the default value before the computation
		fmt.Println(CtxExpensiveObject.MustNonEmptyValue(ctx).result)

		myComputedExpensiveObject := ExpensiveComputation{result: "computed"}
		ctx = CtxExpensiveObject.Set(ctx, &myComputedExpensiveObject)

		useHandler.Handle(ctx)
	}, "compute")

	ctx := context.Background()
	computeHandler.Handle(ctx)
	// Output: pending
	// computed
}

func ExampleNewBoxedWithDefault() {
	type ExpensiveComputation struct {
		result string
	}
	CtxExpensiveObject := NewBoxedWithDefault[*ExpensiveComputation](nil)

	// the compute handler performs some computation that we wish to re-use
	computeHandler := NewHandlerFromFunc(func(ctx context.Context) {
		myComputedExpensiveObject := ExpensiveComputation{result: "computed"}
		CtxExpensiveObject.Set(ctx, &myComputedExpensiveObject)
	}, "compute")

	decorateHandler := NewHandlerFromFunc(func(ctx context.Context) {
		// adds an empty box
		ctx = CtxExpensiveObject.WithBox()(ctx)

		// fills in the box with the value
		computeHandler.Handle(ctx)

		// returns the unboxed value
		fmt.Println(CtxExpensiveObject.Value(ctx).result)
	}, "decorated")

	ctx := context.Background()
	decorateHandler.Handle(ctx)
	// Output: computed
}

func TestNew(t *testing.T) {
	ctxKey := New[string]()
	ctx := context.Background()

	expectPanic(t, func() {
		_ = ctxKey.MustValue(ctx)
	})

	ctx = ctxKey.Set(ctx, "value")
	if ctxKey.MustValue(ctx) != "value" {
		t.Fatal("expected value")
	}
	value, ok := ctxKey.Value(ctx)
	if !ok {
		t.Fatal("expected ok")
	}
	if value != "value" {
		t.Fatal("expected value")
	}
}

func TestNewWithDefault(t *testing.T) {
	ctxKey := NewWithDefault[*string](nil)
	ctx := context.Background()

	expectPanic(t, func() {
		_ = ctxKey.MustNonEmptyValue(ctx)
	})

	if value := ctxKey.Value(ctx); value != nil {
		t.Fatal("expected default")
	}

	nonDefaultValue := "non-default"

	ctx = ctxKey.Set(ctx, &nonDefaultValue)

	if *ctxKey.MustNonEmptyValue(ctx) != "non-default" {
		t.Fatal("expected value")
	}
	if value := ctxKey.Value(ctx); *value != "non-default" {
		t.Fatal("expected value")
	}

	ctxKeyNonNilDefault := NewWithDefault[*string](&nonDefaultValue)
	if value := ctxKeyNonNilDefault.Value(ctx); *value != "non-default" {
		t.Fatal("expected value")
	}
}

func TestNewBoxedWithDefault(t *testing.T) {
	ctxKey := NewBoxedWithDefault[*string](nil)
	ctx := context.Background()

	if ctxKey.Value(ctx) != nil {
		t.Fatal("expected nil")
	}

	ctx = ctxKey.SetBox(ctx)
	if ctxKey.Value(ctx) != nil {
		t.Fatal("expected nil")
	}

	value := "value"
	ctx = ctxKey.With(&value)(ctx)
	if *ctxKey.Value(ctx) != "value" {
		t.Fatal("expected value")
	}

	ctxKeyNonNilDefault := NewBoxedWithDefault[*string](&value)
	if *ctxKeyNonNilDefault.Value(ctx) != "value" {
		t.Fatal("expected value")
	}

	ctxKeySetBoxAndValueTogether := NewBoxedWithDefault[*string](nil)
	ctx = ctxKeySetBoxAndValueTogether.With(&value)(ctx)
	if *ctxKeySetBoxAndValueTogether.Value(ctx) != "value" {
		t.Fatal("expected value")
	}
}

func TestWith(t *testing.T) {
	key1 := New[string]()
	key2 := NewWithDefault[string]("")
	key3 := NewBoxedWithDefault[string]("")
	key4 := NewWithDefault[string]("testDefault")

	ctx := context.Background()
	ctx = With(
		key1.With("value1"),
		With(
			key2.With("value2"),
			key3.With("value3"),
		),
	)(ctx)
	if key1.MustValue(ctx) != "value1" {
		t.Fatal("expected value1")
	}
	if key2.MustNonEmptyValue(ctx) != "value2" {
		t.Fatal("expected value2")
	}
	if key3.Value(ctx) != "value3" {
		t.Fatal("expected value3")
	}
	if key4.MustNonEmptyValue(ctx) != "testDefault" {
		t.Fatal("expected testDefault")
	}
}

func expectPanic(t testing.TB, f func()) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if recover() == nil {
				t.Error("expected panic")
				return
			}
		}()
		f()
	}()
	wg.Wait()
}
