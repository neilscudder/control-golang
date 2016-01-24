package main

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
  "github.com/nu7hatch/gouuid"
  "internal/authority"
  "internal/mpdcache"
)

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/get", get)
  http.HandleFunc("/cmd", cmd)
  http.HandleFunc("/authority", authority)
  http.HandleFunc("/authorize", authorize)
  http.ListenAndServe(":8080", nil)
}

func gui(w http.ResponseWriter, r *http.Request) {
  var p *Params
  kpass := r.FormValue("KPASS")
  p = getParams(kpass)
  t, ror := template.ParseGlob("templates/gui/*"); er(ror)
  t.ExecuteTemplate(w, "GUI" ,p)
}
func get(w http.ResponseWriter, r *http.Request) {
  p,ror := getParams(r.FormValue("KPASS")); er(ror)
  switch r.FormValue("a"){
    case "info":
      w.Header().Set("Content-Type", "text/html")
      mpdcache.MpdStatus(w,r)
  }
}
func cmd(w http.ResponseWriter, r *http.Request) {
  log.Printf("API Call: " + r.FormValue("a") + " " + r.FormValue("LABEL"))
  mpdcache.MpdNoStatus(r)
}


type Params struct {
  APIURL,
  APIALT,
  LABEL,
  EMAIL,
  MPDPORT,
  MPDHOST,
  MPDPASS,
  KPASS string
}

func getParams(kpass string) *Params{
  var p Params
  byteP,err = authority.Authenticate(kpass)
  if err == nil {
    return json.Unmarshal(byteP, &p)
  }
  return nil,err
}

func authority(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "dummy": r.FormValue("dummy"),
  }
  t, ror := template.ParseFiles("templates/authority.html"); er(ror)
  t.Execute(w, p)
}

func authorize(w http.ResponseWriter, r *http.Request) {
  p := &Params{
    APIURL: r.FormValue("APIURL"),
    APIALT: r.FormValue("APIALT"),
    LABEL: r.FormValue("LABEL"),
    EMAIL: r.FormValue("EMAIL"),
    MPDPORT: r.FormValue("MPDPORT"),
    MPDHOST: r.FormValue("MPDHOST"),
    MPDPASS: r.FormValue("MPDPASS"),
    KPASS: r.FormValue("KPASS"),
  }
  cURL := r.FormValue("GUIURL") + "/?"
  if p.APIURL != "" { cURL += "&APIURL=" + p.APIURL }
  if p.APIALT != "" { cURL += "&APIALT=" + p.APIALT }
  rURL := cURL
  cURL += "&KPASS="
  rURL += "&RPASS="
  kpass,rpass,ror := authority.Authorize(p); er(ror)
  cURL += kpass
  rURL += rpass
  u := map[string]string{
    "controlURL": cURL,
    "resetURL": rURL,
  }
  t, ror := template.ParseFiles("templates/authorize.html"); er(ror)
  t.Execute(w,u)
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


