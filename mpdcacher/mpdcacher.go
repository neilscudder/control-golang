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
	Title string
	Deets map[string]string
	Info  map[int]map[string]string
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
	curvol, _ := strconv.Atoi(status["volume"])
	currnd, _ := strconv.Atoi(status["random"])
	switch cmd {
	case "fw":
		vol := curvol
		for vol >= 5 {
			vol = vol - 5
			conn.SetVolume(vol)
			time.Sleep(20 * time.Millisecond)
		}
		conn.Next()
		conn.SetVolume(curvol)
	case "up":
		if curvol <= 90 {
			for i := 0; i < 5; i++ {
				curvol = curvol + 2
				conn.SetVolume(curvol)
				time.Sleep(20 * time.Millisecond)
			}
		}
	case "dn":
		if curvol >= 10 {
			for i := 0; i < 5; i++ {
				curvol = curvol - 2
				conn.SetVolume(curvol)
				time.Sleep(20 * time.Millisecond)
			}
		}
	case "random":
		if currnd == 1 {
			currnd = 0
			ror = conn.Random(false)
			er(ror)

		} else {
			currnd = 1
			ror = conn.Random(true)
			er(ror)
		}
	}
	s.Deets = map[string]string{
		"CurrentRandom": strconv.Itoa(currnd),
		"Volume":        strconv.Itoa(curvol),
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
