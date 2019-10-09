package autotest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/radovskyb/watcher"
)

// Watch begins polling for changes in the specified folder and runs RunTests each time a file is changed, and PrintTest each time a test is ready for printing
func Watch(folder string, fileProcessor func(string) *TestResult, printProcessor func(*TestResult)) error {
	fileReadyToProcess := make(chan string, 100) // run 100 tests at once
	testsReadyToPrint := make(chan *TestResult)  // print one at a time
	debouncedChange := debounceChange(800*time.Millisecond, fileReadyToProcess)

	folders := getGoFolders(folder)
	w, err := getWatcher(folders)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				debouncedChange(&event)
			case filename := <-fileReadyToProcess:
				go func() {
					testsReadyToPrint <- fileProcessor(filename)
				}()
			case test := <-testsReadyToPrint:
				printProcessor(test)
			case <-w.Closed:
				return
			}
		}
	}()
	for _, goFolder := range folders {
		fileReadyToProcess <- goFolder
	}
	fmt.Printf("Monitoring folder %s\n", folder)
	return w.Start(250 * time.Millisecond)
}

func getWatcher(folders []string) (*watcher.Watcher, error) {
	w := watcher.New()
	for _, folder := range folders {
		if err := w.Add(folder); err != nil {
			return nil, fmt.Errorf("Error adding watcher directory - %v", err)
		}
	}
	w.FilterOps(watcher.Create, watcher.Write, watcher.Move, watcher.Rename)
	return w, nil
}

func getGoFolders(root string) []string {
	fmt.Println("reading folders")
	goFolders := map[string]bool{}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !isGoFile(path) {
			return nil
		}
		folder, _ := filepath.Abs(path)
		goFolders[filepath.Dir(folder)] = true
		return nil
	})
	folders := []string{}
	for folder := range goFolders {
		folders = append(folders, folder)
	}
	return folders
}

// debounceChange marks a file as ready to process after remaining unchanged for a given duration.
func debounceChange(interval time.Duration, readyToProcess chan string) func(e *watcher.Event) {
	timer := make(map[string]*time.Timer)
	var tMutex sync.Mutex

	return func(e *watcher.Event) {
		folder := getWatcherPath(e.Path)
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

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}

func getWatcherPath(path string) string {
	if !isGoFile(path) {
		return ""
	}
	filename := path
	toPathIndex := strings.Index(filename, "-> ") // watcher sends e.Path in the format of fromPath -> toPath for Move and Rename events
	if toPathIndex != -1 {
		filename = filename[toPathIndex+3:]
	}
	filename, _ = filepath.Abs(filename)
	return filepath.Dir(filename)
}
