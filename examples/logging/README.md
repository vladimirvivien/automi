Automi Stream Logging
=====================

Automi allows you to inject log messages into the stream at runtime. This
example shows how to send a log message using the stream context and how to
define a log handler to handle logs as shown below:

```go
	stream.WithLogFunc(func(msg interface{}) {
		log.Printf("INFO: %v", msg)
    })

    stream.Filter(func(ctx context.Context, walkResult [2]interface{}) bool {
		err, ok := walkResult[1].(error)
		if ok && err != nil {
			autoctx.Log(ctx, fmt.Sprintf("failed to walk %s: %s", walkResult[0], err))
			return false
		}
		return true
	})
```
