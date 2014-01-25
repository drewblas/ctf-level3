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
  "sort"
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

var minSubstrLen = 6

var debug = false
var debugQueries = false

var indexStart int64

func RoutesToStrings(routes []Route) []string {
  strs := make([]string, len(routes))

  for i, r := range routes {
    strs[i] = fmt.Sprintf("%s:%d", pathList[r.PathIndex], r.Line)
  }

  return strs
}

///////////// INDEXING STUFF
func importWorker(workerNum int, pathChan chan int, statusChan chan string, ingestChan chan Ingest) {
  for i := range pathChan {
    importFile(i, statusChan, ingestChan)
  }
}

func importFile(pathIndex int, statusChan chan string, ingestChan chan Ingest) error {
  path := pathList[pathIndex]

  lineRecord[pathIndex] = make(map[int]string)

  file, err := os.Open(path)
  if err != nil {
    return err
  }
  defer file.Close()

  lineScanner := bufio.NewScanner(file)
  var lineNum = 0

  for lineScanner.Scan() {
    lineNum += 1

    lineRecord[pathIndex][lineNum] = lineScanner.Text()

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

          if idx,ok := dictionary[substr]; ok {
            // if debug {
            //   fmt.Println("Indexing ", , " - ", route)
            // }

            list, _ := indexMap[idx]
            ingestChan <- Ingest{idx, append(list, Route{pathIndex, lineNum})}
          }
        }
      }
    }
  }

  statusChan <- path

  return nil
}

// Handles a synchronized writing of route data
func ingestRoutes(ingestChan chan Ingest) {
  for i := range ingestChan {
    indexMap[i.SubstrIndex] = i.Routes
  }
}

func monitorStatus(c chan string) {
  startTime := time.Now().UnixNano()

  for i := 0; i < len(pathList); i++ {
    path := <- c
    if path == "foo" {
      fmt.Println("foobar")
    }
  }

  indexFinished = true

  endTime := time.Now().UnixNano()

  elapsed := float32(endTime-startTime)/1E6

  fmt.Println("Monitor index finished. ElapsedTime in ms: ", elapsed )
}

//////////// QUERY STUFF
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

func searchManual(q string) []string {
  results := []string{}

  for idx, path := range(pathList) {
    for lineNum, text := range lineRecord[idx] {
      if strings.Contains(text, q) {
        results = append(results, fmt.Sprintf("%s:%d", path, lineNum))
      }
    }
  }

  return results
}

//////////// Handlers

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, `{"success": "true"}`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
  indexStart = time.Now().UnixNano()
  path := r.FormValue("path")

  fmt.Println("Start index")

  os.Chdir(path)

  statusChan := make(chan string)
  pathChan := make(chan int)
  ingestChan := make(chan Ingest)

  startTime := time.Now().UnixNano()

  // Just generate the path list because we want to know the number of paths before we start the monitor
  filepath.Walk("./", func(path string, _ os.FileInfo, _ error) error {
    pathList = append(pathList, path)

    return nil
  })

  fmt.Println("##",len(pathList),"##")

  go monitorStatus(statusChan)
  go ingestRoutes(ingestChan)
  for numWorkers := 0; numWorkers < 4; numWorkers++ {
    go importWorker(numWorkers, pathChan, statusChan, ingestChan)
  }

  go func () {
    for i, _ := range pathList {
      // if debug {
        // fmt.Println(i, "/", len(pathList), " Visit ", path)
        fmt.Printf("%d ..", i)
      // }
      pathChan <- i
    }

    close(pathChan)

    if debug { fmt.Println(pathList) }
  }()

  endTime := time.Now().UnixNano()

  elapsed := float32(endTime-startTime)/1E6

  if debug { fmt.Println("Index API call complete. (but indexing is still happening in the background) ElapsedTime in ms: ", elapsed ) }

  fmt.Fprintf(w, `{"success": "true"}`)
}

func isIndexedHandler(w http.ResponseWriter, r *http.Request) {
  showFinished := indexFinished

  // current := time.Now().UnixNano()
  // if current - indexStart < 120E9 {
  //   showFinished = false
  // }

  if debug { fmt.Printf("Indexed?: %t", indexFinished) }
  fmt.Printf(" < Indexed?: %t > ", indexFinished)
  fmt.Fprintf(w, `{"success": %t}`, showFinished)
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
  // q := strings.ToLower( r.FormValue("q") )
  q := r.FormValue("q")
  var results []string

  if len(q) < minSubstrLen {
    fmt.Println("ERROR TOO SMALL")
    results = searchManual(q)
  }else{
    idx,ok := dictionary[q]

    if ok {
      results = dedup( RoutesToStrings(indexMap[idx]) )
    }
  }

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

func loadDictionary() {
  // matches, _ := filepath.Glob("test/data/words*")

  file, _ := os.Open("./words")

  defer file.Close()

  lineScanner := bufio.NewScanner(file)

  lineNum := 0

  for lineScanner.Scan() {
    lineNum += 1
    // if len(lineScanner.Text()) >= minSubstrLen {
      dictionary[lineScanner.Text()] = lineNum
    // }
  }

  fmt.Println("Dictionary Size: ", len(dictionary))

  if debug { fmt.Println("Dictionary Loaded") } 
}

func main() {
  // m := map[string]string{
  //   "foo": "bar",
  //   "drew": "fus",
  // }
  // j, _ := json.Marshal(m)

  // fmt.Println(string(j))

  pathList = []string{}
  indexMap = make(map[int][]Route)
  dictionary = make(map[string]int)
  lineRecord = make(map[int]map[int]string)

  loadDictionary()

  http.HandleFunc("/healthcheck", healthCheckHandler)
  http.HandleFunc("/index", indexHandler)
  http.HandleFunc("/isIndexed", isIndexedHandler)
  http.HandleFunc("/", queryHandler)

  fmt.Println("Ready to serve")
  http.ListenAndServe(":9090", nil)
}