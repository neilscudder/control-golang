package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	gz "github.com/NYTimes/gziphandler"
	"github.com/neilscudder/control-golang/authority"
	m "github.com/neilscudder/control-golang/mpdcacher"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

var searchBuffer = make(map[string][]string)

func main() {
	resGz := gz.GzipHandler(http.FileServer(http.Dir("res/")))
	setupGz := gz.GzipHandler(http.HandlerFunc(setup))
	postGz := gz.GzipHandler(http.HandlerFunc(post))
	guiGz := gz.GzipHandler(http.HandlerFunc(gui))

	http.Handle("/res/", resGz)
	http.Handle("/authority", setupGz)
	http.Handle("/authorize", setupGz)
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
	var p = make(m.Params)
	kpass := r.FormValue("KPASS")
	p, err := getParams(kpass)
	p["KPASS"] = kpass
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, err.Error())
		return
	}

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html")
	switch r.FormValue("a") {
	case "search":
		query := "any \""
		query += r.FormValue("search")
		query += "\""
		s := m.Search(query, p)
		searchBuffer[kpass] = s.Files
		t, ror := template.ParseFiles("templates/search.html")
		er(ror)
		t.Execute(w, s)
	case "browser":
		t, ror := template.ParseGlob("templates/browser/*")
		er(ror)
		t.ExecuteTemplate(w, "GUI", p)
	case "command":
		cmd := r.FormValue("b")
		if cmd == "info" {
			u := m.MpdStatus(cmd, p)
			t, ror := template.ParseFiles("templates/status.html")
			er(ror)
			t.Execute(w, u)
		} else {
			s := m.MpdState(cmd, p)
			state, _ := json.Marshal(s)
			w.Header().Set("Content-Type", "application/json")
			w.Write(state)
		}
	default:
		t, ror := template.ParseGlob("templates/gui/*")
		er(ror)
		t.ExecuteTemplate(w, "GUI", p)
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
	index, _ := strconv.Atoi(r.FormValue("c"))
	targets := searchBuffer[kpass]
	w.Header().Set("Content-Type", "text/html")
	m.MpdPlay(p, targets, index)
	//fmt.Println(targets)
	ok := []byte("ok")
	w.Write(ok)
}
func getParams(kpass string) (m.Params, error) {
	var p m.Params
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
		p := m.Params{
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
		u := map[string]string{
			"dummy": "dummy",
		}
		t, ror := template.ParseFiles("templates/authority.html")
		er(ror)
		t.Execute(w, u)
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
