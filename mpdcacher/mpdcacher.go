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

// Status holds information on the currrent song.
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
	cmdList := conn.BeginCommandList()
	cmdList.Clear()
	counter := 0
	for _, target := range targets {
		if target != "" {
			cmdList.Add(target)
			counter++
			//	fmt.Println(target)
		}
	}
	fmt.Println("Added ", counter)
	cmdList.Play(index)
	return cmdList.End()
}

// Search performs a case insensitive substring search on the mpd database
// mpd connection parameters must be supplied.
func Search(query string, qType string, p Params) SearchResults {
	conn, ror := mpdConnect(p)
	er(ror)
	defer conn.Close()
	var s SearchResults
	qType += " \""
	qType += query
	qType += "\""
	results, ror := conn.Search(qType)
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

	if cPlay == "pause" && cmd != "play" && cmd != "dn" {
		if bufExists {
			s = stateBuffer[playnode]
		} else {
			t := time.Now()
			n := t.Unix()
			s.Timestamp = n
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
		if cVol <= 95 {
			cVol = cVol + 5
			conn.SetVolume(cVol)
			uLog = username + " raised volume to " + strconv.Itoa(cVol)
		} else if cVol != 100 {
			cVol = 100
			conn.SetVolume(cVol)
			uLog = username + " raised volume to " + strconv.Itoa(cVol)
		} else {
			uLog = "Volume at max"
		}
	case "dn":
		if cVol >= 5 {
			cVol = cVol - 5
			conn.SetVolume(cVol)
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

	if bufExists {
		s = stateBuffer[playnode]
	}
	if uLog != "" {
		userLog(playnode, uLog)
		s.Banner = uLog
	}
	t := time.Now()
	n := t.Unix()
	s.Timestamp = n
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
	var s Status
	conn, err := mpdConnect(p)
	if err != nil {
		var listing = make([]NowList, 1)
		listing[0].Label = "Music player offline"
		listing[0].Current = true
		s.List = listing
		return s
	}
	defer conn.Close()

	playnode := p["LABEL"]
	_, bufExists := statusBuffer[playnode]

	if bufExists {
		b := statusBuffer[playnode]
		t := time.Now()
		n := t.Unix()
		age := n - b.Timestamp
		if age > 10 {
			fmt.Println("Timeout at age = ", age)
			s.Timestamp = n
			getPlaylist(conn, &s)
			statusBuffer[playnode] = s
		} else {
			fmt.Println("Info buffer")
			s = statusBuffer[playnode]
		}
	} else {
		fmt.Println("no buffer for ", playnode)
		go watcher(p, playnode)
		t := time.Now()
		s.Timestamp = t.Unix()
		getPlaylist(conn, &s)
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

// getPlaylist stores other tracks from current playlist
func getPlaylist(conn *mpd.Client, s *Status) {
	var first, last int
	status, ror := conn.Status()
	er(ror)
	curPos, _ := strconv.Atoi(status["song"])
	if curPos > 19 {
		first = curPos - 20
	} else {
		first = 0
	}
	last = curPos + 20
	playlist, ror := conn.PlaylistInfo(first, last)
	er(ror)

	var listing = make([]NowList, len(playlist))
	for i := 0; i < len(playlist); i++ {
		item := playlist[i]
		iPos, _ := strconv.Atoi(item["Pos"])
		filepath := item["file"]
		t := item["title"]
		d := path.Dir(filepath)
		f := path.Base(filepath)
		if f == "." || f == "" {
			continue
		}
		if t != "" {
			if curPos == iPos {
				listing[i].Current = true
				listing[i].Album = item["album"]
			}
			listing[i].Artist = item["artist"]
			listing[i].Label = t
		} else {
			if curPos == iPos {
				listing[i].Current = true
				listing[i].Album = d
			}
			listing[i].Artist = item["artist"]
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
func watcher(p Params, playnode string) {
	host := p["MPDHOST"] + ":" + p["MPDPORT"]
	pass := p["MPDPASS"]
	w, err := mpd.NewWatcher("tcp", host, pass)
	if err != nil {
		fmt.Println("Connection failed to: ", playnode)
		delete(statusBuffer, playnode)
		delete(stateBuffer, playnode)
		return
	}
	fmt.Println("New watcher for: ", playnode)
	defer w.Close()

	go func() {
		for err := range w.Error {
			fmt.Println("Error:", err)
		}
	}()

	for subsystem := range w.Event {
		fmt.Println("Changed subsystem: ", subsystem)
		fmt.Println("Reset buffer for: ", playnode)
		delete(statusBuffer, playnode)
		b := stateBuffer[playnode]
		t := time.Now()
		n := t.Unix()
		age := n - b.Timestamp
		if age >= 1 {
			delete(stateBuffer, playnode)
		}
		return
	}
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
		//log.Fatalln(ror)
	}
}

//
// UNUSED FUNCTIONS
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
