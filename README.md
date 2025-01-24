# `ctxkey` - Typed Context Keys for Go

`ctxkey` is a zero-dependency library adding generically typed accessors for `context.Context`.

## Installation

Add `ctxkey` to your project with:

```bash
go get github.com/authzed/ctxkey
```

## Example Usage

See [examples](./examples) for full, runnable examples.

### Basic Usage

Define a unique key for a context value, without type conversion:

```go
ctxMyKey := ctxkey.New[string]()

ctx := ctxMyKey.WithValue(context.Background(), "value")

fmt.Println(ctxMyKey.Value(ctx)) // "value", true 
```

Check if a key has a value, or panic if not found:

```go
ctxMyKey := ctxkey.New[string]()
ctx := ctxMyKey.WithValue(context.Background(), "value")

value, ok := ctxMyKey.Value(ctx)
if !ok {
    panic("value not found")
}

value = ctxMyKey.MustValue(ctx) // "value"
```

Provide a default value if a key is not found:

```go
ctxMyKey := ctxkey.NewWithDefault("defaultValue")

ctx := context.Background()
value := ctxMyKey.Value(ctx) // "defaultValue"

ctx = ctxMyKey.WithValue(ctx, "newValue")
value = ctxMyKey.MustNonEmptyValue(ctx) // "newValue"
```

Provide a box for a key to store a value, useful in cases where it is not convenient to return a context:

```go
var ctxMyKey = ctxkey.NewBoxedWithDefault[string](nil)

func inner(ctx context.Context) {   
    ctxMyKey.WithValue(ctx, "value")
}

func outer() {
    ctx := context.Background()
	ctxMyKey.WithBox(ctx)

    inner(ctx)

    value := ctxMyKey.MustValue(ctx) // "value"
}
```
