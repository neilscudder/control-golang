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

var statusBuffer = make(map[string]Status)
var stateBuffer = make(map[string]State)

// Params stores mpd connection settings
type Params map[string]string

// State of buttons and banner text per playnode.
type State struct {
	Timestamp              int64
	Random, Repeat, Volume int
	Play                   string
	Banner                 string
}

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

// SearchResults holds the latest search results.
type SearchResults struct {
	Results []mpd.Attrs
	Files   []string
}

// Play replaces the playlist with targets and starts playback at index
func Play(p Params, targets []string, index int) error {
	//	fmt.Println(target)
	conn, ror := mpdConnect(p)
	er(ror)
	defer conn.Close()
	conn.Clear()
	counter := 0
	for _, target := range targets {
		if target != "" {
			conn.Add(target)
			counter++
			//	fmt.Println(target)
		}
	}
	fmt.Println("Added ", counter)
	return conn.Play(index)
}

func Search(query string, p Params) SearchResults {
	conn, ror := mpdConnect(p)
	er(ror)
	defer conn.Close()
	var s SearchResults
	results, ror := conn.Search(query)
	er(ror)

	sort.Sort(ByArtist(results))
	sort.Sort(ByAlbum(results))

	var tracksByAlbum = make([][]mpd.Attrs, 100)
	var files = make([]string, len(results))

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
	counter := 0
	for i := 0; i < len(tracksByAlbum); i++ {
		for j := 0; j < len(tracksByAlbum[i]); j++ {
			files[counter] = tracksByAlbum[i][j]["file"]
			counter++
		}
	}
	s.Files = files
	// All that work modified the original data, so
	s.Results = results
	return s
}

// Command returns a map of data for button states and banner text.
// It executes a command simultaneously.
// mpd connection parameters must be supplied.
func Command(cmd string, p Params) State {
	conn, ror := mpdConnect(p)
	er(ror)
	defer conn.Close()

	var s State
	var uLog string

	username := p["USERNAME"]
	playnode := p["LABEL"]

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

// Info returns current track info for html template.
// mpd connection parameters must be supplied.
func Info(cmd string, p Params) Status {
	conn, ror := mpdConnect(p)
	er(ror)
	defer conn.Close()

	var s Status
	playnode := p["LABEL"]
	_, bufExists := statusBuffer[playnode]

	if bufExists {
		b := statusBuffer[playnode]
		t := time.Now()
		n := t.Unix()
		age := n - b.Timestamp
		if age >= 1 {
			s.Timestamp = n
			//getInfo(conn, &s)
			// getListing(conn, &s)
			getPlaylist(conn, &s)
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

// getPlaylist stores other tracks from current playlist
func getPlaylist(conn *mpd.Client, s *Status) {
	song, _ := conn.CurrentSong()
	status, _ := conn.Status()
	filename := path.Base(song["file"])
	curPos, _ := strconv.Atoi(status["song"])
	first := curPos - 20
	last := curPos + 20
	playlist, _ := conn.PlaylistInfo(first, last)
	var listing = make([]NowList, len(playlist))

	for i := 0; i < len(playlist); i++ {
		m := playlist[i]
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

func mpdConnect(p Params) (*mpd.Client, error) {
	host := p["MPDHOST"] + ":" + p["MPDPORT"]
	pass := p["MPDPASS"]
	return mpd.DialAuthenticated("tcp", host, pass)
}

// All these are for sorting search results:
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
func er(ror error) {
	if ror != nil {
		log.Println(ror)
	}
}
