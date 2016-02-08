package mpdcacher

import (
	"github.com/fhs/gompd/mpd"
	"log"
	"path"
	"strconv"
	"time"
)

// Status is for compiling the status html template
// Holds information on the currrent song and state of mpd
type Status struct {
	Title  string
	Banner string
	Deets  map[string]string
	Info   map[int]map[string]string
}

// MpdStatus returns a map of data for html template
// It optionally executes a command simultaneously.
// mpd connection parameters must be supplied.
func MpdStatus(cmd string, params map[string]string) Status {
	var s Status
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()
	status, _ := conn.Status()
	username := params["USERNAME"]
	cVol, _ := strconv.Atoi(status["volume"])
	cRnd, _ := strconv.Atoi(status["random"])
	cRpt, _ := strconv.Atoi(status["repeat"])
	cPlay, _ := status["state"]
	switch cmd {
	case "fw":
		vol := cVol
		for vol >= 5 {
			vol = vol - 5
			conn.SetVolume(vol)
			time.Sleep(20 * time.Millisecond)
		}
		conn.Next()
		conn.SetVolume(cVol)
	case "bk":
		vol := cVol
		for vol >= 5 {
			vol = vol - 5
			conn.SetVolume(vol)
			time.Sleep(20 * time.Millisecond)
		}
		conn.Previous()
		conn.SetVolume(cVol)
	case "up":
		if cVol <= 90 {
			for i := 0; i < 5; i++ {
				cVol = cVol + 2
				conn.SetVolume(cVol)
				time.Sleep(20 * time.Millisecond)
			}
		}
	case "dn":
		if cVol >= 10 {
			for i := 0; i < 5; i++ {
				cVol = cVol - 2
				conn.SetVolume(cVol)
				time.Sleep(20 * time.Millisecond)
			}
		}
	case "repeat":
		if cRpt == 1 {
			cRpt = 0
			conn.Repeat(false)

		} else {
			cRpt = 1
			conn.Repeat(true)
		}
	case "random":
		if cRnd == 1 {
			cRnd = 0
			conn.Random(false)
		} else {
			cRnd = 1
			conn.Random(true)
		}
	case "play":
		if cPlay == "play" {
			conn.Pause(true)
		} else if cPlay == "pause" {
			conn.Pause(false)
		}
	}
	if cmd != "info" {
		s.Banner = username
	}
	s.Deets = map[string]string{
		"CurrentRandom": strconv.Itoa(cRnd),
		"Repeat":        strconv.Itoa(cRpt),
		"Volume":        strconv.Itoa(cVol),
	}
	song, ror := conn.CurrentSong()
	er(ror)
	getInfo(conn, &s)
	s.Title = song["Title"]
	return s
}

func getInfo(conn *mpd.Client, s *Status) {
	status, ror := conn.Status()
	er(ror)
	song, ror := conn.CurrentSong()
	er(ror)
	if status["state"] == "play" && song["Title"] != "" {
		s.Info = map[int]map[string]string{
			1: {
				"Artist": song["Artist"],
			},
			2: {
				"Album": song["Album"] + " (" + song["Date"] + ")",
			},
		}
	} else if status["state"] == "play" {
		filename := path.Base(song["file"])
		directory := path.Dir(song["file"])
		s.Info = map[int]map[string]string{
			1: {
				"File Name": filename,
			},
			2: {
				"Folder": directory,
			},
		}
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
