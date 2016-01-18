package main

import (
  "fmt"
  "log"
  "path"
  "strconv"
  "io/ioutil"
  "net/http"
  "html/template"
  "github.com/fhs/gompd/mpd"
  "github.com/nu7hatch/gouuid"
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

func mpdStatus(w http.ResponseWriter, r *http.Request) {
  conn := mpdConnect(r)
  defer conn.Close()
  status, ror := conn.Status(); er(ror)
  song, ror := conn.CurrentSong(); er(ror)
  if status["state"] == "play" && song["Title"] != "" {
    p := map[string]string{
      "title": song["Title"],
      "artist": song["Artist"],
      "album": song["Album"],
    }
    t, ror := template.ParseFiles("res/status.gotmp"); er(ror)
    t.Execute(w, p)
  } else if status["state"] == "play" {
    filename := path.Base(song["file"])
    directory := path.Dir(song["file"])
    p := map[string]string{
      "title": filename,
      "artist": song["Artist"],
      "album": directory,
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

type Params struct {
  APIURL string
  APIALT string
  LABEL string
  EMAIL string
  MPDPORT string
  MPDHOST string
  MPDPASS string
  KPASS string
  RPASS string
}

func (p *Params) save() error {
  filename := p.RPASS + "." + p.EMAIL
  byteP := []byte(fmt.Sprintf("%v", p)) 
  return ioutil.WriteFile(filename, byteP, 0600)
}

func gui(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "APIURL": r.FormValue("APIURL"),
    "APIALT": r.FormValue("APIALT"),
    "MPDPORT": r.FormValue("MPDPORT"),
    "LABEL": r.FormValue("LABEL"),
    "MPDHOST": r.FormValue("MPDHOST"),
    "MPDPASS": r.FormValue("MPDPASS"),
    "KPASS": r.FormValue("KPASS"),
  }
  t, ror := template.ParseFiles("res/gui.gotmp"); er(ror)
  t.Execute(w, p)
}

func authority(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "dummy": r.FormValue("dummy"),
  }
  t, ror := template.ParseFiles("res/authority.gotmp"); er(ror)
  t.Execute(w, p)
}

func authorize(w http.ResponseWriter, r *http.Request) {
// TODO validate data here
    controlURL := r.FormValue("GUIURL") + "/?"
    if r.FormValue("MPDPASS") != "" && r.FormValue("MPDHOST") != "" {
      controlURL += "MPDPASS=" + r.FormValue("MPDPASS") + "&MPDHOST=" + r.FormValue("MPDHOST")
    }
    if r.FormValue("MPDPASS") == "" && r.FormValue("MPDHOST") != "" {
      controlURL += "&MPDHOST=" + r.FormValue("MPDHOST")
    }
    if r.FormValue("MPDPORT") != "" { controlURL += "&MPDPORT=" + r.FormValue("MPDPORT") }
    if r.FormValue("LABEL") != "" { controlURL += "&LABEL=" + r.FormValue("LABEL") }
    if r.FormValue("EMAIL") != "" { controlURL += "&EMAIL=" + r.FormValue("EMAIL") }
    if r.FormValue("APIURL") != "" { controlURL += "&APIURL=" + r.FormValue("APIURL") }
    if r.FormValue("APIALT") != "" { controlURL += "&APIALT=" + r.FormValue("APIALT") }
    resetURL := controlURL
    controlURL += "&KPASS="
    resetURL += "&RPASS="
    rkey,ror := uuid.NewV4(); er(ror)
    ckey,ror := uuid.NewV4(); er(ror)
    resetURL += rkey.String()
    controlURL += ckey.String()

  p := map[string]string{
    "controlURL": controlURL,
    "resetURL": resetURL,
  }
  t, ror := template.ParseFiles("res/authorize.gotmp"); er(ror)
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
  http.HandleFunc("/authority", authority)
  http.HandleFunc("/authorize", authorize)
  http.ListenAndServe(":8080", nil)
}
