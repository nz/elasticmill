package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iron-io/iron_mq_go"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var ironmqProject = os.Getenv("IRON_MQ_PROJECT_ID")
var ironmqToken = os.Getenv("IRON_MQ_TOKEN")
var ironmqRegion = ironmq.IronAWSUSEast

var queueName = "elasticsearch_updates"

var client = ironmq.NewClient(ironmqProject, ironmqToken, ironmqRegion)
var queue = client.Queue(queueName)

func main() {
  go processor()
  
	fmt.Println(queue)

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

	fmt.Println("listening on " + os.Getenv("PORT") + "...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
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
	fmt.Fprintln(res, "hello, document delete")
	fmt.Fprint(res, bulk("delete", req))
	queue.Push(bulk("delete", req))
}

// Parse individual document updates and queue them
func queueUpdates(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, document delete")
	queue.Push(bulk("index", req))
}

// Proxy read requests straight through to Elasticsearch
// TODO: for future enhancement, log some stats about this request
func proxy(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, proxy")
}
