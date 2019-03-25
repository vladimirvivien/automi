package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vladimirvivien/automi/api"

	autoctx "github.com/vladimirvivien/automi/api/context"
	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

// Demostrates stream runtime logging. Uses the examples/md5 as basis.
// To run: go run md5.go -p ./..
func emitPathsFor(root string) <-chan [2]interface{} {
	paths := make(chan [2]interface{})
	go func() {
		defer close(paths)
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if !info.Mode().IsRegular() {
				return nil
			}
			paths <- [2]interface{}{path, err}
			return nil
		})
	}()
	return paths
}

func main() {
	var rootPath string
	flag.StringVar(&rootPath, "p", "./", "Root path to start scanning")
	flag.Parse()

	stream := stream.New(emitPathsFor(rootPath))

	stream.WithLogFunc(func(msg interface{}) {
		log.Printf("INFO: %v", msg)
	})

	stream.WithErrorFunc(func(err api.StreamError) {
		log.Printf("ERROR: %v", err)
	})

	// filter out errored walk results
	stream.Process(func(ctx context.Context, walkResult [2]interface{}) interface{} {
		err, ok := walkResult[1].(error)
		if ok && err != nil {
			api.Error(fmt.Sprintf("failed to walk %s: %s", walkResult[0], err))
			return false
		}

		return walkResult
	})

	// optionally, map tuple [2]interface{}{path,err} -> string
	stream.Map(func(ctx context.Context, walkResult [2]interface{}) interface{} {
		val, ok := walkResult[0].(string)
		if !ok {
			return api.Error("Unsupported data type in map")
		}
		if strings.HasPrefix(filepath.Base(val), "a") {
			return api.Error("encountered file that starts with a")
		}
		autoctx.Log(ctx, fmt.Sprintf("calculating md5 for %s", val))
		return val
	})

	stream.Map(func(ctx context.Context, filePath string) [3]interface{} {
		data, err := ioutil.ReadFile(filePath)
		sum := md5.Sum(data)
		autoctx.Log(ctx, fmt.Sprintf("calculated md5 %X for path %s", sum, filePath))
		return [3]interface{}{filePath, md5.Sum(data), err}
	})

	// sink the result
	stream.Into(collectors.Func(func(items interface{}) error {
		sums := items.([3]interface{})
		file := sums[0].(string)
		md5Sum := sums[1].([md5.Size]byte)
		fmt.Printf("file %-64s md5 (%-16x)\n", file, md5Sum)
		return nil
	}))

	// open the stream
	if err := <-stream.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
