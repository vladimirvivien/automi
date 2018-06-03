# Automi

A stream API for the Go programming language (alpha)

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi is an API for processing streams of data using idiomatic Go.  Using Automi, programs can process streaming of data chunks by composing stages of operations that are applied to each element of the stream.  

## Concept

![Stream](./docs/streaming.png)

The Automi API expresses a stream with four primitives including:

- *An emitter*: an in-memory, network, or file resource that can emit elements for streaming
- *The stream*: represents a conduit whithin which data elements are streamed
- *Stream operations*: code which can be attached to the stream to process streamed elements
- *A collector*: an in-memory, network, or file resource that can collect streamed data.

Automi streams use Go channels internally to route data.  This means Automi streams automatically support features such as buffering, automatic back-pressure queuing, and concurrency safety.

## Using Automi

Now, let us explore some examples to see how easy it is to use Automi to stream and process data.  

>See all examples in the [./example](./examples) directory.

### Example: streaming runes

This example shows how to compose and express stream operations with Automi to stream and filter rune values. It uses a slice of runes, as an emitter, which provides the source for the stream.  Then, stream operators are applied to filter out unwanted rune values and then sort the remaining items. Lastly, a collector is used to collect the streamed items to print them.

```go
func main() {
	strm := stream.New([]rune("B世!ぽ@opqDQRS#$%^&*()ᅖ...O6PTnVWXѬYZbcef7ghijCklrAstvw"))

	strm.Filter(func(item rune) bool {
		return item >= 65 && item < (65+26)
	}).Map(func(item rune) string {
		return string(item) 
	}).Batch().Sort() 
	strm.Into(collectors.Writer(os.Stdout))

	if err := <-strm.Open(); err != nil {
		fmt.Println(err)
		return
	}
}
```

> See the full [source code](./examples/emitters/slice0).

Let's decompose the program to see how it works!

#### Create the stream

```go
strm := stream.New([]rune(`B世!ぽ@opqDQRS#$%^&*()ᅖ...O6PTnVWXѬYZbcef7ghijCklrAstvw`))
```

`stream.New` creates a new stream with a specified emitter that will emit elements on the stream.  This example uses an emitter that sources a slice of runes to emit random rune values on the stream.


#### Apply stream operations

```go
strm.Filter(func(item rune) bool {
    return item >= 65 && item < (65+26)
}).Map(func(item rune) string {
    return string(item)
}).Batch().Sort()
```

Next, operations are applied to the stream: 

* `Stream.Filter()` takes a function which is use to filter out stream elements.  Our example filters out non-capitalized latin alpha characters. 
* `Stream.Map()` maps each streamed rune to its correspoding to string value
* `Stream.Batch().Sort()` batches and sorts the streamed runes  


#### Collect the stream

```go
strm.Into(collectors.Writer(os.Stdout))
```

`Stream.Into` routes the stream to a collector. This example uses an io.Writer collector used to output the collected streamed items to `os.Stdout`.


#### Open the stream

```go
if err := <-strm.Open(); err != nil {
    fmt.Println(err)
    return
}  
```

`Stream.Open` opens the stream once it is composed (above).  This starts the emitter and executes the operations attached the stream.

### Example: streaming from `io.Reader`

The next example shows how to use Automi to stream data from an emitter that implements`io.Reader`.  While the example uses an in-memory source, this should work with any value that implements `io.Reader` including `os.File` for streaming file content and `net.Conn` for streaming content from connected sources.

> See the code comment for detail on how Automi is used.

```go
func main() {
	data := `"request", "/i/a", "00:11:51:AA", "accepted"
"response", "/i/a/", "00:11:51:AA", "served"
"response", "/i/a", "00:BB:22:DD", "served"...`

    // create io.Reader
	reader := strings.NewReader(data)
    
	// create stream with reader emitter,
	// buffers data as 50-byte chunks.
	stream := stream.New(emitters.Reader(reader).BufferSize(50))

    // map each 50-byte slice chunk to string
	stream.Map(func(chunk []byte) string {
		str := string(chunk)
		return str
	})

	// filter out string chunks with the word `request` in it
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

> See complete example [here](./examples/emitters/reader/emitreader.go).

### Example: streaming with CSV files
The following example streams the content of a CSV file, process each row as a streamed element, then write the result to another CSV file. 

**Note**: the example also showcases the use of custom types for stream values with type mapping and filtering operations.

```go
type scientist struct {
	FirstName string
	LastName  string
	Title     string
	BornYear  int
}

func main() {
    // creates a stream using a CSV emitter
    // emits each row as []string
    stream := stream.New("./data.txt")

    // Map each CSV row, []string, to type scientist
    stream.Map(func(cs []string) scientist {
        yr, _ := strconv.Atoi(cs[3])
        return scientist{
            FirstName: cs[1],
            LastName:  cs[0],
            Title:     cs[2],
            BornYear:  yr,
        }
    })

    // Filter out scientists born after 1930
    stream.Filter(func(cs scientist) bool {
        return (cs.BornYear > 1930)
    })

    // Map scientist value to []string
    stream.Map(func(cs scientist) []string {
        return []string{cs.FirstName, cs.LastName, cs.Title}
    })

    // Route each streamed []string element
    // to a CSV file collector
    stream.Into("./result.txt")

    // open the stream
    if err := <-stream.Open(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fmt.Println("wrote result to file result.txt")
}
```

## More Examples
[Examples](./examples) - View a long list of examples that cover all aspects of using Automi.

## Automi components
Automi comes with a set of built-in components to get you started with stream processing including the followings.

### Emitters

* `Channel`
* `CSV`
* `Reader`
* `Scanner`
* `Slice`

### Operators

* `Stream.Filter`
* `Stream.Map`
* `Stream.FlatMap`
* `Stream.Reduce`
* `Stream.GroupByKey`
* `Stream.GroupByName`
* `Stream.GroupByPos`
* `Stream.Sort`
* `Stream.SortByKey`
* `Stream.SortByName`
* `Stream.SortByPos`
* `Stream.SortWith`
* `Stream.Sum`
* `Stream.SumByKey`
* `Stream.SumByName`
* `Stream.SumByPos`
* `Stream.SumAllKeys`

### Collectors

* `CSV`
* `Func`
* `Null`
* `Slice`
* `Writer`

## Roadmap

* [] Alpha release (soon)
* [] Automatically turn on batch when batch operators are invoked
* [] Stream emitter to source from another stream
* [] Stream collector to sink to another stream
* [] New stream operators (join, split, broadcast, etc)
* [] Expose parallelization as operators
* [] Add type-specific operators to streams
* [] Performance-related tuning

## Licence
Apache 2.0
