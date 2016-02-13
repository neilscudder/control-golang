package mpdcacher

import (
	//"fmt"
	"github.com/fhs/gompd/mpd"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

// Status is for compiling the status html template
// Holds information on the currrent song and state of mpd
type Status struct {
	Timestamp int64
	Title     string
	YouTube   string
	Info      map[int]map[string]string
	List      []NowList
}

// State stores button states and banner text per playnode
type State struct {
	Timestamp              int64
	Random, Repeat, Volume int
	Play                   string
	Banner                 string
}

type NowList struct {
	Current bool
	Label   string
	Artist  string
	Album   string
}

var statusBuffer = make(map[string]Status)
var stateBuffer = make(map[string]State)

// MpdState returns a map of data for button states
// It executes a command simultaneously.
// mpd connection parameters must be supplied.
func MpdState(cmd string, params map[string]string) State {
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()

	var s State
	var uLog string

	username := params["USERNAME"]
	playnode := params["LABEL"]

	status, _ := conn.Status()
	cVol, _ := strconv.Atoi(status["volume"])
	cRnd, _ := strconv.Atoi(status["random"])
	cRpt, _ := strconv.Atoi(status["repeat"])
	cPlay, _ := status["state"]
	switch cmd {
	case "fw":
		if cPlay == "pause" {
			break
		}
		conn.Next()
		uLog = username + " skipped forward"
	case "bk":
		if cPlay == "pause" {
			break
		}
		conn.Previous()
		uLog = username + " skipped back"
	case "up":
		if cPlay == "pause" {
			break
		}
		if cVol <= 90 {
			for i := 0; i < 5; i++ {
				cVol = cVol + 2
				conn.SetVolume(cVol)
				time.Sleep(20 * time.Millisecond)
			}
			uLog = username + " raised volume to " + strconv.Itoa(cVol)
		} else if cVol != 100 {
			cVol = 100
			conn.SetVolume(cVol)
			uLog = username + " raised volume to " + strconv.Itoa(cVol)
		} else {
			uLog = "Volume at max"
		}
	case "dn":
		if cPlay == "pause" {
			break
		}
		if cVol >= 10 {
			for i := 0; i < 5; i++ {
				cVol = cVol - 2
				conn.SetVolume(cVol)
				time.Sleep(20 * time.Millisecond)
			}
			uLog = username + " lowered volume to " + strconv.Itoa(cVol)
		} else if cVol != 0 {
			cVol = 0
			conn.SetVolume(cVol)
			uLog = username + " lowered volume to " + strconv.Itoa(cVol)
		} else {
			uLog = "Volume at min"
		}
	case "repeat":
		if cPlay == "pause" {
			break
		}
		if cRpt == 1 {
			cRpt = 0
			conn.Repeat(false)
			uLog = username + " disabled repeat"
		} else {
			cRpt = 1
			conn.Repeat(true)
			uLog = username + " enabled repeat"
		}
	case "random":
		if cPlay == "pause" {
			break
		}
		if cRnd == 1 {
			cRnd = 0
			conn.Random(false)
			uLog = username + " disabled random"
		} else {
			cRnd = 1
			conn.Random(true)
			uLog = username + " enabled random"
		}
	case "play":
		if cPlay == "play" {
			conn.Pause(true)
			cPlay = "pause"
			uLog = username + " paused playback"
		} else if cPlay == "pause" {
			conn.Pause(false)
			cPlay = "play"
			uLog = username + " resumed playback"
		}
	}
	_, bufExists := stateBuffer[playnode]
	if bufExists && cmd == "state" {
		s = stateBuffer[playnode]
	} else {
		s = stateBuffer[playnode]
		if uLog != "" {
			userLog(playnode, uLog)
			s.Banner = uLog
		}
		t := time.Now()
		s.Timestamp = t.Unix()
		s.Random = cRnd
		s.Repeat = cRpt
		s.Volume = cVol
		s.Play = cPlay
		stateBuffer[playnode] = s
	}
	return s
}

// MpdStatus returns a map of data for html template
// mpd connection parameters must be supplied.
func MpdStatus(cmd string, params map[string]string) Status {
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()

	var s Status

	playnode := params["LABEL"]

	_, bufExists := statusBuffer[playnode]
	if bufExists {
		b := statusBuffer[playnode]
		t := time.Now()
		n := t.Unix()
		age := n - b.Timestamp
		if age >= 1 {
			s.Timestamp = n
			getInfo(conn, &s)
			statusBuffer[playnode] = s
		} else {
			s = statusBuffer[playnode]
		}
	} else {
		t := time.Now()
		s.Timestamp = t.Unix()
		getInfo(conn, &s)
		statusBuffer[playnode] = s
	}
	return s
}

// userLog returns the newest entry of user activity as a string
// if parameter is not nil, add new entry then return it
// one log file per playnode
func userLog(playnode, details string) {
	filename := "data/" + "log." + playnode
	f, ror := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	er(ror)
	defer f.Close()
	log.SetOutput(f)
	log.Println(details)
}

func getInfo(conn *mpd.Client, s *Status) {
	status, ror := conn.Status()
	er(ror)
	song, ror := conn.CurrentSong()
	er(ror)
	s.Title = song["Title"]
	filename := path.Base(song["file"])
	directory := path.Dir(song["file"])
	thisDir, _ := conn.ListInfo(directory)
	var listing = make([]NowList, len(thisDir))
	for i := 0; i < len(thisDir); i++ {
		m := thisDir[i]
		p := m["file"]
		t := m["title"]
		d := path.Dir(p)
		f := path.Base(p)
		if f == "." {
			continue
		}
		if t != "" {
			if f == filename {
				listing[i].Current = true
				listing[i].Album = m["album"]
			}
			listing[i].Artist = m["artist"]
			listing[i].Label = t
		} else {
			if f == filename {
				listing[i].Current = true
				listing[i].Album = d
			}
			listing[i].Artist = m["artist"]
			listing[i].Label = f
		}
	}
	s.List = listing
	if song["Title"] != "" {
		s.Info = map[int]map[string]string{
			1: {
				"Artist": song["Artist"],
			},
			2: {
				"Album": song["Album"] + " (" + song["Date"] + ")",
			},
		}
		searchParams := song["Artist"] + " music " + song["Title"]
		encQuery := url.QueryEscape(searchParams)
		s.YouTube = "https://www.youtube.com/embed?fs=0&controls=0&listType=search&list=" + encQuery
		//		fmt.Println(encQuery)
	} else if status["state"] == "play" {
		s.Info = map[int]map[string]string{
			1: {
				"Folder": directory,
			},
		}
		searchParams := filename
		encQuery := url.QueryEscape(searchParams)
		s.YouTube = "https://www.youtube.com/embed?fs=0&controls=0&listType=search&list=" + encQuery
	} else {
		s.Info = map[int]map[string]string{
			1: {
				"State": status["state"],
			},
		}
	}
}

func mpdConnect(p map[string]string) (*mpd.Client, error) {
	host := p["MPDHOST"] + ":" + p["MPDPORT"]
	pass := p["MPDPASS"]
	return mpd.DialAuthenticated("tcp", host, pass)
}

func er(ror error) {
	if ror != nil {
		log.Println(ror)
	}
}
