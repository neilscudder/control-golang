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
  conn, ror := mpd.DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if ror != nil { log.Fatalln(ror) }
  return conn
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}

func mpdNoStatus(cmd string) {
  conn := mpdConnect()
  defer conn.Close()
  switch cmd {
    case "fw":
      ror := conn.Next(); er(ror)
    case "up":
      status, ror := conn.Status(); er(ror)
      var current int
      current, ror = strconv.Atoi(status["volume"]); er(ror)
      if current < 100 {
        new := current + 5
        ror = conn.SetVolume(new); er(ror)
      }
    case "dn":
      status, ror := conn.Status(); er(ror)
      var current int
      current, ror = strconv.Atoi(status["volume"]); er(ror)
      if current > 0 {
        new := current - 5
        ror = conn.SetVolume(new); er(ror)
      }
    case "random":
      status, ror := conn.Status(); er(ror)
      var current int
      current, ror = strconv.Atoi(status["random"]); er(ror)
      if current == 1 {
        ror = conn.Random(false); er(ror)
      } else {
        ror = conn.Random(true); er(ror)
      }
   }
}


func mpdStatus() string {
  conn := mpdConnect()
  defer conn.Close()
  bufferedStatus := ""
  currentStatus := ""
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
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
  t, ror := template.ParseFiles("res/gui.gotmp"); er(ror)
  t.Execute(w, p)
}

func api(w http.ResponseWriter, r *http.Request) {
  switch r.FormValue("a"){
    case "info":
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
