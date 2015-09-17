package db

// DbUpdate is a processor that can submit SQL updates for items in its input.
// The compoenent allows for preparation of statement and handling of result.
type DbUpdate type {
    Name string
    Input <-chan interface{}
    DB *sql.DB
    SQL string
    
    logs <-chan interface{}
    output <-chan interface{}
}

func (t *DbUpdate) Init() error {
    if t.Name ==""{
        return api.ProcError{Err:fmt.Errorf("Missing Name attribute")}
    }
    if t.Input == nil {
        return api.ProcError{
            ProcName : t.Name,
            Err : fmt.Errorf("Missing Input attribute"),
        }
    }
    if t.DB == nil {
        return api.ProcError{
            ProcName : t.Name,
            Err : fmt.Errorf("Missing DB attribute"),
        }
    }
    if t.SQL == "" {
        return api.ProcError{
            ProcName : t.Name,
            Err : fmt.Errorf("Missing Input attribute"),
        }
    }

}
        
    