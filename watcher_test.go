package autotest

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	os.MkdirAll("testdata", 0755)
	var folderLock sync.Mutex
	folderCount := make(map[string]int16)
	printCount := 0
	folderChanged := func(folder string) *TestResult {
		folderLock.Lock()
		folderCount[folder] = folderCount[folder] + 1
		folderLock.Unlock()
		return nil
	}
	print := func(res *TestResult) {
		printCount++
	}

	go Watch("testdata", folderChanged, print)
	for i := 0; i < 100; i++ {
		path := path.Join("testdata", fmt.Sprintf("test%d.go", i))
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
		time.Sleep(1 * time.Millisecond)
	}
	os.Rename("testdata/test1.go", "testdata/test1a.go")
	time.Sleep(1000 * time.Millisecond)
	absFolder, _ := filepath.Abs("testdata")
	if folderCount[absFolder] != 1 {
		t.Error("Watch didn't work", folderCount[absFolder])
	}
	os.RemoveAll("testdata")
}
