package main

import (
  "fmt"
  "log"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)


func mpdNext() {
  conn, err := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil { log.Fatalln(err) }
  defer conn.Close()

  err = conn.Next()
  if err != nil { log.Fatalln(err) }
}

func mpdStatus() string {
  conn, err := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil { log.Fatalln(err) }
  defer conn.Close()

  bufferedStatus := ""
  currentStatus := ""
  status, err := conn.Status()
  if err != nil { log.Fatalln(err) }
  song, err := conn.CurrentSong()
  if err != nil { log.Fatalln(err) }
  if status["state"] == "play" {
    currentStatus = fmt.Sprintf("%s - %s", song["Artist"], song["Title"])
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
    "APIURL": "192.168.9.114:8080/api",
    "APIALT": "localhost:8080",
    "MPDPORT": "6600",
    "LABEL": "PORTO",
    "MPDHOST": "192.168.9.108",
    "MPDPASS": "user",
    "KPASS": "dev",
  }
  t, err := template.ParseFiles("res/gui.gotmp")
  if err != nil { log.Fatalln(err) }
  t.Execute(w, p)
}

func api(w http.ResponseWriter, r *http.Request) {
  switch r.FormValue("a"){
    case "info":
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Content-Type", "text/html")
      fmt.Fprintf(w,mpdStatus())
    case "fw":
      mpdNext()
  }
}

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/api", api)
  http.ListenAndServe(":8080", nil)
}
