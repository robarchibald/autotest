package autotest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/EndFirstCorp/execfactory"
)

var runCoverageArgs = []string{"test", "-json", "-short", "-coverprofile", "cover.out", "-timeout", "5s"}
var getCoverageArgs = []string{"tool", "cover", "-func=cover.out"}

// TestResult contains the full results of a test run
type TestResult struct {
	Folder    string
	BuildFail []byte
	TestFail  bool
	Status    []TestStatus
	Coverage  []FunctionCoverage
}

// TestStatus contains the status for a single test run
type TestStatus struct {
	Elapsed    float64
	Package    string
	Test       string
	TestResult string
	Output     string
}

// FunctionCoverage contains the code coverage for a function
type FunctionCoverage struct {
	Filename        string
	Function        string
	CoveragePercent float32
}

type testEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

var exec = execfactory.NewOSCreator()

// RunTests will run a new set of tests whenever a file changes
func RunTests(folder string) *TestResult {
	result := &TestResult{Folder: folder}
	out, exitCode := runGoTool(folder, runCoverageArgs)
	if exitCode == 2 { // build failure
		result.BuildFail = out
		return result
	}
	result.TestFail = exitCode != 0
	result.Status = getTestEvents(out)
	if exitCode != 0 { // skip coverage
		return result
	}
	out, _ = runGoTool(folder, getCoverageArgs)
	result.Coverage = getCoverage(out)
	return result
}

func hasAnyPrefix(input string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(input, prefix) {
			return true
		}
	}
	return false
}

func runGoTool(folder string, args []string) ([]byte, int) {
	var exitCode int
	cmd := exec.Command("go", args...)
	cmd.SetDir(folder)
	out, exitCode := cmd.SimpleOutput()
	return out, exitCode
}

func getTestEvents(output []byte) []TestStatus {
	results := []testEvent{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		results = append(results, *parseTestEventLine(scanner.Bytes()))
	}
	return groupTestEvents(results)
}

func groupTestEvents(events []testEvent) []TestStatus {
	type packageTest struct {
		Package string
		Test    string
	}

	orderedTests := []packageTest{}
	grouped := map[packageTest][]testEvent{}
	for _, event := range events {
		pt := packageTest{event.Package, event.Test}
		if _, ok := grouped[pt]; !ok {
			orderedTests = append(orderedTests, pt)
		}
		grouped[pt] = append(grouped[pt], event)
	}

	orderedAndGrouped := []TestStatus{}
	for _, test := range orderedTests {
		orderedAndGrouped = append(orderedAndGrouped, *getGroupedTestEvent(grouped[test]))
	}
	return orderedAndGrouped
}

func getGroupedTestEvent(events []testEvent) *TestStatus {
	var pkg, test, testResult, output string
	var elapsed float64
	var buf strings.Builder
	var start time.Time
	for i, event := range events {
		pkg = event.Package
		test = event.Test
		output = strings.TrimSpace(event.Output)
		if i == 0 {
			start = event.Time
		}
		if i == len(events)-1 { // get elapsed time. Overwritten if elapsed is actually provided
			elapsed = event.Time.Sub(start).Seconds()
		}
		if event.Action == "run" {
			continue
		}
		if event.Action == "output" && hasAnyPrefix(output, []string{"===", "---", "PASS", "ok  \t", "FAIL", "SKIP"}) {
			continue
		}
		if event.Action == "pass" || event.Action == "skip" || event.Action == "fail" {
			testResult = event.Action
			elapsed = event.Elapsed
		}
		if event.Action == "output" {
			buf.WriteString(output)
			buf.WriteString("\n")
		}
	}
	return &TestStatus{Elapsed: elapsed, TestResult: testResult, Package: pkg, Test: test, Output: strings.TrimSpace(buf.String())}
}

func parseTestEventLine(line []byte) *testEvent {
	event := &testEvent{}
	if err := json.Unmarshal(line, &event); err != nil {
		lineStr := strings.TrimSpace(string(line))
		if lineStr != "" {
			return &testEvent{Output: string(line)}
		}
	}
	return event
}

func getCoverage(output []byte) []FunctionCoverage {
	results := []FunctionCoverage{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		filename, funcName, funcPercent := parseCoverageLine(line)
		results = append(results, FunctionCoverage{
			Filename:        filename,
			Function:        funcName,
			CoveragePercent: funcPercent,
		})
	}
	return results
}

func parseCoverageLine(line string) (string, string, float32) {
	lastColonIndex := strings.LastIndex(line, ":")
	if lastColonIndex == -1 {
		return line, "", 0
	}
	filename := getFilename(line)
	funcName, funcPercent := parseNameAndPercent(strings.TrimSpace(line[lastColonIndex+1:]))
	return filename, funcName, funcPercent
}

func getFilename(line string) string {
	return filepath.Base(line[:strings.Index(line, ":")])
}

func parseNameAndPercent(nameAndPercent string) (string, float32) {
	spaceIndex := strings.Index(nameAndPercent, "\t")
	if spaceIndex == -1 {
		return nameAndPercent, 0
	}
	funcName := nameAndPercent[:spaceIndex]
	percent := parsePercent(strings.TrimSpace(nameAndPercent[spaceIndex+1:]))
	return funcName, percent
}

func parsePercent(percent string) float32 {
	if len(percent) == 0 {
		return 0
	}
	if percent[len(percent)-1] == '%' {
		percent = percent[:len(percent)-1]
	}
	v, _ := strconv.ParseFloat(percent, 32)
	return float32(v)
}
