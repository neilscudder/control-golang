package control

import (
  "fmt"
  "net/http"
  "mpd"
)

func mpdConnect() {
  conn, err := DialAuthenticated("tcp", "192.168.9.108:6600", "user")
  if err != nil {
    log.Fatalln(err)
  }
  defer conn.Close()
}

func mpdStatus() {
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
    fmt.Println(line)
  }
}

func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, mpdStatus())]
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
