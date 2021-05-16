# TODO

## 0.2.0 - Tasks
* [x] Add go module
* [x] Use GitHub Actions CI
* [ ] Submit to Awesome-Go  
* [x] Add CI/CD badges
  * [ ] Build passing, Go Reference, Go Report 
  * [ ] Awesome-Go
* [ ] Add copyright to source files
* [ ] Remove dead code  
* [ ] Signal stream cancellation from error handler ErrorFunc
* [ ] Add Stream.OnOpen(func(ctx)) hook to trigger before stream starts
* [ ] Add Stream.Notify() to catch os/signal notif
  
## 0.2.5
* [ ] Automatically turn on batch when batch operators are invoked
* [ ] Add support for stream.Into(func(item <type>){}) which is equivalent to stream.Into(collectors.Func(func))
* [ ] Change collectors signature to return StreamError instead of error

## 0.3 - Tasks
* [ ] Stream emitter - source from another stream
* [ ] Stream collector - sink into another stream
* [ ] New stream operators (join, split, broadcast, etc)
* [ ] Parallelization operator
* [ ] Add type-specific operators to streams (i.e. `SortInt()`)

## 0.1 Alpha - Tasks
* [x] Refactor log - implement logging with handler func
* [x] Error propagation and handling strategies
* [x] Support ability to pass context to user-defined operation functions
* [x] Dependency mgmt
* [x] Fix tests with `-race` failure
* [x] net/htt examples
* [x] Examples with error handling
* [x] Examples with logging
* [x] Context cancellation check for all operators
* [x] Examples with gRPC streaming
* [x] Add logo to README.md
* [x] Rebase commit history