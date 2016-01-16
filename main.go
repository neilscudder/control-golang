package main

import (
  "fmt"
  "log"
  "strconv"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect() *mpd.Client {
  conn, err := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil { log.Fatalln(err) }
  return conn
}

func e(err error){
  if err != nil { log.Fatalln(err) }
}

func mpdNoStatus(cmd string) {
  conn := mpdConnect()
  defer conn.Close()
  switch cmd {
    case "fw":
      err := conn.Next(); e(err)
    case "up":
      status, err := conn.Status(); e(err)
      var current int
      current, err = strconv.Atoi(status["volume"]); e(err)
      if current < 100 {
        new := current + 5
        err = conn.SetVolume(new); e(err)
      }
    case "dn":
      status, err := conn.Status(); e(err)
      var current int
      current, err = strconv.Atoi(status["volume"]); e(err)
      if current > 0 {
        new := current - 5
        err = conn.SetVolume(new); e(err)
      }
    case "random":
      status, err := conn.Status(); e(err)
      var current int
      current, err = strconv.Atoi(status["random"]); e(err)
      if current == 1 {
        err = conn.Random(false); e(err)
      } else {
        err = conn.Random(true); e(err)
      }
   }
}


func mpdStatus() string {
  conn := mpdConnect()
  defer conn.Close()
  bufferedStatus := ""
  currentStatus := ""
  status, err := conn.Status(); e(err)
   song, err := conn.CurrentSong(); e(err)
  if status["state"] == "play" {
    currentStatus = fmt.Sprintf("%s - %s, (%s)", song["Artist"], song["Title"], status["volume"])
  } else {
    currentStatus = fmt.Sprintf("State: %s", status["state"])
  }
  if bufferedStatus != currentStatus {
    bufferedStatus = currentStatus
  }
  return bufferedStatus
}

func gui(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "APIURL": "http://192.168.9.114:8080/api",
    "APIALT": "",
    "MPDPORT": "6600",
    "LABEL": "PORTO",
    "MPDHOST": "192.168.9.108",
    "MPDPASS": "user",
    "KPASS": "dev",
  }
  t, err := template.ParseFiles("res/gui.gotmp"); e(err)
  t.Execute(w, p)
}

func api(w http.ResponseWriter, r *http.Request) {
  switch r.FormValue("a"){
    case "info":
      w.Header().Set("Status", "200")
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Content-Type", "text/html")
      fmt.Fprintf(w,mpdStatus())
    default:
      log.Printf("API Call: " + r.FormValue("a") + " " + r.FormValue("LABEL"))
      mpdNoStatus(r.FormValue("a"))
  }
}

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/api", api)
  http.ListenAndServe(":8080", nil)
}
