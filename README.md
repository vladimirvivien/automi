Automi
======
A stream API for the Go programming language (alpha)

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi is an API for processing streams of data using idiomatic Go.  Using Automi, programs can process streaming of data chunks by composing stages of operations that are applied to each element of the stream.  

![Stream](./docs/streaming.png)

The Automi API expresses a stream with four primitives including:

- *An emitter*: an in-memory, network, or file resource that can emit elements for streaming
- *The stream*: represents a conduit whithin which data elements are streamed
- *Stream operations*: code which can be attached to the stream to process streamed elements
- *A collector*: an in-memory, network, or file resource that can collect streamed data.

Automi streams use Go channels internally to route data.  This means Automi streams automatically support features such as buffering, automatic back-pressure queuing, and concurrency safety.

# Using Automi

This section explores several examples which show the ease-of-use of Automi.  All examples are found in the [./example](./examples) directory.

## A simple rune filter

This example shows how simple it is to compose and express stream operations with Automi. This example uses a slice of runes, as collector source, for a stream.  It then filters  out unwanted rune values, sort the remaining items, and lastly prints the characters.

```go
func main() {
	// create stream with emitter of rune slice
	strm := stream.New([]rune("B世!ぽ@opqDQRS#$%^&*()ᅖ...O6PTnVWXѬYZbcef7ghijCklrAstvw"))

    // apply stream operations
	strm.Filter(func(item rune) bool {
		return item >= 65 && item < (65+26) // filter out unwanted chars
	}).Map(func(item rune) string {
		return string(item) // map rune to string
	}).Batch().Sort() // batch and sort result
	strm.Into(collectors.Writer(os.Stdout)) // send result to stdout

	// open the stream
	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
```

See full [source code](./examples/emitters/slice0).

Let's decompose the program to see what's going on.

#### Stream.New()

`stream.New(...)` - Creates a new stream with a (rune) slice emitter which will emit random rune values unot the stream.

```go
func main() {
	strm := stream.New([]rune("B世!ぽ@opqDQRS#$%^&*()ᅖ...O6PTnVWXѬYZbcef7ghijCklrAstvw"))
...
}
```

#### Apply stream operations

`stream.{Filter,Map,Batch/Sort}` - Next, operations are applied to the stream to filter out unwanted runes that have values less than 65 and greater than 91, map each streamed rune value to its correspoding to string value, then lastly sort the remaining runes.  

```go
func main(){
...
    strm.Filter(func(item rune) bool {
		return item >= 65 && item < (65+26)
	}).Map(func(item rune) string {
		return string(item)
	}).Batch().Sort()
...
}
```

#### Collect the stream
`Stream.Into` - routes the stream to a collector. This example uses an io.Writer collector which is used to output the collected streamed items to `os.Stdout`.

```go
func main(){
...
    strm.Into(collectors.Writer(os.Stdout))
...
}
```

#### Open the stream

`Stream.Open` - After the stream is composed (above), it can be opened.  This runs starts the emitter and execute the operations against the stream.

```go
func main(){
...
  if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
  }  
...
}
```
## Stream from an io.Reader

The next example shows how to use Automi to stream data from an emitter that implements`io.Reader`.  While the example uses an in-memory source, this should work with any value that implements io.Reader including file and network sources.

```go
func main() {
	data := `"request", "/i/a", "00:11:51:AA", "accepted"
	"response", "/i/a/", "00:11:51:AA", "served"
...
	"response", "/i/a", "00:BB:22:DD", "served"`

    // create io.Reader
	reader := strings.NewReader(data)
    
	// create stream with reader emitter,
	// which buffers data as 50-byte chunks.
	stream := stream.New(emitters.Reader(reader).BufferSize(50))

    // map each 50-byte []byte chunk to string
	stream.Map(func(chunk []byte) string {
		str := string(chunk)
		return str
	})

	// filter out string values with `request`
	stream.Filter(func(e string) bool {
		return (strings.Contains(e, `"response"`))
	})

	// route result in a collector function which prints it
	stream.Into(collectors.Func(func(data interface{}) error {
		e := data.(string)
		fmt.Println(e)
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
```

See complete example [here]("./examples/emitters/reader/emitreader.go").
