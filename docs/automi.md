Automi
------

Automi is an API to process stream of data over Go channels. Each stream is made of a single source of data, operators that process the data, and an a sink which is the endpoint of the stream.

## The Stream
The stream is an abstraction that represents data flowing from an emitter to a collector.  Internally, Automi uses Go channels as a conduit to pipe the data from emitter to collector.  A `Stream` is created with:

```go
  strm := stream.New(<source>)
```
### Stream Source
The `<source>`  is an emitter (see *Emitter* later) that supplies the data that flows through the stream.  A `Stream` can only have one source, which is specified during the stream creation, as the following types (anything else causes a runtime error):

- `[]<T>` : a slice of type T
- `io.Reader` : a value that implements `io.Reader`
- `types.Emitter` : a value that implements `types.Emitter`
- `<-chan T` : a send channel of type `T`

For instance, the following creates a stream that sources its data from a slice of runes:
```go
   strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
```
### Stream Operators
Once the stream is created, operators (in the form of functions) can be added to the stream to process each item as it flows through.  For instance, the following adds a Filter operator that uses a function that only allow strings with value `"car"` to flow through the stream. 

```go
    stream.Filter( func (cs string) bool {
        return (cs == "car")
    })
```
See *Operator* later in this document.

### Stream Sink
After processing, the streamed data can be directed to an endpoint known as a `sink`.  The sink is a terminal `Collector` (see *Collector* later) that collects the data but offers no further path for data flow.

A `Stream` value can only have one sink which may be of the following types (any other types will cause an error):
- `[]<T>` : a slice of type `T`
- `io.Writer` : a value that implements `io.Writer`
- `types.Collector` : a value that implements `types.Collector`
- `chan<- <T>` : a receive channel of type `T`

A sink is added to the stream with the `SinkTo(<sink>)` method.  For instance, the following stream sinks its data into an `io.Writer`:
```go
    snk := new(bytes.Buffer)
    strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
    err := <- strm.SinkIn(snk)
```
The following shows a fully realized stream with a source, operator, and sink defined:

```go
    snk := new(bytes.Buffer)
    strm := automi.New([]rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
    strm.Filter( func (chr rune) bool{
        return chr >= 77
    })
    <- strm.SinkIn(snk)
```

## Emitter
An Automi Emitter is able to emit data into the stream.  An emitter does not have or expose any stream processing methods.  It simply pumps its data into the stream.  The stream relies on the defined operators for downstream processing.  Automi provides several emitters built-in emitters for slice, IO readers, and send-only channels.  Developers are encouraged to create emitters that better address their needs by implementing `types.Emitter`.

Emitters can be categorized as:

- Batch
- Streaming

### Batch Emitter
A batch emitter is expected to have a finite window of emission where it starts streaming data at time *t* and stops at time *t + d*.  A batch emitter is sourced by a finite resource like a file, a memory structure, etc. For instance, a `CSVSource` is sourced by a file that will stop emitting once all records are read from the file. 

### Stream Emitter
A stream emitter is open-ended and theoretically may never stop emitting.  A stream emitter may support emission windows that can specify a start and an end to the emission.  A stream emitter, for instance, may be sourced by a network resource such as socket, a message bus, etc.

## Collector
An Automi Collector is a node that collects the streamed data (partial or all).  A collector basically de-stream the data so it can be processed as a batch.  Depending on implementation, the collector can store collected data in-memory, disk, database, etc.  Collectors can be terminal where they do not expose a stream for downstream consumption. Or a collector can be provide downstream path allowing the data to be consumed downstream.

### Terminal Collectors
The terminal collector collects the streamed data (all or partial) and do not provide a way for the data to flow out to a downstream node.  For instance, a terminal collector may store its collected data to a file.  Once the data is written to the file, no further processing is offered by the collector.  The stream terminates.

Automi includes the following built-in terminal collectors:

- `CsvSink` : a collector that writes its collected data to a CSV file.

### Group Collector
The `Group` collector is non-terminal as it its methods can offer an escape valves (return a stream) allowing collected/processed data to continue downstream.  The Group is created on a stream using method Group as shown below:
```
strm := stream.New([]string{"hello", "goodbye"})
strm.Group(stream.Continuously).ByPos(0)
<- strm.SinkTo(os.Stdout)
```
#### Grouping Window
A group window specifies how large the group items would be.

- stream.All
- stream.ForTime(`<time.Duration>`)
- stream.ForCount(`<number of items>`)
- stream.WithFunc(`<func()bool>`)


#### Group Methods
The Group offers several methods that allow batch processing of the collected data.

  * `By(<func>)` - groups the collected items using the specified function, returns a `Stream`
  * `ByPos()`  - groups collected items by their position in a slice, returns a `Steram`
  * `ByValue()` - groups collected items by their values in a slice, returns a `Stream`
  * `ByFieldName()` - groups collected items by field names of structs in a slice, returns a `Stream`
  * `SumInts()` - sums the collected integers from slice of integers, returns a `Stream`  of integer slices
  * `SumFloats()` - sums the collected floats from slice of floats
  * `Sort(<func>)` - sorts all collected items using the provided comparator function, returns a `Stream`
  * `ForAll(<func>)` - apply the provided function to all collected items, returns a `Stream`

For instance, the following uses the `Group` collector to group its content by struct field name `id`.  The collector stream window of 1 second where after every second the collected data is grouped and pushed downstream.

```go
src := []struct{id string; val int}{
  {"A", 0},   {"B", 2},    {"C", 4},
  {"A", 10},  {"B", 20},   {"C", 40},
  {"D", 100}, {"E", 200},  {"C", 400},
}  
strm := stream.New(src)
strm.Group(stream.Every(1 * time.Second)).ByFieldName("id")
<-strm.SinkTo(os.Stdout)
```

## Operations
Operations are methods attached to the stream directly.  They provide a means to express the processing intent on the data as it streams from source to sink.

### Stream.Process()
This is a generic operator that can 
- Filter
- Map
- FlatMap
- Reduce
