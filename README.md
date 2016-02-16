Automi
======
Composable Stream Processing on top of Go Channels!

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi abstracts away (not too far away) the gnarly details of using Go channels to create pipelined and staged processes.  It exposes higher-level API to compose and integrate stream of data over Go channels for processing.  This is `still alpha work`. The API is still evolving and changing rapidly with each commit (beware).  Nevertheless, the core concepts are have been bolted onto the API.  The following example shows how Automi could be used to compose a multi-stage pipeline to process stream of data from a csv file.

#### Example
Automi is being developed as a pure API to create stream processors in Go.  The following code snippet shows how the Automi API could be used to apply multiple operators to data items as they are streamed.

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
 - `stram.To()` - writes sream values to a CsvSink
 - `stream.Open()` - opens and executes stream operator and wait for completion.

The code implements stream processing based on the pipeline patterns.  What is clearly absent, however, is the low level channel communication code to coordinate and synchronize goroutines.  The programmer is provided a clean surface to express business code without the noisy channel infrastructure code.  Underneath the cover however, Automi is using patterns similar to the pipeline patterns to create safe and concurrent structures to execute the processing of the data stream.

# Roadmap
The API is still taking shape into something that, hopefully will be enjoyable and practical code to create stream processors.  The project is a moving target right now, however the code will focus stabilizing the core API, additional operators, and sources/sinks connectors.  In the near future, there's plan to add functionalities to support execution windows to control stream growth and pressure on reductive operators.

#### Operators
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
