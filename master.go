package main

import (
  "fmt"
  "net/http"
  "strings"
  "time"
  "sort"
  "io/ioutil"
)

type Ingest struct {
    SubstrIndex int
    Routes []Route
}

type Route struct {
    PathIndex int
    Line  int
}

var dictionary map[string]int

var indexMap map[int][]Route
var pathList []string
var lineRecord map[int]map[int]string
var indexFinished = false

var minSubstrLen = 20

var debug = false
var debugQueries = false

var servers = []string{"http://localhost:9091", "http://localhost:9092", "http://localhost:9093"}

func dedup(data []string ) []string {
  sort.Strings(data)

  length := len(data) - 1

  for i := 0; i < length; i++ {
    for j := i + 1; j <= length; j++ {
      if (data[i] == data[j]) {
        data[j] = data[length]
        data = data[0:length]
        length--
        j--
      }
    }
  }

  return data
}

//////////// Handlers

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  for _, path := range servers {
    http.Get(path + "/index?path=" + r.FormValue("path"))
  }

  fmt.Fprintf(w, `{"success": "true"}`)
}

func isIndexedHandler(w http.ResponseWriter, r *http.Request) {
  time.Sleep(2000 * time.Millisecond)

  indexFinished := true

  for _, path := range servers {
    resp, _ := http.Get(path + "/isIndexed")
    body, _ := ioutil.ReadAll(resp.Body)
    // defer resp.Body.Close()
    indexFinished = indexFinished && strings.Contains(string(body), "true")
  }

  if debug { fmt.Printf("M:Indexed?: %t", indexFinished) }
  fmt.Printf(" < M:Indexed?: %t > ", indexFinished)
  fmt.Fprintf(w, `{"success": %t}`, indexFinished)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
  q := r.FormValue("q")
  var results = []string{}

  for _, path := range servers {
    r, _ := http.Get(path + "/?q=" + q)
    body, _ := ioutil.ReadAll(r.Body)
    // defer r.Body.Close()
    if len(body) > 0 {
      results = append(results, strings.Split(string(body), ",")...)
    }
  }

  results = dedup(results)

  var response string

  if len(results) == 0 {
    response = `{"success": true,"results":[]}`
    fmt.Println("No results is probably not good!!!!!")
  }else{
    response = `{"success": true,"results":["`
    response = response + strings.Join(results, `","`) + `"]}`
  }

  if debugQueries {
    fmt.Println(q)
    fmt.Println(response)
  }
  fmt.Fprintf(w, response)
}

func main() {

  http.HandleFunc("/healthcheck", healthCheckHandler)
  http.HandleFunc("/index", indexHandler)
  http.HandleFunc("/isIndexed", isIndexedHandler)
  http.HandleFunc("/", queryHandler)

  fmt.Println("Ready to serve")
  http.ListenAndServe(":9090", nil)
}