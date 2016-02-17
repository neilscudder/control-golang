package mpdcacher

import (
	"fmt"
	"github.com/neilscudder/gompd/mpd"
	"log"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"time"
)

// Status is for compiling the status html template.
// Holds information on the currrent song and state of mpd.
type Status struct {
	Timestamp int64
	Title     string
	YouTube   string
	Info      map[int]map[string]string
	List      []NowList
}

// NowList holds items for the tracklist surrounding the current track.
type NowList struct {
	Current bool
	Label   string
	Artist  string
	Album   string
}

// State of buttons and banner text per playnode.
type State struct {
	Timestamp              int64
	Random, Repeat, Volume int
	Play                   string
	Banner                 string
}

// MpdPlay replaces the playlist with target and starts playback
func MpdPlay(params map[string]string, target string) error {
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()
	conn.Clear()
	conn.Add(target)
	return conn.Play(0)
}

var statusBuffer = make(map[string]Status)
var stateBuffer = make(map[string]State)

type ByArtist []mpd.Attrs

func (this ByArtist) Len() int {
	return len(this)
}
func (this ByArtist) Less(i, j int) bool {
	return this[i]["Artist"] < this[j]["Artist"]
}
func (this ByArtist) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type ByAlbum []mpd.Attrs

func (this ByAlbum) Len() int {
	return len(this)
}
func (this ByAlbum) Less(i, j int) bool {
	return this[i]["Album"] < this[j]["Album"]
}
func (this ByAlbum) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type ByTitle []mpd.Attrs

func (this ByTitle) Len() int {
	return len(this)
}
func (this ByTitle) Less(i, j int) bool {
	return this[i]["Title"] < this[j]["Title"]
}
func (this ByTitle) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

type ByTrack []mpd.Attrs

func (this ByTrack) Len() int {
	return len(this)
}
func (this ByTrack) Less(i, j int) bool {
	return this[i]["file"] < this[j]["file"]
}
func (this ByTrack) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func Search(query string, params map[string]string) []mpd.Attrs {
	conn, ror := mpdConnect(params)
	er(ror)
	defer conn.Close()
	results, ror := conn.Search(query)
	er(ror)

	sort.Sort(ByArtist(results))
	sort.Sort(ByAlbum(results))

	// results is an array of maps like this:
	// map[Time:347
	// Album:Blackstar
	// Track:7
	// Last-Modified:2016-01-11T23:55:59Z
	// Title:I Can't Give Everything Away
	// Artist:David Bowie
	// Disc:1
	// Date:2016
	// file:REDACTED]

	fmt.Println(results[0])

	var tracksByAlbum = make([][]mpd.Attrs, 100)

	for range results {
		start := 0
		end := 0
		c := 0
		a := ""
		for i := end; i < len(results); i++ {
			if a == results[i]["Album"] {
				continue
			} else if a == "" {
				start = i
				a = results[i]["Album"]
			} else {
				a = results[i]["Album"]
				end = i
				tracksByAlbum[c] = results[start:end]
				c++
				start = end
			}
		}
	}
	for _, album := range tracksByAlbum {
		sort.Sort(ByTrack(album))
	}

	// All that work modified the original data, so
	return results
}

// MpdState returns a map of data for button states and banner text.
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

	_, bufExists := stateBuffer[playnode]

	if cPlay == "pause" && cmd != "play" {
		if bufExists {
			s = stateBuffer[playnode]
		} else {
			s = stateBuffer[playnode]
			s.Random = cRnd
			s.Repeat = cRpt
			s.Volume = cVol
			s.Play = cPlay
			stateBuffer[playnode] = s
		}
		return s
	}

	switch cmd {
	case "fw":
		conn.Next()
		uLog = username + " skipped forward"
	case "bk":
		conn.Previous()
		uLog = username + " skipped back"
	case "up":
		if cVol <= 90 {
			for i := 0; i < 3; i++ {
				cVol = cVol + 3
				conn.SetVolume(cVol)
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
		if cVol >= 10 {
			for i := 0; i < 3; i++ {
				cVol = cVol - 3
				conn.SetVolume(cVol)
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

	s = stateBuffer[playnode]
	if uLog != "" {
		userLog(playnode, uLog)
		s.Banner = uLog
	}
	s.Random = cRnd
	s.Repeat = cRpt
	s.Volume = cVol
	s.Play = cPlay
	stateBuffer[playnode] = s

	return s
}

// MpdStatus returns a map of data for html template.
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
			//getInfo(conn, &s)
			getListing(conn, &s)
			statusBuffer[playnode] = s
		} else {
			s = statusBuffer[playnode]
		}
	} else {
		t := time.Now()
		s.Timestamp = t.Unix()
		//getInfo(conn, &s)
		getListing(conn, &s)
		statusBuffer[playnode] = s
	}
	return s
}

// userLog stores the newest entry of user activity.
// One log file per playnode.
func userLog(playnode, details string) {
	filename := "data/" + "log." + playnode
	f, ror := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	er(ror)
	defer f.Close()

	log.SetOutput(f)
	log.Println(details)
}

// getInfo formats details for the current track
func getInfo(conn *mpd.Client, s *Status) {
	song, _ := conn.CurrentSong()
	s.Title = song["Title"]
	filename := path.Base(song["file"])
	directory := path.Dir(song["file"])
	var searchParams string

	if song["Title"] != "" {
		s.Info = map[int]map[string]string{
			1: {
				"Artist": song["Artist"],
			},
			2: {
				"Album": song["Album"] + " (" + song["Date"] + ")",
			},
		}
		searchParams = song["Artist"] + " music " + song["Title"]
	} else {
		s.Info = map[int]map[string]string{
			1: {
				"Folder": directory,
			},
		}
		searchParams = filename
	}
	encQuery := url.QueryEscape(searchParams)
	tldr := "https://www.youtube.com/embed?fs=0&controls=0&listType=search&list="
	s.YouTube = tldr + encQuery
}

// getListing stores other tracks from same folder as current track
func getListing(conn *mpd.Client, s *Status) {
	song, _ := conn.CurrentSong()
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
		if f == "." || f == "" {
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
