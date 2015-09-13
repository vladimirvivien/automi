Automi
======
Automi abstracts away (not too far away) the low-level manipulation of Go channels and exposes higher-level components to compose data and process flows.
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
    Dsn:"<connection string>"
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

 - **CsvRead**: emits csv records for downstream processing
 - **HttpReq**: Generates Http requests for each item received

###Coming Soon
 - CsvWrite
 - DbUpdate 
 - DbQuery
 - Kafka Consumer
 - Kafka Producer
 - Etcd component
 - Hdfs compoents
 - Data Filters
 - Data Collectors
 - Routers and Balancers
 - Etc
