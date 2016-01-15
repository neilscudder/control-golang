package main

import (
  "fmt"
  "log"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)

func mpdStatus() string {
  conn, err := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil {
    log.Fatalln(err)
  }
  defer conn.Close()

  line := ""
  line1 := ""
  status, err := conn.Status()
  if err != nil {
    log.Fatalln(err)
  }
  song, err := conn.CurrentSong()
  if err != nil {
    log.Fatalln(err)
  }
  if status["state"] == "play" {
    line1 = fmt.Sprintf("%s - %s", song["Artist"], song["Title"])
  } else {
    line1 = fmt.Sprintf("State: %s", status["state"])
  }
  if line != line1 {
    line = line1
  }
  return line
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
  }
}

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/api", api)
  http.ListenAndServe(":8080", nil)
}
