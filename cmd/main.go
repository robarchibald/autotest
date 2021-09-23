package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/robarchibald/autotest"
	"github.com/robarchibald/gobounce"
)

func main() {
	w, err := gobounce.New(gobounce.Options{RootFolders: []string{"."}, FolderExclusions: []string{"node_modules"}, FollowNewFolders: true}, 20*time.Millisecond)
	if err != nil {
		panic(err)
	}

	watchFolders := w.WatchFolders()
	tmpDir, err := setupTempDir(watchFolders)
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	go handleChanges(w, tmpDir)

	for _, folder := range watchFolders {
		w.FolderChanged <- folder
	}

	w.Start()
}

func setupTempDir(watchFolders []string) (string, error) {
	tmpDir := filepath.Join(os.TempDir(), "autotest-"+time.Now().Format("20060102-150405"))
	if err := os.Mkdir(tmpDir, 0755); err != nil {
		return "", err
	}
	for _, watchFolder := range watchFolders {
		tmpWatchFolder := filepath.Join(tmpDir, watchFolder)
		fmt.Println(tmpWatchFolder)
	}
	return tmpDir, nil
}

func handleChanges(w *gobounce.Filewatcher, tempDir string) {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	testsToTrack := make(chan *autotest.TestResult, 100) // track tests in parallel as they come in
	testsToPrint := make(chan *autotest.TestResult)      // print one at a time

	for {
		select {
		case <-w.FileChanged:
		case folder := <-w.FolderChanged:
			go func() {
				fmt.Println("\nrunning tests for", folder)
				testsToTrack <- autotest.RunTests(folder, tempDir)
			}()
		case <-w.Closed:
			return
		case <-w.Error:
		case track := <-testsToTrack:
			go func() {
				if print := autotest.Track(track); print != nil {
					testsToPrint <- print
				} else {
					fmt.Println("unchanged")
				}
			}()
		case print := <-testsToPrint:
			autotest.PrintTest(print)
		case <-term:
			w.Close()
			return
		}
	}
}
