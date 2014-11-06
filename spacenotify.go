package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/guelfey/go.dbus"
)

var API = "https://api.openlab-augsburg.de/13"

func callAPI() (bool, time.Time, error) {

	resp, err := http.Get(API)
	if err != nil {
		return false, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, time.Time{}, errors.New("Failed calling SpaceAPI: " + resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)

	var result struct {
		State struct {
			Open       interface{}
			LastChange int64
		}
	}

	err = decoder.Decode(&result)

	switch result.State.Open.(type) {
	case bool:
		return result.State.Open.(bool), time.Unix(result.State.LastChange, 0), nil
	}

	return false, time.Time{}, errors.New("State currently unknown.")
}

func dbusNotify(conn *dbus.Conn, title, message, imagePath string) error {
	// see https://developer.gnome.org/notification-spec/
	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")
	call := obj.Call("org.freedesktop.Notifications.Notify", 0,
		"SpaceNotify",       // Application name
		uint32(0),           // Replaces ID
		"file://"+imagePath, // Notification Icon
		title,                     // Summary
		message,                   // Body
		[]string{},                // Actions
		map[string]dbus.Variant{}, // Hints
		int32(5000),               // Expiration Timeout
	)

	return call.Err
}

func genNotification(state bool, lastChange time.Time, err error) (message, imagePath string) {

	//FIXME send image-data instead of file path.
	dir, perr := filepath.Abs(filepath.Dir(os.Args[0]))
	if perr != nil {
		panic(err)
	}
	dir = filepath.Join(dir, "icons")

	if err != nil {
		return err.Error(),
			filepath.Join(dir, "unknown.png")
	}
	if state {
		return "OpenLab is open since " + lastChange.Format("Monday, 15:04"),
			filepath.Join(dir, "open.png")
	}

	return "OpenLab is closed since " + lastChange.Format("Monday, 15:04"),
		filepath.Join(dir, "closed.png")
}

func main() {

	watch := flag.Bool("watch", false, "keep running and call SpaceAPI periodically.")
	freq := flag.Duration("frequency", 5*time.Minute, "set a frequency at which the API is called.")
	flag.Parse()

	conn, err := dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	state, lastChange, err := callAPI()

	if !(*watch) {
		message, imagePath := genNotification(state, lastChange, err)
		dbusNotify(conn, "Space State", message, imagePath)
		return
	}

	lastState := state
	var lastErr error

	for {
		state, lastChange, err := callAPI()
		if state != lastState || (lastErr == nil && err != nil) || (lastErr != nil && err == nil) || (err != nil && err.Error() != lastErr.Error()) {
			lastState = state
			lastErr = err
			message, imagePath := genNotification(state, lastChange, err)
			dbusNotify(conn, "Space State", message, imagePath)
		}
		time.Sleep(*freq)
	}
}
