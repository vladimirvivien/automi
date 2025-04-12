<h1 align="center">
    <img src="./docs/automi-logo.png" alt="automi">
</h1>

<h4 align="center">Stream Processing API for Go</h4>
<br/>

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Workflow](https://github.com/vladimirvivien/automi/actions/workflows/build-test.yaml/badge.svg)](https://github.com/vladimirvivien/automi/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vladimirvivien/automi)](https://goreportcard.com/report/github.com/vladimirvivien/automi)

Automi is a Go package for processing streams of data. It enables composing powerful data pipelines by chaining operations that are applied to each element flowing through the stream. The project has been completely re-implemented to leverage Go's generics support, bringing significant advantages:

- **Type safety across the entire pipeline**: Compile-time type checking eliminates runtime type assertions and conversions
- **Enhanced developer experience**: Better IDE autocompletion and error detection during development
- **Reduced boilerplate**: No more manual type casting or interface{} conversions
- **Improved performance**: Elimination of runtime type checks and assertions leads to more efficient execution

With generics, you can now create strongly-typed streams where each operation's input and output types are verified by the compiler, catching errors early and making your stream processing code more robust and maintainable.

## Concept

Automi implements a data processing pipeline using a streaming architecture. Data flows through a series of connected components that process each element as it passes through. This approach enables efficient handling of large datasets without loading everything into memory at once.

<h1 align="center">
    <img src="./docs/automi-stream.png" alt="Automi streaming concepts">
</h1>



The Automi API is built around four key primitives:

- *Source*: The entry point that emits data elements into the stream. Sources can read from files, channels, slices, or any data provider.
- *Stream*: The central abstraction that coordinates the flow of data. It connects the source to operations and ultimately to a sink.
- *Operations*: Processing steps applied to each element in the stream. These include transformations (map), filters, aggregations, and windowing operations.
- *Sink*: The termination point that consumes the processed data. Sinks can write to files, collect results into slices, or forward data to other systems.

Data flows through this pipeline in a single direction: from source → through operations → to sink. Each operation receives data, processes it, and passes the result to the next stage.

Automi streams use Go channels internally to route data, providing built-in concurrency safety and automatic back-pressure handling. When a downstream component processes data more slowly, upstream components naturally slow down to match the pace, preventing memory overflow.

## Using Automi

Now, let us explore some examples and see how easy it is to use Automi to process data streams.  

>See [automi-eamples](https://github.com/vladimirvivien/automi-examples) for a collection of all Automi examples.

As an introduction to Automi, let us explore a simple example that uses all primitives of the API to compose and express stream operations. 

The following code snippet:
* Streams individual rune values from a slice source
* Applies a `Filter` operation to keep only uppercase ASCII letters
* Maps each rune to its string representation
* Batches the results
* And finally sorts the result alphabetically

```go
func main() {
	// Define slice source
	slice := sources.Slice([]rune(`B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw`))

	// creates stream from the source
	strm := stream.From(slice)

	// Define stream operations
	strm.Run(
		exec.Filter(func(_ context.Context, item rune) bool {
			return item >= 65 && item < (65+26) // remove unwanted chars
		}),
		
		exec.Map(func(_ context.Context, item rune) string {
			return string(item) // map rune to string
		}),

		// batch incoming string items into []string
		window.Batch[string](),

		// sort batched items
		exec.SortSlice[[]string](),
	)

	// Send result to stdout
	strm.Into(sinks.Writer[string](os.Stdout)) 

	// open the stream
	if err := <-strm.Open(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
}
```

> See the complete example [here](https://github.com/vladimirvivien/automi-examples/blob/main/hello-automi/rune1/main.go).

#### How it works
1. The first step defines a stream `Source` from a Go slice using function `sources.Slice`. 

```go
slice := sources.Slice([]rune(`B世!ぽ@opqDQRS#$%^&*()ᅖ4x5Њ8yzUd90E12a3ᇳFGHmIザJuKLMᇙNO6PTnVWXѬYZbcef7ghijCklrAstvw`))
```

2. Next, the code creates a new Automi using the stream source. Each element of the slice source will be emitted individually to the stream. 

```go
strm := stream.From(slice)
```

3. Next, we define `Filter`, `Map`, `Batch`, and `Sort` operations to be applied to each item as it streams.

```go
	// Define stream operations
	strm.Run(
		exec.Filter(func(_ context.Context, item rune) bool {
			return item >= 65 && item < (65+26) // remove unwanted chars
		}),

		exec.Map(func(_ context.Context, item rune) string {
			return string(item) // map rune to string
		}),

		// batch incoming string items into []string
		window.Batch[string](),

		// sort batched items
		exec.SortSlice[[]string](),
	)

```

4. Next, we define a `Sink` at the end of the stream to collect the result in a Go `io.Writer` which streams the as a string item into standard output:

```go
strm.Into(sinks.Writer[string](os.Stdout)) 
```

5. Lastly, the code opens the stream to start it:

```go
	if err := <-strm.Open(context.Background()); err != nil {
		fmt.Println(err)
		return
	}
```


#### Example: streaming from a CSV file
Let's explore another example that streams data from a CSV source file. Each CSV row will be 
* Mapped to a Go struct type
* Filtered by value
* Then mapped to a slice of strings which is then collected into another CSV file. 

```go
type scientist struct {
	FirstName string
	LastName  string
	Title     string
	BornYear  int
}

func main() {
	// Source csv
	src, _ := os.Open("./data.txt")
    source := sources.CSV(src)

	// Sink csv
	snk, _ := os.Create("./result.txt")
    sink := sinks.CSV(snk)

    // Start new stream from source
	stream := stream.From(source)

    // setup execution operations
	stream.Run(
		// map csv row to struct scientist
		exec.Map(func(ctx context.Context, cs []string) scientist {
			yr, _ := strconv.Atoi(cs[3])
			return scientist{
				FirstName: cs[1],
				LastName:  cs[0],
				Title:     cs[2],
				BornYear:  yr,
			}
		}),

		// apply data filter
		exec.Filter(func(ctx context.Context, cs scientist) bool {
			return (cs.BornYear > 1930)
		}),

		// remap value of type scientst to []string
		exec.Map(func(ctx context.Context, cs scientist) []string {
			return []string{cs.FirstName, cs.LastName, cs.Title}
		}),
	)

	// stream result into sink
	stream.Into(sink)
}
```

> See the complete example [here](https://github.com/vladimirvivien/automi-examples/tree/main/customtype).

#### Example: streaming HTTP requests and responses
This example showcases the versatility of Automi by streaming and processing data from HTTP requests and responses. 
The example is an HTTP server program that streams data from the request Body, encodes it using base64, and streams the result into the HTTP response:

```go
	http.HandleFunc(
		"/",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Add("Content-Type", "text/html")
			resp.WriteHeader(http.StatusOK)

			// setup new stream with HTTP body as source
			strm := stream.From(sources.Reader(req.Body))
			strm.Run(
				exec.Execute(func(_ context.Context, data []byte) string {
					return base64.StdEncoding.EncodeToString(data)
				}),
			)

			// route result into response
			strm.Into(sinks.Writer[[]byte](resp))

			// run the stream
			if err := <-strm.Open(req.Context()); err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				slog.Error("Stream failed to open", "error", err)
			}
		},
	)
```
> See the complete example [here](https://github.com/vladimirvivien/automi-examples/tree/main/net/http).

## Automi components
Automi comes with a set of built-in components to get you started with stream processing including the followings.

### Sources
* `sources.Chan` - source from Go channels
* `sources.CSV` - source from CSV
* `io.Reader` - source from io.Reader
* `io.Scanner` - source from io.Scanner
* `Slice` - source from Go slices

### Windowing
* `window.Batch` - batch all
* `window.Size` - widow by size
* `windows.Duration` window by duration

### Function operators
* `exec.Execute` - user-defined funcion
* `exec.Filter` - filter func
* `exec.Map` - map func

### Aggregation operators

#### Group 
* `exec.GroupByIndex` - goup by slice index
* `exec.GroupByStructField` - group by struct field name
* `exec.GroupByMapKey` - group by map key value

#### Sum
* `exec.SumByIndex` - sum by slice index
* `exec.SumByStructField` - sum by struct field name
* `exec.SumByMapKey` - sum by map key value
* `exec.Sum` - sum either 1D or 2D slice
* `exec.SumAll1D` - sum 1D slice
* `exec.SumAll2D` - sum 2D slice

#### Sort
* `exec.SortSlice` - sort slice
* `exec.SortSliceByIndex` - sort by slice index
* `exec.SortByStructField` - sort by struct field name
* `exec.SortByMapKey` - sort by map key value
* `exec.SortWithFunc` - sort with user-defined func

### Sinks

* `sinks.CSV` - sink into CSV
* `sinks.Func` - sink into a user-defined func
* `sinks.Discard` - no op sink
* `sinks.Slice` - sink into Go slice
* `sinks.Slog` - sink into Go slog Logger
* `sinks.Writer` - sink into Go io.Writer

## Previous version
The previous version (v0.1.0) of Automi, which uses Go reflection, has been moved
to branch `v0.1.0-automi-reflection` and will not be maintained.

## Licence
MIT
