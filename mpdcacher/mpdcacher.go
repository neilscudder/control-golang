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

func MpdStatus(cmd string,params map[string]string) map[string]map[int]map[string]string {
  var deets map[int]map[string]string
  conn,ror := mpdConnect(params); er(ror)
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  switch cmd {
    case "fw":
      ror := conn.Next(); er(ror)
    case "up":
      current, ror := strconv.Atoi(status["volume"]); er(ror)
      if current <= 95 {
        current = current + 5
        ror = conn.SetVolume(current); er(ror)
      }
      deets = map[int]map[string]string{
        1: map[string]string{
          "Volume": strconv.Itoa(current),
        },
      }
    case "dn":
      current, ror := strconv.Atoi(status["volume"]); er(ror)
      if current >= 5 {
        current = current - 5
        ror = conn.SetVolume(current); er(ror)
      }
      deets = map[int]map[string]string{
        1: map[string]string{
          "Volume": strconv.Itoa(current),
        },
      }
    case "random":
      current, ror := strconv.Atoi(status["random"]); er(ror)
      if current == 1 {
        current = 0
        ror = conn.Random(false); er(ror)

      } else {
        current = 1
        ror = conn.Random(true); er(ror)
      }
      deets = map[int]map[string]string{
        1: map[string]string{
          "Random": strconv.Itoa(current),
        },
      }
    case "info":
      // nothing
    }
    song, ror := conn.CurrentSong(); er(ror)
    a := map[string]map[int]map[string]string {
      "info": getInfo(conn),
      "deets": deets,
      "title": map[int]map[string]string{
        1: map[string]string{
          "Title": song["Title"],
        },
      },
    }
  return a
}

func getInfo(conn *mpd.Client) map[int]map[string]string{
  var p map[int]map[string]string
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
  if status["state"] == "play" && song["Title"] != "" {
    p = map[int]map[string]string{
      1: map[string]string{
        "Artist": song["Artist"],
      },
      2: map[string]string{
        "Album": song["Album"] + " (" + song["Date"] + ")",
      },
    }
  } else if status["state"] == "play" {
    filename := path.Base(song["file"])
    directory := path.Dir(song["file"])
    p = map[int]map[string]string{
      1: map[string]string{
        "File Name": filename,
      },
      2: map[string]string{
        "Folder": directory,
      },
    }
  } else {
    p = map[int]map[string]string{
      1: map[string]string{
        "State": status["state"],
      },
    }
  }
  return p
}

func er(ror error){
  if ror != nil { log.Println(ror) }
}


