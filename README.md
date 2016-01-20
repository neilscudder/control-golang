# control-golang
A web client for mpd, the music player daemon, (https://github.com/MaxKellermann/MPD). 

Run it with:
go run main.go

Usage:
Navigate to http://localhost:8080/authority to configure an instance of mpd. Two links will be generated, one to the control interface, and another to re-generate both URLs with new codes, (re-generator not yet implemented).

Control may be installed on the same host as mpd, on the same network, or on a remote web server for network-independant control.

Purpose:
Fast sharing and re-generating of password-less controls over music in a lightweight, highly compatible mobile web interface is intended for use by staff controlling music at the local branch of a chain of restaurants.

Status: ALPHA
