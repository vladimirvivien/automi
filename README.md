Automi
======
Composable Stream Processing over Go Channels!

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi abstracts away (not too far away) the gnarly details of using Go channels to create pipelined and staged processes.  It exposes higher-level API to compose and integrate stream of data over Go channels for processing.  This is `still alpha work`. The API is still evolving and changing rapidly with each commit (beware).  Nevertheless, the core concepts are have been bolted onto the API.  The following example shows how Automi could be used to compose a multi-stage pipeline to process stream of data from a csv file.

### Example
The following abbreviated code snippet shows how the Automi API could be used to apply multiple operators to data items as they are streamed. [See full example here.](example https://github.com/vladimirvivien/automi/blob/master/examples/ex0/procdata.go)

```Go
type scientist struct {
	FirstName string
	LastName  string
	Title     string
	BornYear  int
}
// Example of stream processing with multi operators applied
// src - CsvSource to load data file, emits each record as []string
// out - CsvSink to write result file, expects each entry as []string
func main() {
	in := src.New().WithFile("./data.txt")
	out := snk.New().WithFile("./result.txt")

	stream := stream.New().From(in)
	stream.Map(func(cs []string) scientist {
		yr, _ := strconv.Atoi(cs[3])
		return scientist{
			FirstName: cs[1],
			LastName:  cs[0],
			Title:     cs[2],
			BornYear:  yr,
		}
	})
	stream.Filter(func(cs scientist) bool {
		if cs.BornYear > 1930 {
			return true
		}
		return false
	})
	stream.Map(func(cs scientist) []string {
		return []string{cs.FirstName, cs.LastName, cs.Title}
	})
	stream.To(out)

	<-stream.Open() // wait for completion
}
```
The previous code sample creates a new stream to process data ingested from a csv file using several.  In the code, each method call on the stream (`From()`, `Filter()`, `Map()`,, and `To()`) represents a stage in the pipeline as illustrated in the following:

	From(source) -> Map(item) -> Filter(item) -> Map(item) -> To(sink)

The stream operators that are applied to the stream as follows:

 - `stream.From()` - reads stream from CsvSource
 - `stream.Map()` - maps []string to to scientist type
 - `stream.Filter()` - filters out scientist.BornYear > 1938
 - `stream.Map()` - maps scientist value to []string
 - `stream.To()` - writes stream values to a CsvSink
 - `stream.Open()` - opens and executes stream operator and wait for completion.

The code implements stream processing based on the pipeline patterns.  What is clearly absent, however, is the low level channel communication code to coordinate and synchronize goroutines.  The programmer is provided a clean surface to express business code without the noisy channel infrastructure code.  Underneath the cover however, Automi is using patterns similar to the pipeline patterns to create safe and concurrent structures to execute the processing of the data stream.

# Roadmap
The API is still taking shape into something that, hopefully will be enjoyable and practical code to create stream processors.  The project is a moving target right now, however the code will focus stabilizing the core API, additional operators, and sources/sinks connectors.  In the near future, there's plan to add functionalities to support execution windows to control stream growth and pressure on reductive operators.

### Operators
The focus of the project will be to continue to implement pre-built operators
to help stream processor creators. 
 - **Transformation** Filter, Maps, Join
 - **Accumulation** Reduce, Aggregation, Grouping
 - **Etc** 

### Features
Automi will strive to implement more features to make it more useful as the
project matures including some of the followings.
 - **Continuous streams**  Better control of non-ending streams
 - **Concurrency** Control of concurrently running operators
 - **Timout and Cancellation Policies** Establish constraints for running
   processes
 - **Metrics** Expose metrics of running processes

### Core Sources/Sinks
The followings are potential sources and sink components that will be part of
the core API.
 - **io.Reader** Stream from io.Reader sources
 - **io.Writer** Write stream data to io.Writer sinks
 - **Csv**: Stream from/to value-separated files
 - **Socket** Stream from/to network sockets
 - **Http**: Source from http end-point, stream from/to http sinks
 - **Database**: Source from DB tables, stream to DB sinks
 
### External Source/Sink Ideas
The following is a list of sources and sinks ideas that should be implemented as
externals projects to avoid unwanted dependencies.
 - **Messaging** Kafka, NATs, Etc
 - **Distributed FS** HDFS, S3, etc
 - **Distributed DB** Cassandra, Mongo, etc
 - **Logging** Flume, statsd, syslog, etc
 - Whatever source/sink users find useful
 - Etc
