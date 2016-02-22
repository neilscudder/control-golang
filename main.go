package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	gz "github.com/NYTimes/gziphandler"
	"github.com/neilscudder/control-golang/authority"
	"github.com/neilscudder/control-golang/mpdcacher"
	"html/template"
	"log"
	"net/http"
)

var searchBuffer = make(map[string][]string)

func main() {
	resGz := gz.GzipHandler(http.FileServer(http.Dir("res/")))
	getGz := gz.GzipHandler(http.HandlerFunc(get))
	setupGz := gz.GzipHandler(http.HandlerFunc(setup))
	authGz := gz.GzipHandler(http.HandlerFunc(setup))
	postGz := gz.GzipHandler(http.HandlerFunc(post))
	guiGz := gz.GzipHandler(http.HandlerFunc(gui))

	http.Handle("/res/", resGz)
	http.Handle("/get", getGz)
	http.Handle("/authority", setupGz)
	http.Handle("/authorize", authGz)
	http.Handle("/post", postGz)
	http.Handle("/", guiGz)

	var pemfile = flag.String("pem", "", "Path to pem file")
	var keyfile = flag.String("key", "", "Path to key file")
	flag.Parse()
	if *pemfile == "" {
		ror := http.ListenAndServe(":8080", nil)
		er(ror)
	} else {
		config := &tls.Config{MinVersion: tls.VersionTLS10}
		server := &http.Server{Addr: ":443", Handler: nil, TLSConfig: config}
		ror := server.ListenAndServeTLS(*pemfile, *keyfile)
		er(ror)
	}
}

func gui(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html")
	var p = make(map[string]string)
	kpass := r.FormValue("KPASS")
	p, err := getParams(kpass)
	p["KPASS"] = kpass
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}
	switch r.FormValue("a") {
	case "search":
		query := "any \""
		query += r.FormValue("search")
		query += "\""
		s := mpdcacher.Search(query, p)
		searchBuffer[kpass] = s.Files
		t, ror := template.ParseFiles("templates/search.html")
		er(ror)
		t.Execute(w, s)
	case "browser":
		t, ror := template.ParseGlob("templates/browser/*")
		er(ror)
		t.ExecuteTemplate(w, "GUI", p)
	default:
		t, ror := template.ParseGlob("templates/gui/*")
		er(ror)
		t.ExecuteTemplate(w, "GUI", p)
	}
}
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	kpass := r.FormValue("KPASS")
	p, err := getParams(kpass)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}
	cmd := r.FormValue("a")
	if cmd == "info" {
		w.Header().Set("Content-Type", "text/html")
		u := mpdcacher.MpdStatus(cmd, p)
		t, ror := template.ParseFiles("templates/status.html")
		er(ror)
		t.Execute(w, u)
	} else {
		s := mpdcacher.MpdState(cmd, p)
		state, _ := json.Marshal(s)
		w.Header().Set("Content-Type", "application/json")
		w.Write(state)
	}
}

func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	kpass := r.FormValue("KPASS")
	p, err := getParams(kpass)
	// fmt.Println(r.FormValue("KPASS"))
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}
	//target := r.FormValue("b")
	targets := searchBuffer[kpass]
	w.Header().Set("Content-Type", "text/html")
	mpdcacher.MpdPlay(p, targets)
	//fmt.Println(targets)
	ok := []byte("ok")
	w.Write(ok)
}
func getParams(kpass string) (map[string]string, error) {
	var p map[string]string
	byteP, err := authority.Authenticate(kpass)
	if err == nil {
		err = json.Unmarshal(byteP, &p)
		return p, err
	}
	return p, err
}

func setup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html")
	if r.FormValue("APIURL") != "" {
		p := map[string]string{
			"APIURL":   r.FormValue("APIURL"),
			"LABEL":    r.FormValue("LABEL"),
			"EMAIL":    r.FormValue("EMAIL"),
			"USERNAME": r.FormValue("USERNAME"),
			"MPDPORT":  r.FormValue("MPDPORT"),
			"MPDHOST":  r.FormValue("MPDHOST"),
			"MPDPASS":  r.FormValue("MPDPASS"),
		}
		cURL := r.FormValue("GUIURL") + "/?"
		if p["APIURL"] != "" {
			cURL += "&APIURL=" + p["APIURL"]
		}
		rURL := cURL
		cURL += "&KPASS="
		rURL += "&RPASS="
		byteP, ror := json.Marshal(p)
		er(ror)
		kpass, rpass := authority.Authorize(byteP)
		cURL += kpass
		rURL += rpass
		u := map[string]string{
			"controlURL": cURL,
			"resetURL":   rURL,
		}
		t, ror := template.ParseFiles("templates/authorize.html")
		er(ror)
		t.Execute(w, u)
	} else {
		p := map[string]string{
			"dummy": "dummy",
		}
		t, ror := template.ParseFiles("templates/authority.html")
		er(ror)
		t.Execute(w, p)
	}
}

func er(ror error) {
	if ror != nil {
		log.Fatalln(ror)
	}
	/*	if ror != nil {
		log.Println(ror)
	}*/
}
