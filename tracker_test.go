package autotest

import "testing"

func TestTrack(t *testing.T) {

}

func TestGetAndSetFolderResults(t *testing.T) {
	v := getFolderResults("hello")
	if v != nil {
		t.Error("Expected nil")
	}
	r := &TestResult{Folder: "hello"}
	saveFolderResults(r)
	v = getFolderResults("hello")
	if v.Original != r || v.Last != r {
		t.Error("Expected save to work right", v)
	}
}
func TestGetResultDiff(t *testing.T) {
	getResultDiff(&tracking{Original: &TestResult{}, Last: &TestResult{}}, &TestResult{})
}

func TestGetCoverageDiff(t *testing.T) {
	if diff := getCoverageDiff([]FunctionCoverage{{Function: "1", CoveragePercent: 25}}, []FunctionCoverage{{Function: "1", CoveragePercent: 30}}); len(diff) != 1 || diff[0].CoveragePercent != 30 {
		t.Error("Expected correct diff", diff)
	}
	if diff := getCoverageDiff([]FunctionCoverage{}, []FunctionCoverage{{Function: "1", CoveragePercent: 0}}); len(diff) != 1 || diff[0].CoveragePercent != 0 {
		t.Error("Expected correct diff", diff)
	}
}
