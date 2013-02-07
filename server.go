package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", intercept)
	fmt.Println("listening...")
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}

func intercept(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "hello, world")
}
