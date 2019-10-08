package autotest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EndFirstCorp/execfactory"
)

var runCoverageArgs = []string{"test", "-json", "-short", "-coverprofile", "cover.out", "-timeout", "5s"}
var getCoverageArgs = []string{"tool", "cover", "-func=cover.out"}

type groupedTestEvent struct {
	Elapsed    float64
	Package    string
	Test       string
	TestResult string
	Output     string
}

type testEvent struct {
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type coverage struct {
	Filename        string
	Function        string
	CoveragePercent float32
}

var exec = execfactory.NewOSCreator()

// RunTests will run a new set of tests whenever a file changes
func RunTests(folder string) error {
	fmt.Println("Changed folder:", folder)
	out, err, exitCode := runGoTool(folder, runCoverageArgs)
	if exitCode == 2 { // build failure
		return fmt.Errorf("Error running build\n%v", string(err))
	}
	printTestEvents(getTestEvents(out))

	out, _, _ = runGoTool(folder, getCoverageArgs)
	printCoverage(getCoverage(out))

	return nil
}

func printTestEvents(groupedEvents []groupedTestEvent) {
	for _, event := range groupedEvents {
		fmt.Println(event.Elapsed, event.Package, event.Test, event.TestResult, strings.TrimSpace(event.Output))
	}
}

func hasAnyPrefix(input string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(input, prefix) {
			return true
		}
	}
	return false
}

func printCoverage(coverageItems []coverage) {
	for _, coverage := range coverageItems {
		fmt.Println(coverage)
	}
}

func runGoTool(folder string, args []string) ([]byte, []byte, int) {
	var exitCode int
	cmd := exec.Command("go", args...)
	cmd.SetDir(folder)
	out, err, exitCode := cmd.SeparateOutput()
	return out, err, exitCode
}

func getTestEvents(output []byte) []groupedTestEvent {
	results := []testEvent{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		results = append(results, *parseTestEventLine(scanner.Bytes()))
	}
	return groupTestEvents(results)
}

func groupTestEvents(events []testEvent) []groupedTestEvent {
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

	orderedAndGrouped := []groupedTestEvent{}
	for _, test := range orderedTests {
		orderedAndGrouped = append(orderedAndGrouped, *getGroupedTestEvent(grouped[test]))
	}
	return orderedAndGrouped
}

func getGroupedTestEvent(events []testEvent) *groupedTestEvent {
	var pkg, test, testResult string
	var elapsed float64
	var buf strings.Builder
	for _, event := range events {
		pkg = event.Package
		test = event.Test
		if event.Action == "run" {
			continue
		}
		if event.Action == "output" && hasAnyPrefix(event.Output, []string{"===", "---", "PASS", "ok  \t", "FAIL", "SKIPPED"}) {
			continue
		}
		if event.Action == "pass" || event.Action == "skip" || event.Action == "fail" {
			testResult = event.Action
			elapsed = event.Elapsed
		}
		if event.Action == "output" {
			buf.WriteString(event.Output)
		}
	}
	return &groupedTestEvent{Elapsed: elapsed, TestResult: testResult, Package: pkg, Test: test, Output: buf.String()}
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

func getCoverage(output []byte) []coverage {
	results := []coverage{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		filename, funcName, funcPercent := parseCoverageLine(line)
		results = append(results, coverage{
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
