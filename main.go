package main

import (
  "fmt"
  "log"
  "net/http"
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

func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, mpdStatus())
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
