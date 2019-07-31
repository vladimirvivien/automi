package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/vladimirvivien/automi/collectors"
	"github.com/vladimirvivien/automi/stream"
)

// Uses filepath.Walk to emit visited paths from root
// Both visited paths and produced error are returned
// as a tuple of type [2]interface{}{path,error}
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

	// filter out errored walk results
	stream.Filter(func(walkResult [2]interface{}) bool {
		err, ok := walkResult[1].(error)
		if ok && err != nil {
			return false
		}
		return true
	})

	// optionally, map tuple [2]interface{}{path,err} -> string
	stream.Map(func(walkResult [2]interface{}) string {
		val, ok := walkResult[0].(string)
		if !ok {
			return ""
		}
		return val
	})

	stream.Map(func(filePath string) [3]interface{} {
		data, err := ioutil.ReadFile(filePath)
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
