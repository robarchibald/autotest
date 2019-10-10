package autotest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	os.MkdirAll("testdata", 0755)
	ioutil.WriteFile("testdata/test0.go", nil, 0644) // create go file so it'll be viewed as a go folder
	var folderLock sync.Mutex
	folderCount := make(map[string]int16)
	folderChanged := func(folder string) *TestResult {
		folderLock.Lock()
		folderCount[folder] = folderCount[folder] + 1
		folderLock.Unlock()
		return nil
	}
	track := func(res *TestResult) *TestResult { return nil }
	print := func(res *TestResult) {}

	go Watch("testdata", folderChanged, track, print)
	time.Sleep(1 * time.Millisecond)
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
	if folderCount[absFolder] != 2 { // once for initial create of Watch. Once for file change
		t.Error("Watch didn't work", folderCount[absFolder])
	}
	os.RemoveAll("testdata")
}
