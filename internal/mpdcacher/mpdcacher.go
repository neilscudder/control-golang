package mpdcacher

import (
//  "fmt"
  "log"
  "path"
  "path/filepath"
  "strconv"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect(r *http.Request) (*mpd.Client,error) {
  var p *Params
  kpass := r.FormValue("KPASS")
  p = authenticate(kpass)
  host := p.MPDHOST + ":" + p.MPDPORT
  pass := p.MPDPASS
  return mpd.DialAuthenticated("tcp", host, pass)
}

func MpdNoStatus(r *http.Request) {
  cmd := r.FormValue("a")
  conn,err := mpdConnect(r)
  if err != nil { return }
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  switch cmd {
    case "fw":
      ror := conn.Next(); er(ror)
    case "up":
      current, ror := strconv.Atoi(status["volume"]); er(ror)
      if current <= 95 {
        new := current + 5
        ror = conn.SetVolume(new); er(ror)
      }
    case "dn":
      current, ror := strconv.Atoi(status["volume"]); er(ror)
      if current >= 5 {
        new := current - 5
        ror = conn.SetVolume(new); er(ror)
      }
    case "random":
      current, ror := strconv.Atoi(status["random"]); er(ror)
      if current == 1 {
        ror = conn.Random(false); er(ror)
      } else {
        ror = conn.Random(true); er(ror)
      }
   }
}

func MpdStatus(w http.ResponseWriter, r *http.Request) {
  conn,err := mpdConnect(r)
  if err != nil { return }
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
  t, ror := template.ParseFiles("templates/status.html"); er(ror)
  if status["state"] == "play" && song["Title"] != "" {
    p := map[string]string{
      "title": song["Title"],
      "artist": song["Artist"],
      "album": song["Album"],
    }
    t.Execute(w, p)
  } else if status["state"] == "play" {
    filename := path.Base(song["file"])
    directory := path.Dir(song["file"])
    p := map[string]string{
      "title": filename,
      "artist": song["Artist"],
      "album": directory,
    }
    t.Execute(w, p)
  } else {
    p := map[string]string{
      "title": status["state"],
      "artist": "",
      "album": "",
    }
    t.Execute(w, p)
  }
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


