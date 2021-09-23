package autotest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	diff := getCoverageDiff(
		[]FunctionCoverage{{Function: "1", CoveragePercent: 25}},
		[]FunctionCoverage{{Function: "1", CoveragePercent: 30}})
	require.Equal(t, 1, len(diff))
	assert.Equal(t, float32(30.0), diff[0].CoveragePercent)

	diff = getCoverageDiff(
		[]FunctionCoverage{},
		[]FunctionCoverage{{Function: "1", CoveragePercent: 0}})
	require.Equal(t, 1, len(diff))
	assert.Equal(t, 0, diff[0].CoveragePercent)
}
