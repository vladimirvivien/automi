# Automi

Automi is a type-safe Go package for building efficient data stream processing pipelines. Using Go's generics, it provides a composable API that enables developers to process data streams with compile-time type checking, eliminating runtime errors common in reflection-based approaches.

### Key Features

- **Type Safety**: Leverages Go's generics for compile-time verification of operation types
- **Composable Design**: Chain multiple operations into complex data transformation pipelines
- **Concurrency Support**: Built on Go's concurrency primitives for efficient concurrent processing
- **Extensible Architecture**: Create custom sources, operators, and sinks to fit your use case

### Core Components

The Automi API consists of four main components:

1. **Sources**: Components that emit data into a stream (CSV files, slices, channels, etc.)
2. **Stream**: Central component that coordinates data flow through the pipeline
3. **Operators**: Transformations applied to each element (map, filter, reduce, etc.)
4. **Sinks**: Terminal components that collect or output the processed data

### Typical Usage Pattern

```go
// 1. Create a stream with a source
stream := stream.From(sources.Slice(data))

// 2. Add processing operators
stream.Run(
    exec.Filter(filterFunc),
    exec.Map(transformFunc),
    // Add more operators as needed
)

// 3. Define where to send the results
stream.Into(sinks.Func(handleResults))

// 4. Open the stream for processing
<-stream.Open(context.Background())
```

Continue reading to learn how to build powerful data processing pipelines with minimal code.

## The Automi Stream

The Stream is the central component in Automi's architecture. It orchestrates the flow of data through your processing pipeline, connecting sources, operators, and sinks into a cohesive data processing unit.

### Stream Architecture

A Stream represents a complete data processing pipeline with three key parts:

1. **Source**: The entry point that emits data into the pipeline
2. **Operators**: A sequence of transformations applied to each data element
3. **Sink**: The terminal component that consumes the processed data

Think of a Stream as a conveyor belt where data items flow from the source, through various processing stations (operators), before reaching their final destination (sink).

### Stream Lifecycle

Streams follow a clear lifecycle:

1. **Creation**: Initialize with a source using `stream.From(source)`
2. **Configuration**: Add operations with `stream.Run(op1, op2, ...)` 
3. **Termination**: Define output destination with `stream.Into(sink)`
4. **Execution**: Begin processing with `stream.Open(ctx)`

Every stream node executes asynchronously, leveraging Go's concurrency model for efficient processing. The processing concludes when either the source is exhausted or the context is canceled.

### Stream Type Safety

Streams in Automi leverage Go's generics to maintain type safety throughout the pipeline. Each operation in the chain accepts the output type of the previous operation as its input type:

```go
results := make([]string)
// Creating a type-safe stream pipeline
stream.From(sources.Slice[[]int]([]int{1, 2, 3, 4, 5})).
    Run(
        // This operator takes int and returns string
        exec.Map[int](func(_ context.Context, i int) string {
            return fmt.Sprintf("%d", i)
        }),
        // This operator takes string and returns string
        exec.Filter[string](func(_ context.Context, s string) bool {
            return strings.Contains(s, "3") || strings.Contains(s, "5")
        }),
    ).
    // Receives a tring
    Into(sinks.Slice[string](results))
```

The example above explicitly uses Go generic type instantiation to specify the type parameter for each operation. However, Go's compiler can often infer these types automatically based on the context. This allows for a more concise syntax:

```go
results := make([]string)
// Creating a type-safe stream pipeline with type inference
stream.From(sources.Slice([]int{1, 2, 3, 4, 5})).
    Run(
        // Compiler infers that this takes int and returns string
        exec.Map(func(_ context.Context, i int) string {
            return fmt.Sprintf("%d", i)
        }),
        // Compiler infers that this takes string and returns string
        exec.Filter(func(_ context.Context, s string) bool {
            return strings.Contains(s, "3") || strings.Contains(s, "5")
        }),
    ).
    Into(sinks.Slice(results))
```

## Sources

A `Source` is the entry point of an Automi stream that produces data items into the pipeline. Sources implement the `Emitter` interface, allowing them to send data through the stream:

```go
type Source interface {
    Emitter
    Open(context.Context) error  // Initiates data production with context support
}

type Emitter interface {
    GetOutput() <-chan any       // Returns the channel through which data flows
}
```

A source implementation emits data items on its output channel. The Automi runtime continuously consumes data from this channel until it's closed. Once closed, the runtime initiates an orderly shutdown of the entire stream.

### Example: Creating a Stream with a Source

The following example demonstrates using a `Slice` source to create a stream of log entries:

```go
func TestSliceSource(t *testing.T) {
    // Create a source from a slice of string slices (log entries)
    src := sources.Slice([][]string{
        {"request", "/i/a", "00:11:51:AA", "accepted"},
        {"response", "/i/a/", "00:11:51:AA", "served"},
        {"request", "/i/b", "00:11:22:33", "accepted"},
        {"response", "/i/b", "00:11:22:33", "served"},
        {"request", "/i/c", "00:11:51:AA", "accepted"},
    })

    // Create a stream using the source
    strm := stream.From(src)
    
    // Continue with stream operations...
}
```

In this example, Automi will stream each element of type `[]string` through the pipeline.

### Built-in Sources

Automi provides several ready-to-use source implementations in the `sources` package:

* `sources.Slice[T]` - Emits items from Go slices of type `[]T`
* `sources.Chan[T]` - Emits items from a Go channel of type `chan T`
* `sources.CSV` - Emits rows from CSV source as `[]string` slices
* `sources.Reader` - Emits data from a `io.Reader` as []byte
* `sources.Scanner` - Emits `[]byte` from a `io.Reader` tokenized with `bufio.SplitFunc`

## Operators

Operators are the transformational components that process data as it flows through a stream. They sit between sources and sinks, applying operations like filtering, mapping, or aggregating elements in the stream. Each operator:

- Consumes data from an upstream component (via `Collector` interface)
- Processes the data according to its logic 
- Emits transformed data to the next component (via `Emitter` interface)

```go
type Operator interface {
	Collector
	Emitter
	Exec(context.Context) error
}
```

### Adding Operators to a Stream

The `stream.Stream` type provides the `Run()` method to add one or more operators to your data pipeline. Operators are executed in the order they're added:

```go
func main() {
	slice := sources.Slice([]rune(`B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw`))

	strm := stream.From(slice)

	// Define stream operations
	strm.Run(
		// Filter: keeps only uppercase English letters (A-Z)
		exec.Filter(func(_ context.Context, item rune) bool {
			return item >= 65 && item < (65+26) // ASCII range for uppercase letters
		}),
		
		// Map: converts each rune to a string
		exec.Map(func(_ context.Context, item rune) string {
			return string(item) 
		}),

		// Batch: collects individual strings into a slice
		window.Batch[string](),

		// Sort: arranges the batched strings alphabetically
		exec.SortSlice[[]string](),
	)
}
```

### Built-in Operators

Automi provides several ready-to-use operators in the `exec` package:
* `exec.Execute[T,R](func(T)R)` - Transforms each element from type T to type R with a user-defined function
* `exec.Map[T,R](func(T)R)` - Transforms each element from type T to type R with a user-defined function
* `exec.Filter[T](func(T)bool)` - Keeps only elements that satisfy a predicate user-defined function

## Sinks
An Automi `Sink` is a terminal component in a stream pipeline, responsible for collecting and processing streamed items. It implements the `Collector` interface to receive data and the `Sink` interface to manage its lifecycle:

```go
type Sink interface {
    Collector
    Open(context.Context) <-chan error
}

type Collector interface {
    SetInput(<-chan any)
}
```

Sinks serve as the endpoint of a stream, where data is stored, processed, or forwarded to external systems such as files, databases, or in-memory structures. To ensure smooth data flow, sinks must promptly consume incoming items to prevent back-pressure. The Automi runtime monitors the sink's state to determine when to close the stream.

### Example: Using a `Func` Sink
The following example demonstrates a `Func` sink, which allows you to define a custom function to handle streamed items:

```go
func TestSinkFunc(t *testing.T) {
    src := sources.Slice([][]string{
        {"request", "/i/a", "00:11:51:AA", "accepted"},
        {"response", "/i/a/", "00:11:51:AA", "served"},
        {"request", "/i/b", "00:11:22:33", "accepted"},
        {"response", "/i/b", "00:11:22:33", "served"},
        {"request", "/i/c", "00:11:51:AA", "accepted"},
    })

    strm := From(src)

    strm.Into(sinks.Func(func(item []string) error {
        // Process each item here
        return nil
    }))
    // ...additional code...
}
```

### Built-in Sinks
Automi provides several built-in sinks to suit various use cases:

- `sinks.CSV`: Writes items in CSV format
- `sinks.Func[T](func(T)error)`: Processes items using a user-defined function
- `sinks.Discard`: Ignores all items (no-op sink)
- `sinks.Slice[T]`: Appends items of type T to a Go slice
- `sinks.Slog`: Logs items using Go's `slog` package
- `sinks.Writer[[]byte|string]`: Writes items (of type `[]byte` or `string`) to an `io.Writer`

## Aggregation

Automi provides powerful data aggregation capabilities through its batching mechanisms. Batches (also called windows) allow you to collect multiple stream items together before applying operations to the entire group.

### Understanding Batches

A batch in Automi:
- Acts as an intermediate operator that collects individual stream items
- Groups these items based on specified criteria or trigger conditions
- Emits the entire collection as a single unit for downstream processing

Batching is particularly useful for operations that need to see multiple items at once, such as sorting, grouping, or statistical analysis.

### Basic Batching Example

```go
// Create a stream from a slice of integers
strm := stream.From(sources.Slice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))

// Use window.Batch() to collect all items into a single slice
// Then apply a sort operation to the batched items
strm.Run(
    window.Batch[int](),         // Collects all integers into a []int
    exec.SortSlice[[]int](),     // Sorts the collected integers
)

// Sink the result to a slice
var result []int
strm.Into(sinks.Slice(result))

// Open the stream and wait for completion
<-strm.Open(context.Background())
// result now contains [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
```

### Operating on Batches

Automi offers a suite of specialized operators designed to work with batched items or slices. These operators perform sophisticated transformations on collections, enabling complex data processing with minimal code.

### Built-in Batch Operators

Automi provides several operators to process batched data:

#### Grouping Operators
- `exec.GroupByKey[K, V]` - Groups items of type `[]map[K]V` by their key values
- `exec.GroupByName[T]` - Groups items of type `[]T` (where T is a struct) by a specified field
- `exec.GroupByPos[T]` - Groups items of type `[]T` or `[][]T` by a specified position/index

#### Sorting Operators
- `exec.SortByKey[K, V]` - Sorts items of type `[]map[K]V` by their key values
- `exec.SortByName[T]` - Sorts items of type `[]T` (where T is a struct) by a specified field
- `exec.SortByPos[T]` - Sorts items of type `[]T` or `[][]T` by a specified position/index
- `exec.SortSlice[T]` - Sorts a slice of type `T` using Go's built-in `sort.Slice`
- `exec.SortWith[T]` - Sorts a slice using a custom comparison function

#### Aggregation Operators
- `exec.Sum[T]` - Computes the sum of numeric items in `[]T`
- `exec.SumByKey[K, V]` - Sums numeric values in `[]map[K]V` by key
- `exec.SumByName[T]` - Sums a numeric field across structs in `[]T`
- `exec.SumByPos[T]` - Sums values at a specific position in `[]T` or `[][]T`

### Grouping Example

Here's an example demonstrating how to group items by a struct field:

```go
// Sample data with ID and value fields
data := []struct{
    ID  string
    Val int
}{
    {"A", 1}, {"B", 2}, {"C", 3},
    {"A", 4}, {"B", 5}, {"C", 6},
    {"A", 7}, {"D", 8}, {"E", 9},
}

// Create a stream from the data
strm := stream.From(sources.Slice(data))

// Group items by ID field, creating a map[string][]struct{...}
strm.Run(
    window.Batch[struct{ID string; Val int}](),
    exec.GroupByName[[]struct{ID string; Val int}]("ID"),
)

// Define a map to store the result
result := make(map[string][]struct{ID string; Val int})
strm.Into(sinks.Func(func(_ context.Context, grouped map[string][]struct{ID string; Val int}) error {
    result = grouped
    return nil
}))

<-strm.Open(context.Background())
// result now contains a map with keys "A", "B", "C", "D", "E" and their associated items
```

### Batch trigger control

Automi allows you to control when batches trigger/emit with various strategies:

- `window.Batch[T]()` - Batches all items until the source is exhausted
- `window.BySize[T](n int)` - Emits batches after collecting n items
- `window.ByDuration[T](d time.Duration)` - Emits batches after a specified duration
- `window.ByFunc[T](f func(T) bool)` - Emits batches when the predicate function returns true