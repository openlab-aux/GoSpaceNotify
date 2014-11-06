SpaceNotify
===========

a simple go program that can watch a SpaceAPI instance for state changes and
displays it via DBus notification.

![screenshot](misc/screenshot_gnome.png)

## Building
cd into the directory. Then type
```bash
go get github.com/guelfey/go.dbus
```
to install go.dbus. To build SpaceNotify type

```bash
go build spacenotify.go
```


## Usage
Type
```bash
./spacenotify
```
to run it in "one shot" mode. This immidiately calls the API and displays the
returned state and lastchange via DBus. The process exits afterwards. In 
addition you can run it in watch mode with
```bash
./spacenotify --watch
```
SpaceNotify will then call the API periodically to watch for changes. If the
state has changed since the last call a notification is displayed via DBus. You
can change the frequency (default 5m) to anything with 
```bash
./spacenotify --watch --frequency 1337s
```

