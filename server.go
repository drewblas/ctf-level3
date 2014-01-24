package main

import (
  "bufio"
  "os"
  "fmt"
  "net/http"
  // "encoding/json"
  "path/filepath"
)

var indexMap map[string][]string

///////////// INDEXING STUFF

func visit(path string, f os.FileInfo, err error) error {
  fmt.Printf("Visited: %s\n", path)
  return nil
} 

// readLines reads a whole file into memory
// and returns a slice of its lines.
func importFile(path string) ([]string, error) {
  file, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  var lineNum = 1

  for scanner.Scan() {
    lineNum += 1
    lines = append(lines, scanner.Text())
  }

  return lines, scanner.Err()
}

//////////// QUERY STUFF

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  path := r.FormValue("path")

  os.Chdir(path)

  filepath.Walk("./", visit)

  fmt.Fprintf(w, `{"success": "true"}`)
}

func isIndexedHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
  

  // j, err := json.Marshal(m)
}

func main() {
  // m := map[string]string{
  //   "foo": "bar",
  //   "drew": "fus",
  // }
  // j, _ := json.Marshal(m)

  // fmt.Println(string(j))

  http.HandleFunc("/healthcheck", healthCheckHandler)
  http.HandleFunc("/index", indexHandler)
  http.HandleFunc("/isIndexed", isIndexedHandler)
  http.HandleFunc("/", queryHandler)
  http.ListenAndServe(":9090", nil)
}