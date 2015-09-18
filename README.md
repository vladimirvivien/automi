Automi
======
Automi abstracts away (not too far away) the low-level manipulation of Go channels and exposes higher-level components to compose data and process flows.

[![GoDoc](https://godoc.org/github.com/vladimirvivien/automi?status.svg)](https://godoc.org/github.com/vladimirvivien/automi)

###Example
The following illustrates how Automi can be used to compose a data and process flow where the content of CSV file saved to a database and 

```
csv := &file.CsvRead{
    Name:"FileReader",
    FilePath : "./loca/file.txt",
    SeparaterChar:',',
    HasHeaderRow:true,
}
csv.Init()
csv.Exec()

db := &DbUpdate{
    Name:"dbWriter",
    Db:"<sql.DB value>"
    Input:csv.GetOutput(),
    Sql:"INSERT INTO tbl(name, age)VALUES(?,?)",
    Prepare: func(data interface{})[] interface{} {
	    result := make([]interface,2)
        values := data.([]string)
        result[0] = values[0]
        result[1], _ = strconv.Atoi(values[1])
        return vesult
    },
    Handle: func(data interface{}, result sql.Result) interface{} {
        if result.RowsAffected() != 1 {
            return api.ProcError {
                ProcName:"dbWriter",
                Err:fmt.Errorf("Failed to insert row: %s", data),
            }
        }
        return data
    },
}
db.Init()
db.Exec()
```
#Status
This is super alpha-level work.  The API is still taking shape into something that is enjoyable and easy to use.  The following components are available

 - **CsvRead**: Source component that emits csv records to output channel
 - **CsvWrite**: Sink component that writes channel items to a csv file
 - **HttpReq**: Allows Http requests/responses for each item sent to its input channel
 - **ItemProc**: A component for item processing with custom code
 - **ItemCollector**: Merges collected channel items from other components
 - **LogrusProc**: Logs messages with Logrus 
 - **DBUpdate**: A sink component that saves channel items to a database

###Coming Soon
 - DbQuery (Rdbms)

**System Integration**
 - File system source/sink
 - Ftp source/sink

**Distributed Data**
 - Kafka Consumer
 - Kafka Producer
 - Etcd component
 - Hdfs source/sink
 - Cassandra
 - Vitess
 - Gluster
 - Etc

**Routing/Topology**
 - Data Filters
 - Data Collectors
 - Broadcasters
 - Balancers
 - Etc
