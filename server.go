package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

// Basic options for the batching size and frequency
const batchSize int = 1000                               // maximum documents per batch
const batchPause time.Duration = time.Millisecond * 1000 // milliseconds between batch updates

// Configure the server from the environment
var serverPort = os.Getenv("PORT")
var elasticsearchUrl = os.Getenv("BONSAI_URL")

// Buffered channel for updates. These should be strings in the Bulk API
// format.
var updates chan string = make(chan string, batchSize)

// Start up the processor and the HTTP server.
func main() {
	go processor()
	server()
}

// Processor receives batch updates from the updates channel and sends them to
// Elasticsearch, subject to a short cooldown interval. The strings in the
// updates channel should be formatted for the bulk API
func processor() {

	// Create an array of strings to help prepare our bulk request, set its
	// length to zero.
	batch := make([]string, batchSize)
	batch = batch[0:0]

	// A short interval to wait between requests. This gives a bit of time for
	// Elasticsearch to think, as well as our channel to refill.
	limiter := time.Tick(batchPause)

	// Infinite loop to pull updates out of a channel, and periodically send them
	// to Elasticsearch
	for {
		select {

		// Pull as many updates as we can and batch them into a slice.
		case update := <-updates:
			batch = append(batch, update)

		// When the channel is empty, combine the batch, send it to ES, then wait a
		// bit.
		default:
			if len(batch) > 0 {
				log.Println("Processed", len(batch), "updates")

				// POST the batch to the Elasticsearch cluster _bulk handler
				resp, err := http.Post(
					elasticsearchUrl+"/_bulk", "application/json",
					strings.NewReader(strings.Join(batch, "")))

				if err != nil {
					log.Println("Error:", err)
				}

				// TODO: log something interesting from the response
				if false {
					body, _ := ioutil.ReadAll(resp.Body)
					log.Println("Response", string(body))
				}

				// Reset the batch array
				batch = batch[0:0]
			}
			<-limiter
		}
	}
}

// Set up and run an HTTP server that intercepts updates and formats them for
// batches, and also proxies searches through to Elasticsearch.
func server() {

	r := mux.NewRouter()
	s := r.Methods("PUT", "POST").Subrouter()

	// Match the various _bulk handlers
	s.Path("/_bulk").
		HandlerFunc(queueClusterBulk)
	s.Path("/{index}/_bulk").
		HandlerFunc(queueIndexBulk)
	s.Path("/{index}/{type}/_bulk").
		HandlerFunc(queueTypeBulk)

	// Handle individual document updates
	r.Methods("POST", "PUT").
		Path("/{index}/{type}/{id}").
		HandlerFunc(queueUpdates)

	// Handle individual document deletes as well
	r.Methods("DELETE").
		Path("/{index}/{type}/{id}").
		HandlerFunc(queueDeletes)

	// Proxy read requests straight through to Elasticsearch
	r.PathPrefix("/").
		HandlerFunc(proxy)

	// Start the HTTP server
	http.Handle("/", r)
	fmt.Println("listening on " + serverPort + "...")
	err := http.ListenAndServe(":"+serverPort, nil)
	if err != nil {
		panic(err)
	}
}

//
// Request Handlers
//

// Proxy read requests straight through to Elasticsearch
// TODO: parse the URL and initialize the proxy somewhere else?
func proxy(res http.ResponseWriter, req *http.Request) {
	// Set up a simple proxy for read requests
	targetUrl, err := url.Parse(elasticsearchUrl)
	if err != nil {
		log.Println("ERROR - Couldn't parse URL:", elasticsearchUrl, " - ", err)
		fmt.Fprintln(res, "Couldn't parse BONSAI_URL: '", elasticsearchUrl, "'")
	} else {
		proxy := httputil.NewSingleHostReverseProxy(targetUrl)
		proxy.ServeHTTP(res, req)
	}
}

// Parse cluster bulk updates and queue them.
func queueClusterBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, cluster bulk")
	// var p []byte
	// req.Body.Read(p)
	// queue.Push()
}

// Parse index bulk requests and queue them.  The index name may have been
// implicit from the URL and omitted in the payload, so we have to deal with
// that.
func queueIndexBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, index bulk")
	// queue.Push(req.Body)
}

// Parse index bulk requests and queue them.  The index and type names may have
// been implicit from the URL and omitted in the payload, so we have to deal
// with that.
func queueTypeBulk(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, type bulk")
	// queue.Push(req.Body)
}

// Parse individual document updates and queue them
func queueDeletes(res http.ResponseWriter, req *http.Request) {
	updates <- bulk("delete", req)
}

// Parse individual document updates and queue them
func queueUpdates(res http.ResponseWriter, req *http.Request) {
	updates <- bulk("index", req)
}

//
// Request helpers
//

// Helper to read the body of a document update request, and enforce a trailing
// newline.
func readBody(req *http.Request) string {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "{}"
	}
	return strings.TrimRight(string(body), "\n")
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
