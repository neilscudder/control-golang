package authority

import (
	"errors"
	"github.com/nu7hatch/gouuid"
	"io/ioutil"
	"log"
	"path/filepath"
)

func save(kpass string, rpass string, obj []byte) error {
	filename := "data/" + kpass + "." + rpass
	return ioutil.WriteFile(filename, obj, 0600)
}

// Authenticate returns the contents of a text file named kpass.*
// If the file doesn't exist, returns "Access Denied" error
func Authenticate(kpass string) ([]byte, error) {
	file, ror := filepath.Glob("data/" + kpass + ".*")
	er(ror)
	if file != nil {
		return ioutil.ReadFile(file[0])
	}
	return nil, errors.New("Access Denied")
}

// Authorize stores obj as text in a file named with two newly generated uuids
// Returns these uuids
func Authorize(obj []byte) (string, string) {
	kpass, ror := uuid.NewV4()
	er(ror)
	rpass, ror := uuid.NewV4()
	er(ror)
	k := kpass.String()
	r := rpass.String()
	ror = save(k, r, obj)
	er(ror)
	return k, r
}

func er(ror error) {
	if ror != nil {
		log.Fatalln(ror)
	}
}
