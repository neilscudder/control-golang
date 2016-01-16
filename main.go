package main

import (
  "fmt"
  "log"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect() *mpd.Client {
  conn, err := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil { log.Fatalln(err) }
  return conn
}

func mpdNoStatus(cmd string) {
  conn := mpdConnect()
  defer conn.Close()
  switch cmd {
    case "fw":
      err := conn.Next()
      if err != nil { log.Fatalln(err) }
    case "up":
      err := conn.SetVolume(90)
      if err != nil { log.Fatalln(err) }
    case "dn":
      err := conn.SetVolume(30)
      if err != nil { log.Fatalln(err) }
   }
}


func mpdStatus() string {
  conn := mpdConnect()
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
    "APIURL": "http://192.168.9.114:8080/api",
    "APIALT": "",
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
  log.Printf("API Call: " + r.FormValue("a") + " " + r.FormValue("LABEL"))
  switch r.FormValue("a"){
    case "info":
      w.Header().Set("Status", "200")
      w.Header().Set("Access-Control-Allow-Origin", "*")
      w.Header().Set("Content-Type", "text/html")
      fmt.Fprintf(w,mpdStatus())
    default:
      mpdNoStatus(r.FormValue("a"))
  }
}

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/api", api)
  http.ListenAndServe(":8080", nil)
}
