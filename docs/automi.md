Automi
------

Automi is a streaming API for the Go programming language.  An Automi stream uses Go channels as a conduit to create a data flow using a pipeline metaphor.  The API uses Go semantics and constructs to compose powerful multi-stage streams for processing data.

A stream is composed of several nodes that are arranged to emit, process, and collect streamed data.  A stream starts with an `emitter` which is a node capable of emitting (producing) data items for the stream.  An Automi stream can have one or more `operator` nodes that can apply a function to the to the stream items as they flow.  Finally, the stream can have `collector`  nodes which is a component that accumulates stream items.

## Emitters
An emitter is an Automi component that is capable of emitting data items unto a stream.  Emitters are represented by interface:
```go
type Emitter interface {
    GetOutput() <-chan interface{}
}
```
The emitter exposes a `receive-only channel` where its data is pumped onto the stream.  Any component can act as an emitter as long as it exposes a way to push data on the stream.

### Batch vs Stream Emitters
A `batch emitter` is one that is expected to have a finite window of emission where it starts streaming data at time *t* and eventually stops at time *t + d*.  A batch emitter is usually sourced by a resource with finite data items like a file or an in-memory structure. 

A `stream emitter` is open-ended and theoretically may never stop emitting.  A stream emitter component may support emission windows or be forced to be batched.  A stream emitter, for instance, may be sourced by a network resource such as socket, a message bus, etc.

### Built-in Emitters
Automi comes with many emitters that are part of the API and ready to use.

- `emitters.Slice` - a batch emitter that emits items from a slice
- `emitters.Reader` -  a stream emitter that emits from an `io.Reader` 
- `emitters.Chan` - a stream emitter that emits from a `receive-only channel`
- `emitters.CSV` - a batch emitter that emits data from CSV file

## Collectors
An Automi `collector` is a component design to collect data items from a stream.  Collectors are represented with the following interface.
```go
type Collector interface {
    SetInput(<-chan interface{})
}
```
A collector takes a `receive-only channel` from which it can collect streamed data.  Any component that implements method `SetInput` can act as a collector.  

### Terminal Collectors
Depending on implementation, a collector can be terminal where it provides no path for the data to continue for downstream processing.  A terminal collector node can collect and store collected items in-memory, in a file, in a database, etc.  When a terminal collector is added to a stream, it will mark the end of the stream.

### Built-in Collectors
Automi comes with many built-in collectors that are part of the API including:

- `collectors.Slice` - a terminal collector that stores collected items in a slice
- `collectors.Writer` - a terminal collector that writes collected items to an `io.Writer`
- `collectors.Chan` - a terminal collector that writes collected items to a `send-only channel`
- `collectors.CSV` - a terminal collector that stores collected item to a CSV file

## Operators
An operator is a node that applies a function to items that are flowing though a stream.  The functions applied to the stream may be user-provided or opaque at runtime.  Operators implement both `Collector` and `Emitter` interfaces allowing them to receive data as input and produce output items respectively.

### Built-in Operators
Automi offers unary operators that are designed to apply transformative functions to streamed data.  For instance, a filter operator may be applied to remove certain items from the stream using a user-provided function.  Automi also provides binary operators primarily designed for accumulative operations.  Both types of operators are covered later in the Stream section of this documentation.

## The Stream
An Automi stream is an abstraction that represents data flowing from an emitter to a collector.  Internally, Automi uses Go channels as a conduit to pipe the data from emitter to collector.  A `Stream` is created with:

```go
  strm := stream.New(<source>)
```
Where `<source>` represents an `emitter` that can emit data onto the stream as discussed in the next section.

### Stream Source
The `<source>`  is an emitter (see *Emitter* covered earlier) that supplies the data that flows through the stream.  A `Stream` can only have one source, which is specified during the stream creation, as shown in the following:
```go
   strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
```
The previous creates a stream and uses a slice of `rune` as its source.  The `stream.New(src)` function can use parameter `src` of different types including the followings:

- `[]T` - a slice of  type `T` used as data emitter
- `io.Reader` - an io.Reader that can emit data 
- `<-chan T` - a receive-only channel of type `T` that can emit data
- `api.Emitter` - a value that implements `api.Emitter`

Internally, `stream.New` will convert the value passed in to the appropriate `Emitter` type.

### Stream Operators
Once the stream is created, operator nodes are added to the stream to process each item as it flows through.  For instance, the following adds a `Filter` operator that uses a user-provided filter function that only allow strings with value `"car"` to flow through the stream. 

```go
    strm := stream.New([]string{"auto", "car", "plane", "boat"})
    strm.Filter( func (cs string) bool {
        return (cs == "car")
    })
```
### Built-in Stream Operators
Automi comes with several built-in operators as outline below:

- `stream.Process(func(T) R)` - applies `func(T) R` to each streamed item where `T` is the type of the incoming item and `R` is the type of the result item to be sent downstream.
- `stream.Filter(func(T) bool)` - applies `func(T) bool` to each incoming item where `T` is the type of the incoming data and returns a `bool` when true allows the data to continue downstream.
- `stream.Map(func(T) R)` - applies `func(T) R` which takes stream item of type `T` and generates a new value of type `R`.
- `stream.FlatMap(func(T) []R)` - applies `func(T)[]R` where `T` is the type of an incoming streamed item and the function is expected to return `[]R` which is a slice of values to be consumed downstream.
- `stream.Reduce(S, func(T0, T1) R)` - uses initial seed value `S` that is applied to an accumulative function `func(T0, T1) R` which takes partial result `T0` and streamed item `T1` to produce new result `R`.
- `stream.Batch()` - is an operator that collects incoming data into batches of N size.  The batched items are pushed downstream as a slice `[]T`.
- `stream.ReStream` - is an operator that takes incoming items of composite types (`[]T` and `map[K]V`) and decompose and stream stream each item individually.


### Stream Sink
An Automi stream terminates at an endpoint called as a `sink` specified as stream method `SinkTo` as shown below.
```go
    snk := new(bytes.Buffer)
    strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
    err := <- strm.SinkTo(snk)
```
The sink is a terminal `Collector` that collects data from the stream for off-stream storage or processing (see *Collector* earlier).  Each stream can only have one sink which may be one of the following types:
- `[]<T>` : a slice of type `T`
- `io.Writer` : a type implementing `io.Writer`
- `chan<- <T>` : a send-only channel of type `T`
- `api.Collector` : a value that implements `types.Collector`

The following shows a fully realized stream with a source, operator, and sink defined:

```go
    strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
    strm.Filter( func (chr rune) bool{
        return chr >= 77
    })
    <- strm.SinkTo(new(bytes.Buffer))
```
# Batches
Automi supports the notion of batches (or windows) that can be applied to a stream. A batch is
an operator that collects items from a stream.  Then the batch releases the collected items, as 
a group, based on trigger logic associated with the batch.  Batches allow operations that can only
be applied on collections to take place (see batch operations below).  For intance, once a batch is
realized, a `Sort` operation can be applied to the batched items.


## Operating on Batches
Automi comes with several operators that are intended to process streamed items that have been batched (using the `Batch` operator) or are of slice type `[]T` where `T` is type of elements in the slice.  Automi comes with several batch operators that apply grouping, sorting, or a reductive operators on the streamed items.

### Built-in Batch Operators
- `GroupByKey` - groups incoming items of type `[]map[K]V` by their K values.
- `GroupByName` - groups items of type `[]struct{N}` by the value of field N 
- `GroupByPos` - groups items of type `[]T` or `[][]T` by the slice index
- SortByKey - sorts items of type `[]map[K]V` by their K values 
- SortByName - sort items of type `[]struct{N}` by field N
- SortByPos - sort items of type `[]T` or `[][]T` by the slice index
- `SortWith` - sort items of type `[]T` using a Less function `f(i, j int) bool`
- `Sum` - sums items of type `[]T` or `[][]T` where T is a valid integer or floating point
- `SumByKey` sums items of type `[]map[K]V` where K returns integer or floating point value
- `SumByName`- sums items of type `[]struct{N}` where field `N` returns an integer or floating point value
- `SumByPos` - sums items of type `[]T` or `[][]T` where specified index returns a numeric value

The following shows an example of how to group 
```go
src := []struct{id string; val int}{
  {"A", 0},   {"B", 2},    {"C", 4},
  {"A", 10},  {"B", 20},   {"C", 40},
  {"D", 100}, {"E", 200},  {"C", 400},
}  
strm := stream.New(src)
strm.Batch().GroupByName("id")
<-strm.SinkTo(os.Stdout)
```
### Windowed Batches (Future)
Automi makes it possible to create emission windows for batch emitter nodes.  A window provides policy on how large a batch should get.  Automi provides several pre-built batch strategies:
- `batch.All` - batches all items until the upstream emitter stops
- `batch.ByDuration(time.Duration)` - indicate a batch based on time duration
- `batch.ByTime(time.Time)` - indicate a batch for up to a speficied clock time
- `batch.BySize(int)` - batches item until N count of items are batched
- `batch.ByFunc(func(interface{}) bool)` - uses specified function to trigger batch 

The Stream method `Batch(<strategy>)` can accept one of the strategy above to internally select the proper batch operator to place on the stream.
 ```go
src := []struct{id string; val int}{
  {"A", 0},   {"B", 2},    {"C", 4},
  {"A", 10},  {"B", 20},   {"C", 40},
  {"D", 100}, {"E", 200},  {"C", 400},
}  
strm := stream.New(src)
strm.Batch(batch.BySize(2)).GroupByName("id")
<-strm.SinkTo(os.Stdout)
```
Functions `batch.ByXXX` are simply idiomatic devices which return the expected type.  Method Batch(interface{}) can take the expected values directly.  So the above can be rewritten as:
```go
src := []struct{id string; val int}{
  {"A", 0},   {"B", 2},    {"C", 4}, ...
}  
strm := stream.New(src)
strm.Batch(2).GroupByName("id")
<-strm.SinkTo(os.Stdout)
```
