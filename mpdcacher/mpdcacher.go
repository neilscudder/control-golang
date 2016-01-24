package mpdcacher

import (
  "log"
  "path"
  "strconv"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect(params map[string]string) (*mpd.Client,error) {
  host := params.MPDHOST + ":" + params.MPDPORT
  pass := params.MPDPASS
  return mpd.DialAuthenticated("tcp", host, pass)
}

func MpdStatus(cmd string,params map[string]string) map[string]string {
  var p map[string]string
  conn,err := mpdConnect(params)
  if err != nil { return }
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
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
    case "info":
      // nothing
   }
  return func() map[string]string{
    if status["state"] == "play" && song["Title"] != "" {
      p = map[string]string{
	"title": song["Title"],
	"artist": song["Artist"],
	"album": song["Album"],
      }
    } else if status["state"] == "play" {
      filename := path.Base(song["file"])
      directory := path.Dir(song["file"])
      p = map[string]string{
	"title": filename,
	"artist": song["Artist"],
	"album": directory,
      }
    } else {
      p = map[string]string{
	"title": status["state"],
	"artist": "",
	"album": "",
      }
    }
    return p
  }
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


