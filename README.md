Automi
======
Composable Stream Processing on top of Go Channels!

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi abstracts away (not too far away) the gnarly details of using Go channels to create pipelined and staged processes.  It exposes higher-level API to compose and integrate stream of data over Go channels for processing.  This is `still alpha work`. The API is still evolving and changing rapidly with each commit (beware).  Nevertheless, the core concepts are have been bolted onto the API.  The following example shows how Automi could be used to compose a multi-stage pipeline to process stream of data from a csv file.

#### Example
Automi, at the moment, is being developed as a pure API to create stream processors in Go.  The following code snippet shows how the Automi API could be used to stream and process the content of a file using multiple stages.

```Go
svc := someDataService.Create(context.Background())  // illustration, stand-in for some service

strm := stream.New()

// set stream source as csv file, emits []string
strm.From(file.CsvSource("./local/in-forms.csv"))

// Only allows record where col 0 starts with "CLEAR_"
strm.Filter(func(item interface{}) bool{
    row := item.([]string)
    return strings.HasPrefix(row[0], "CLEAR_")
})

// maps stream item from []string to struct Form
strm.Map(func(item interface{}) interface{} {
    row := item.([]string)
    return Form{Status:row[0], Id:row[1], Data:row[5]}
})

// Func to invoke some service call on data item
// Emits a []string for downstream
strm.Do(func(ctx context.Context, item interface{}) interface{} {
    form := item.(Form)
    resp, err := svc.Validate(form)
    if err != nil {
        return nil 
    }
    return []string{resp.Id, resp.Code, resp.Content}
})

// Terminal step, sinks data into a csv flat file
strm.To(db.CsvSink("./local/resp-forms.txt"))

// open stream and wait for execution
err := <-strm.Open()
if err != nil {
    fmt.Println("Processing failed!")
}
```
The previous code sample creates a new stream to process data ingested from a csv file using several steps (see code comment).  In the code, each method call on the stream (`From()`, `Filter()`, `Map()`, `Do()`, and `To()`) represents a stage in the pipeline as illustrated in the following.  

	From(source) -> Filter(item) -> Map(item) -> Do(item) -> To(sink)

The `From()` method, for instance, starts the stream by ingesting the content of a csv file and emits a `[]string` for each row.  `Filter()` does what you would expect, it filters out csv rows from the stream based on record content.  `Map()` takes the `[]string` from the previous stage and emits struct `Form{}` for downstream consumption.  The `Do()` function provides a place for arbitrary logic to be applied to the stream.  It makes a call to a service (here for illustrative purpose), then returns [] for the next processing element.  Lastly, the stream is terminated with csv sink (with the `To()` function) that writes the result to a file.

The code implements stream processing based on the pipeline patterns.  What is clearly absent, however, is the low level channel communication code to coordinate and synchronize goroutines.  The programmer is provided a clean surface to express business code without the noisy infrastructure code.  Underneath the cover however, Automi is using patterns similar to the pipeline patterns discussed earlier to create safe and concurrent structures to execute the processing of the data stream.

# What it wants to be when it grows up
The API is still taking shape into something that, hopefully will be enjoyable and practical code to create stream processors.  The project is a moving target right now, but hopefully some of the following features will find their way into the code base.

#### Functions
 - **Transformation** Filter, Maps, Join
 - **Accumulation** Reduce, Aggregation, Grouping, etc
 - **Action** Count, min/max,  
 - **Etc** 

#### Core Sources/Sinks
 - **SocketSource** Network socket stream source
 - **SocketSink** Network socket stream sink
 - **CsvSource**: Source for CSV files
 - **CsvSink**: Sink component 
 - **HttpSource**: Sources stream data from http
 - **HttpSink**: Sink for posting data via http
 - **DbSource**: Database source for streaming data items
 - **DbSink**: A sink component for streaming data
 
#### More Source/Sink Ideas
 - Kafka Source
 - Kafka Sink
 - Hdfs source/sink
 - Cassandra
 - Sources/sinks for messaging systems
 - Sources/sinks for logging systems
 - Whatever source/sink users find useful
 - Etc

#### Other ideas/features being considered
 - **Functions/support for continuous streams**
 - **Parallelism and Concurrency support**
 - **Timout and Cancellation Policies**
 - **Metrics**
 - **Streaming service** 

