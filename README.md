# control-golang
A web client for mpd, the music player daemon, (https://github.com/MaxKellermann/MPD). 

This is a proof-of-concept application.

## Purpose:
Fast sharing and re-generating of password-less controls over music in a lightweight, highly compatible mobile web interface is intended for use by staff controlling music at the local branch of a chain of restaurants.

## Usage:
Navigate to http://localhost:8080/authority to configure a control for an existing instance of mpd. Two links will be generated, one to the control interface, and another to re-generate both URLs with new codes.

Invoke the ssl server by passing -pem and -key flags.

To use ports below 1024 on linux, you must use cli utility setcap to grant permission, as outlined here: http://stackoverflow.com/a/14573592

To use priveleged ports on freebsd: http://crossbar.io/docs/Running-on-privileged-ports/
