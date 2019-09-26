package autotest

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

// Watch begins polling for changes in the specified folder and runs fileProcessor each time a file is changed
func Watch(folder string, fileProcessor func(string) error) error {
	fileReadyToProcess := make(chan string)
	debouncedChange := debounceChange(800*time.Millisecond, fileReadyToProcess)

	w, err := getWatcher(folder)
	if err != nil {
		return err
	}

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		debouncedChange(watcher.Event{FileInfo: info, Op: watcher.Create, Path: path})
		return nil
	})
	go func() {
		for {
			select {
			case event := <-w.Event:
				debouncedChange(event)
			case filename := <-fileReadyToProcess:
				if err := fileProcessor(filename); err != nil {
					log.Println(err)
				}
			case <-w.Closed:
				return
			}
		}
	}()
	fmt.Printf("Monitoring folder %s\n", folder)
	return w.Start(250 * time.Millisecond)
}

func getWatcher(folder string) (*watcher.Watcher, error) {
	w := watcher.New()
	if err := w.AddRecursive(folder); err != nil {
		return nil, fmt.Errorf("Error adding watcher directory - %v", err)
	}
	w.FilterOps(watcher.Create, watcher.Write, watcher.Move, watcher.Rename)
	return w, nil
}

// debounceChange marks a file as ready to process after remaining unchanged for a given duration.
func debounceChange(interval time.Duration, readyToProcess chan string) func(e watcher.Event) {
	timer := make(map[string]*time.Timer)
	var tMutex sync.Mutex

	return func(e watcher.Event) {
		folder := getValidatedGoFolder(&e)
		if folder == "" {
			return
		}
		tMutex.Lock()
		t, ok := timer[folder]
		if !ok || t == nil {
			t = time.NewTimer(interval)
			timer[folder] = t
			tMutex.Unlock()

			go func(dt *time.Timer) {
				<-dt.C
				dt.Stop()

				tMutex.Lock()
				delete(timer, folder)
				tMutex.Unlock()
				if _, err := os.Stat(folder); os.IsNotExist(err) {
					return // file has been deleted since we started the timer, so ignore
				}
				readyToProcess <- folder
			}(t)
		} else {
			tMutex.Unlock()
			t.Reset(interval)
		}
		return
	}
}

func getValidatedGoFolder(e *watcher.Event) string {
	filename := getWatcherPath(e)
	return getGoFolder(filename)
}

func getWatcherPath(e *watcher.Event) string {
	filename := e.Path
	toPathIndex := strings.Index(filename, "-> ") // watcher sends e.Path in the format of fromPath -> toPath for Move and Rename events
	if toPathIndex != -1 {
		filename = filename[toPathIndex+3:]
	}
	filename, _ = filepath.Abs(filename)
	return filename
}

func getGoFolder(filename string) string {
	folder := filepath.Dir(filename)
	if filepath.Ext(filename) == ".go" {
		return folder
	}
	return ""
}
