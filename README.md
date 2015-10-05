Automi
======
Compose data streams into pipelined processes on top of Go channels!

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)
[![Build Status](https://travis-ci.org/vladimirvivien/automi.svg)](https://travis-ci.org/vladimirvivien/automi)

Automi abstracts away (not too far away) the gnarly details of using Go channels to create pipelined and staged processes.  It exposes higher-level modules to compose and integrate sstream of data over Go channels.  This is alpha work. The API is still taking shape (though more stable lately).  Nevertheless, the following example shows how create powerful assemblies of components to stream and process data from many sources (see component lists).

###Example
The following illustrates how Automi can be used to stream data from a csv file to a remote service with some data type transformation in between.

```Go
reader := &file.CsvRead{
    Name:"LogReader",
    FilePath : "./loca/file.txt",
    SeparaterChar:',',
    HasHeaderRow:true,
    Function:func(ctx context.Context, item interface{}) interface{}{
		row, _ := item.([]string)
        return LogData{ID:row[0], Path:row[1], Asset:row[2]}
	},
}

procItem := &proc.Item {
    Name:"procItem",
	Concurrency: 4,
    Function: func(ctx context.Context, item interface{}) interface{}{
	    data, _ := item.(LogData)
		if data.ID == "" {
		    return fmt.Errorf("Missing ID in log")
		}
        if err := somePersistance.Save(data); err != nil{
			return err
		}
		return data
	},
}

remote,_ := someRemoteSvc.New("http://svc.internal:8544")

sendRemote := &proc.Endpoint{
    Name: "sendRemote",
    Function: func(ctx context.Context, item interface{}) error {
		data, _ := item.(LogData)
		err := remote.Submit(data)
		if err != nil {
			return err
		}
		return nil
	},
}

// build process graph
p := plan.Plan(Config{})
p.Flow(plan.From(reader).To(procItem))
p.Flow(plan.From(procItem).To(sendRemote))

// lauch/wait for plan completion
<-p.Exec()
```
This simple example replaces hundreds of lines of code.  What you don't see is the intricate details required to pipeline streams of data into multiple synchronized stages.  What you do get, however, is an easy API to create complex cohegraphies to process data from any source.

#Features
The API is still taking shape into something that is enjoyable and easy to use.  The following components are available:

 - **CsvRead**: Source component that emits records from a csv file
 - **CsvWrite**: Sink component that writes channel items to a csv file
 - **HttpReq**: Allows Http requests/responses for each item sent to its input channel
 - **Item**: A component for item processing items and push result downstream
 - **Endpoint**: A generic sink component useful to mark the end of a flow
 - **ItemCollector**: Merges collected channel items from other components
 - **DBUpdate**: A sink component that saves channel items to a database

###Other Features
 - Ability to build complex process graph
 - Uses Go channels for synchronization
 - Available auxiliary channel for non-data events
 - Logging support with Logrus
 - More features to come

###Coming Later (Eventually)

**Timout and Cancellation**
 - Support for cancellation of executing flow
 - Ability to specify per-process timeouts
 
**Metrics**
 - Comprehensive metrics capture/report
   
**System Integration**
 - File system source/sink
 - Ftp source/sink

**More Data Itegration**
 - Kafka Consumer
 - Kafka Producer
 - Etcd component
 - Hdfs source/sink
 - Cassandra
 - Gluster
 - Etc

**Routing/Topology**
 - Data Filters
 - Broadcasters
 - Balancers
 - Etc

# Motivation
This work was highly-motivated by Sameer Ajmani's blog entry on Pipelining data over Go channels https://blog.golang.org/pipelines.
