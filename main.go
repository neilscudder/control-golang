package main

import (
  "fmt"
  "log"
  "encoding/json"
  "net/http"
  "html/template"
  "github.com/neilscudder/control-golang/authority"
  "github.com/neilscudder/control-golang/mpdcacher"
)

func main() {
  http.HandleFunc("/", gui)
  http.HandleFunc("/get", get)
  http.HandleFunc("/authority", setup)
  http.HandleFunc("/authorize", auth)
  http.ListenAndServe(":8080", nil)
//  ror := http.ListenAndServeTLS(":8080", "config/server.pem", "config/server.key", nil);er(ror)
}

func gui(w http.ResponseWriter, r *http.Request) {
  var p map[string]string
  kpass := r.FormValue("KPASS")
  p,err := getParams(kpass)
  if err != nil {
    log.Println(err.Error())
    fmt.Fprintf(w, err.Error())
    return
  }
  t, ror := template.ParseGlob("templates/gui/*"); er(ror)
  t.ExecuteTemplate(w, "GUI" ,p)
}
func get(w http.ResponseWriter, r *http.Request) {
  kpass := r.FormValue("KPASS")
  p,err := getParams(kpass)
  if err != nil {
    log.Println(err.Error())
    fmt.Fprintf(w, err.Error())
    return
  }
  cmd := r.FormValue("a")
//  w.Header().Set("Content-Type", "text/html")
  u := mpdcacher.MpdStatus(cmd,p)
  t, ror := template.ParseFiles("templates/status.html"); er(ror)
  t.Execute(w,u)
}

func getParams(kpass string) (map[string]string,error){
  var p map[string]string
  byteP,err := authority.Authenticate(kpass)
  if err == nil {
    err = json.Unmarshal(byteP,&p)
    return p,err
  }
  return p,err
}

func setup(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "dummy": r.FormValue("dummy"),
  }
  t, ror := template.ParseFiles("templates/authority.html"); er(ror)
  t.Execute(w, p)
}

func auth(w http.ResponseWriter, r *http.Request) {
  p := map[string]string{
    "APIURL": r.FormValue("APIURL"),
    "LABEL": r.FormValue("LABEL"),
    "EMAIL": r.FormValue("EMAIL"),
    "MPDPORT": r.FormValue("MPDPORT"),
    "MPDHOST": r.FormValue("MPDHOST"),
    "MPDPASS": r.FormValue("MPDPASS"),
    "KPASS": r.FormValue("KPASS"),
  }
  cURL := r.FormValue("GUIURL") + "/?"
  if p["APIURL"] != "" { cURL += "&APIURL=" + p["APIURL"] }
  rURL := cURL
  cURL += "&KPASS="
  rURL += "&RPASS="
  byteP,ror := json.Marshal(p); er(ror)
  kpass,rpass := authority.Authorize(byteP)
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


