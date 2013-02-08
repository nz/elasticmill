package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const batchSize int = 1000                               // maximum documents per batch
const batchPause time.Duration = time.Millisecond * 1000 // milliseconds between batch updates

var serverPort = os.Getenv("PORT")
var updates chan string = make(chan string, batchSize)

func main() {
	go processor()
	server()
}

// Processor receives batch updates from the updates channel and sends them to Elasticsearch,
// subject to a short cooldown interval. The strings in the updates channel should be formatted
// for the bulk API
func processor() {

	// Create an array of strings to help prepare our bulk request,
	// set its length to zero.
	batch := make([]string, batchSize)
	batch = batch[0:0]

	// A short interval to wait between requests.
	limiter := time.Tick(batchPause)

	// Infinite loop to pull updates out of a channel, and periodically send them to Elasticsearch
	for {
		select {

		// Pull as many updates as we can from a channel, up to its maximum length.
		case update := <-updates:
			batch = append(batch, update)

		// When the channel is empty, combine and send the batch in a Bulk API request
		default:
			if len(batch) > 0 {
				log.Println("Processed", len(batch), "updates")
				// TODO: send the batch to Elasticsearch
				// Reset the batch array
				batch = batch[0:0]
			}
			<-limiter
		}
	}
}

// Set up and run an HTTP server that intercepts updates and formats them for batches,
// and also proxies searches through to Elasticsearch.
func server() {

	r := mux.NewRouter()
	s := r.Methods("PUT", "POST").Subrouter()

	// Match _bulk handlers
	s.Path("/_bulk").
		HandlerFunc(queueClusterBulk)
	s.Path("/{index}/_bulk").
		HandlerFunc(queueIndexBulk)
	s.Path("/{index}/{type}/_bulk").
		HandlerFunc(queueTypeBulk)

	// Match individual document updates
	r.Methods("POST", "PUT").
		Path("/{index}/{type}/{id}").
		HandlerFunc(queueUpdates)

	r.Methods("DELETE").
		Path("/{index}/{type}/{id}").
		HandlerFunc(queueDeletes)

	// Proxy read requests
	r.PathPrefix("/").
		HandlerFunc(proxy)

	http.Handle("/", r)

	fmt.Println("listening on " + serverPort + "...")
	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil {
		panic(err)
	}
}

// Parse bulk updates and queue them
func queueClusterBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, cluster bulk")
	// var p []byte
	// req.Body.Read(p)
	// queue.Push()
}

func queueIndexBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, index bulk")
	// queue.Push(req.Body)
}

func queueTypeBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, type bulk")
	// queue.Push(req.Body)
}

// Helper to turn individual document updates into bulk JSON
func bulk(action string, req *http.Request) string {
	vars := mux.Vars(req)

	msg := fmt.Sprintf(
		"{\"index\":{\"_index\":\"%s\",\"_type\":\"%s\",\"_id\":\"%s\"}}\n",
		vars["index"], vars["type"], vars["id"])

	return msg + readBody(req) + "\n"
	// return msg + string.TrimRight(req.Body, `\n`) + `\n`
}

func readBody(req *http.Request) string {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "{}"
	}
	return strings.TrimRight(string(body), "\n")
}

// Parse individual document updates and queue them
func queueDeletes(res http.ResponseWriter, req *http.Request) {
	updates <- bulk("delete", req)
}

// Parse individual document updates and queue them
func queueUpdates(res http.ResponseWriter, req *http.Request) {
	updates <- bulk("index", req)
}

// Proxy read requests straight through to Elasticsearch
// TODO: for future enhancement, log some stats about this request
func proxy(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, proxy", req)
}
