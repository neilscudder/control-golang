package mpdcacher

import (
  "log"
  "path"
  "strconv"
  "github.com/fhs/gompd/mpd"
)

func mpdConnect(p map[string]string) (*mpd.Client,error) {
  host := p["MPDHOST"] + ":" + p["MPDPORT"]
  pass := p["MPDPASS"]
  return mpd.DialAuthenticated("tcp", host, pass)
}

func MpdStatus(cmd string,params map[string]string) map[string]string {
  conn,ror := mpdConnect(params); er(ror)
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
  return getStatus(song,status)
}

func getStatus(song,status map[string]string) map[string]string{
  var p map[string]string
  if status["state"] == "play" && song["Title"] != "" {
    p = map[string]string{
      "Album": song["Album"],
      "Artist": song["Artist"],
      "Title": song["Title"],
    }
  } else if status["state"] == "play" {
    filename := path.Base(song["file"])
    directory := path.Dir(song["file"])
    p = map[string]string{
      "File": filename,
      "Artist": song["Artist"],
      "Folder": directory,
    }
  } else {
    p = map[string]string{
      "State": status["state"],
    }
  }
  return p
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


