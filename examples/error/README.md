Automi Error Handling
=====================

This example shows how to setup Automi to handle stream errors by defining a
function that will be used to handle stream errors as they are generated as shown
below.

```go
	stream.WithErrorFunc(func(err api.StreamError) {
		log.Printf("ERROR: %v", err)
    })
```

The function receives a value of type `api.StreamError`, which implements `error`,
every time a stream error is injected in the stream.

 