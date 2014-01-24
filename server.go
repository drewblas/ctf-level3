package main

import (
  "bufio"
  "os"
  "fmt"
  "net/http"
  // "encoding/json"
  "path/filepath"
  "strings"
)

type IndexValue struct {
    Substr, Location string
}

var indexMap map[string][]string
var pathList []string

var minSubstrLen = 3

var debug = false

type QueryResponse struct {
    Success string
    Body string
    Time int64
}

///////////// INDEXING STUFF
func importFile(path string, c chan IndexValue) error {
  file, err := os.Open(path)
  if err != nil {
    return err
  }
  defer file.Close()

  lineScanner := bufio.NewScanner(file)
  var lineNum = 0
  var route = ""

  for lineScanner.Scan() {
    lineNum += 1
    route = fmt.Sprintf("%s:%d", path, lineNum)

    // r := strings.NewReader( strings.ToLower(lineScanner.Text()) )
    r := strings.NewReader( lineScanner.Text() )
    
    wordScanner := bufio.NewScanner( r )
    wordScanner.Split(bufio.ScanWords)

    for wordScanner.Scan() {
      str := wordScanner.Text()

      // for all substrings
      for substrLen := minSubstrLen; substrLen <= len(str); substrLen++ {
        for i := 0; i <= len(str) - substrLen; i++ {
          substr := str[i:i+substrLen]

          // if debug {
          //   fmt.Println("Indexing ", , " - ", route)
          // }

          list, _ := indexMap[substr]
          // if present {
            // indexMap[substr] = []string{route}
          // }else{
            indexMap[substr] = append(list, route)
          // }
        }
      }
    }
  }

  return nil
}

//////////// QUERY STUFF

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  path := r.FormValue("path")

  os.Chdir(path)

  c := make(chan IndexValue)

  // go storeIndexes(c)

  filepath.Walk("./", func(path string, _ os.FileInfo, _ error) error {
    if debug {
      fmt.Println("Visit ", path)
    }
    pathList = append(pathList, path)
    importFile(path, c)

    return nil
  })

  if debug { fmt.Println(pathList) }

  fmt.Fprintf(w, `{"success": "true"}`)
}

func isIndexedHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
  // q := strings.ToLower( r.FormValue("q") )
  q := r.FormValue("q")
  var results []string

  if len(q) < minSubstrLen {
    fmt.Println("ERROR TOO SMALL")
    results = []string{}
  }else{
    results = indexMap[q]
  }

  var response string

  if len(results) == 0 {
    response = `{"success": true,"results":[]}`
  }else{
    response = `{"success": true,"results":["`
    response = response + strings.Join(results, `","`) + `"]}`
  }

  if debug { fmt.Println(response) }
  fmt.Fprintf(w, response)
}

func main() {
  // m := map[string]string{
  //   "foo": "bar",
  //   "drew": "fus",
  // }
  // j, _ := json.Marshal(m)

  // fmt.Println(string(j))

  pathList = []string{}
  indexMap = make(map[string][]string)

  http.HandleFunc("/healthcheck", healthCheckHandler)
  http.HandleFunc("/index", indexHandler)
  http.HandleFunc("/isIndexed", isIndexedHandler)
  http.HandleFunc("/", queryHandler)

  fmt.Println("Ready to serve")
  http.ListenAndServe(":9090", nil)
}