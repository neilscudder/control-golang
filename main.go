package main

import (
  "log"
  "strconv"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect(r *http.Request) *mpd.Client {
  host := r.FormValue("MPDHOST") + ":" + r.FormValue("MPDPORT")
  pass := r.FormValue("MPDPASS")
  conn, ror := mpd.DialAuthenticated("tcp", host, pass); er(ror)
  return conn
}

func mpdNoStatus(r *http.Request) {
  cmd := r.FormValue("a")
  conn := mpdConnect(r)
  defer conn.Close()
  switch cmd {
    case "fw":
      ror := conn.Next(); er(ror)
    case "up":
      status, ror := conn.Status(); er(ror)
      var current int
      current, ror = strconv.Atoi(status["volume"]); er(ror)
      if current <= 95 {
        new := current + 5
        ror = conn.SetVolume(new); er(ror)
      }
    case "dn":
      status, ror := conn.Status(); er(ror)
      var current int
      current, ror = strconv.Atoi(status["volume"]); er(ror)
      if current >= 5 {
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

func mpdStatus(w http.ResponseWriter, r *http.Request) {
  conn := mpdConnect(r)
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
  if status["state"] == "play" {
    p := map[string]string{
      "title": song["Title"],
      "artist": song["Artist"],
      "album": song["Album"],
    }
    t, ror := template.ParseFiles("res/status.gotmp"); er(ror)
    t.Execute(w, p)
  } else {
    p := map[string]string{
      "title": status["state"],
      "artist": "",
      "album": "",
    }
    t, ror := template.ParseFiles("res/status.gotmp"); er(ror)
    t.Execute(w, p)
  }
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
      mpdStatus(w,r)
    default:
      log.Printf("API Call: " + r.FormValue("a") + " " + r.FormValue("LABEL"))
      mpdNoStatus(r)
  }
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/api", api)
  http.ListenAndServe(":8080", nil)
}
