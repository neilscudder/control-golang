package authority

import (
  "errors"
  "log"
  "path/filepath"
  "io/ioutil"
  "github.com/nu7hatch/gouuid"
)

func save(kpass string,rpass string,obj []byte) error {
  filename := "data/" + kpass + "." + rpass
  return ioutil.WriteFile(filename,obj,0600)
}

func Authenticate(kpass string) ([]byte,error) {
  file,ror := filepath.Glob("data/" + kpass + ".*"); er(ror)
  if file != nil {
    return ioutil.ReadFile(file[0])
  }
  return nil,errors.New("Access Denied")
}

func Authorize(obj []byte) (string,string) {
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


