package main

import (
	"encoding/base64"
	"log"
	"net/http"

	"github.com/vladimirvivien/automi/stream"
)

// HTTP example that returns a base64 encoding of http Body.
// Start server: go run httpsvr.go
// Test with: curl -d "Hello World"  http://127.0.0.1:4040/
func main() {

	http.HandleFunc(
		"/",
		func(resp http.ResponseWriter, req *http.Request) {
			resp.Header().Add("Content-Type", "text/html")
			resp.WriteHeader(http.StatusOK)

			strm := stream.New(req.Body)
			strm.Process(func(data []byte) string {
				return base64.StdEncoding.EncodeToString(data)
			}).Into(resp)

			if err := <-strm.Open(); err != nil {
				resp.WriteHeader(http.StatusInternalServerError)
				log.Printf("Stream error: %s", err)
			}
		},
	)

	log.Println("Server listening on :4040")
	http.ListenAndServe(":4040", nil)
}
