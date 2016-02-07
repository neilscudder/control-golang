# control-golang
A web client for mpd, the music player daemon, (https://github.com/MaxKellermann/MPD). Features a custom authentication framework to allow sharing of controls without passwords.

Alpha status.

## Purpose:
Fast sharing and re-generating of password-less controls over music in a lightweight, highly compatible mobile web interface.

## Usage:
Navigate to /authority to configure a control for an existing instance of mpd. Two links will be generated, one to the control interface, and another to re-generate both URLs with new codes.

Invoke the ssl server by passing -pem and -key with paths to the respective files.

To use ports below 1024 on linux, you must use cli utility setcap to grant permission, as outlined here: http://stackoverflow.com/a/14573592

To use priveleged ports on freebsd: http://crossbar.io/docs/Running-on-privileged-ports/
