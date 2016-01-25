package authority

import (
  "errors"
  "log"
  "path/filepath"
  "encoding/json"
  "io/ioutil"
  "github.com/nu7hatch/gouuid"
)

func save(kpass string,rpass string,obj map[string]string) error {
  filename := "data/" + kpass + "." + rpass
  byteP,ror := json.Marshal(obj); er(ror)
  return ioutil.WriteFile(filename, byteP, 0600)
}

func Authenticate(kpass string) ([]byte,error) {
  file,ror := filepath.Glob("data/" + kpass + ".*"); er(ror)
  if file != nil {
    log.Printf("Authenticated: %v", kpass)
    return ioutil.ReadFile(file[0])
  }
  return nil,errors.New("Access Denied")
}

func Authorize(obj map[string]string) (string,string) {
  kpass,ror := uuid.NewV4(); er(ror)
  rpass,ror := uuid.NewV4(); er(ror)
  k := kpass.String()
  r := rpass.String()
  ror = save(k,r,obj); er(ror)
  return k,r
}

func er(ror error){
  if ror != nil { log.Fatalln(ror) }
}


