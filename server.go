package main

import (
  "bufio"
  "os"
  "fmt"
  "net/http"
  // "encoding/json"
  "path/filepath"
  "strings"
  "time"
)

var indexMap map[string][]string
var pathList []string
var indexFinished = false

var minSubstrLen = 3

var debug = false

///////////// INDEXING STUFF
func importFile(path string, c chan string) error {
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

  c <- path

  return nil
}

func monitorStatus(c chan string) {
  for i := 0; i < len(pathList); i++ {
    path := <- c
    if path == "foo" {
      fmt.Println("foobar")
    }
  }

  indexFinished = true
}

//////////// QUERY STUFF

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  path := r.FormValue("path")

  fmt.Println("Start index")

  os.Chdir(path)

  c := make(chan string)

  startTime := time.Now().UnixNano()

  // Just generate the path list because we want to know the number of paths before we start the monitor
  filepath.Walk("./", func(path string, _ os.FileInfo, _ error) error {
    pathList = append(pathList, path)

    return nil
  })

  go monitorStatus(c)

  for _, path := range pathList {
    if debug {
      fmt.Println("Visit ", path)
    }
    go importFile(path, c)
  }

  if debug { fmt.Println(pathList) }

  endTime := time.Now().UnixNano()

  elapsed := float32(endTime-startTime)/1E6

  fmt.Println("Index complete. ElapsedTime in ms: ", elapsed )

  fmt.Fprintf(w, `{"success": "true"}`)
}

func isIndexedHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "%t"}`, indexFinished)
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