package mpdcacher

import (
	"github.com/fhs/gompd/mpd"
	"log"
	"path"
	"strconv"
)

type Status struct {
	Info, Deets, Title map[int]map[string]string
}

func mpdConnect(p map[string]string) (*mpd.Client, error) {
	host := p["MPDHOST"] + ":" + p["MPDPORT"]
	pass := p["MPDPASS"]
	return mpd.DialAuthenticated("tcp", host, pass)
}

// MpdStatus returns a map of data for insertio to template for presentation
// It optionally executes a command simultaneously.
// mpd connection parameters must be supplied.
func MpdStatus(cmd string, params map[string]string) Status {
	var s Status
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()
	status, ror := conn.Status()
	er(ror)
	switch cmd {
	case "fw":
		ror := conn.Next()
		er(ror)
	case "up":
		current, ror := strconv.Atoi(status["volume"])
		er(ror)
		if current <= 95 {
			current = current + 5
			ror = conn.SetVolume(current)
			er(ror)
		}
		s.Deets = map[int]map[string]string{
			1: {
				"Volume": strconv.Itoa(current),
			},
		}
	case "dn":
		current, ror := strconv.Atoi(status["volume"])
		er(ror)
		if current >= 5 {
			current = current - 5
			ror = conn.SetVolume(current)
			er(ror)
		}
		s.Deets = map[int]map[string]string{
			1: {
				"Volume": strconv.Itoa(current),
			},
		}
	case "random":
		current, ror := strconv.Atoi(status["random"])
		er(ror)
		if current == 1 {
			current = 0
			ror = conn.Random(false)
			er(ror)

		} else {
			current = 1
			ror = conn.Random(true)
			er(ror)
		}
		s.Deets = map[int]map[string]string{
			1: {
				"Random": strconv.Itoa(current),
			},
		}
	}
	song, ror := conn.CurrentSong()
	er(ror)
	getInfo(conn, &s)
	s.Title = map[int]map[string]string{
		1: {
			"Title": song["Title"],
		},
	}
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

func er(ror error) {
	if ror != nil {
		log.Println(ror)
	}
}
