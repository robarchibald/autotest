package autotest

import (
	"sync"
)

var trackedFolders = make(map[string]*tracking)
var folderMutex sync.RWMutex

type tracking struct {
	Original *TestResult
	Last     *TestResult
}

// Track keeps track of initial results and returns changed coverage results
func Track(test *TestResult) *TestResult {
	saved := getFolderResults(test.Folder)
	if saved == nil {
		saveFolderResults(test)
		return nil
	}
	diff := getResultDiff(saved, test)
	saved.Last = test
	saveTracking(saved)
	return diff
}

func getFolderResults(folder string) *tracking {
	folderMutex.RLock()
	saved := trackedFolders[folder]
	folderMutex.RUnlock()
	return saved
}

func saveTracking(v *tracking) {
	folderMutex.Lock()
	trackedFolders[v.Original.Folder] = v
	folderMutex.Unlock()
}

func saveFolderResults(test *TestResult) {
	saveTracking(&tracking{Original: test, Last: test})
}

func getResultDiff(v *tracking, current *TestResult) *TestResult {
	if current.Error != nil {
		return current
	}

	return &TestResult{
		Folder:   current.Folder,
		Status:   current.Status,
		Coverage: getCoverageDiff(v.Original.Coverage, current.Coverage),
	}
}

func getCoverageDiff(first, current []FunctionCoverage) []FunctionCoverage {
	coverageMap := make(map[string]float32)
	for _, item := range first {
		coverageMap[item.Function] = item.CoveragePercent
	}
	differentCoverage := []FunctionCoverage{}
	for _, item := range current {
		if coveragePercent, ok := coverageMap[item.Function]; !ok || coveragePercent != item.CoveragePercent {
			differentCoverage = append(differentCoverage, item)
		}
	}
	return differentCoverage
}
