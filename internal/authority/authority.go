package authority

import (
//  "fmt"
  "errors"
  "log"
  "path"
  "path/filepath"
  "strconv"
  "encoding/json"
  "io/ioutil"
  "net/http"
  "html/template"
  "github.com/nu7hatch/gouuid"
)


func save(kpass,rpass,obj string) error {
  filename := "data/" + kpass + "." + rpass
  byteP,ror := json.Marshal(obj); er(ror)
  return ioutil.WriteFile(filename, byteP, 0600)
}

func Authenticate(kpass string) string,error {
  file,ror := filepath.Glob("data/" + kpass + ".*"); er(ror)
  if file != nil {
    log.Printf("Authenticated: %v", kpass)
    return ioutil.ReadFile(file[0])
  }
  log.Printf("Access Denied: %v", kpass)
  return nil,errors.New("Access Denied")
}

func Authorize(obj map[string]string) string,string,error {
  kpass = string(uuid.NewV4())
  rpass = string(uuid.NewV4())
  ror := save(kpass,rpass,obj; er(ror)
  return kpass
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


