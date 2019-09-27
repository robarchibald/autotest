package autotest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/EndFirstCorp/execfactory"
)

var runCoverageArgs = []string{"test", "-json", "-short", "-coverprofile", "cover.out", "-timeout", "5s"}
var getCoverageArgs = []string{"tool", "cover", "-func=cover.out"}

type testEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
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

var exec = execfactory.OS

// RunTests will run a new set of tests whenever a file changes
func RunTests(folder string) error {
	fmt.Println("Changed file", folder)
	out, err := runGoTool(folder, runCoverageArgs)
	if err != nil {
		return err
	}
	for _, event := range getTestEvents(out) {
		fmt.Println(event)
	}

	out, err = runGoTool(folder, getCoverageArgs)
	if err != nil {
		return err
	}
	for _, coverage := range getCoverage(out) {
		fmt.Println(coverage)
	}

	return err
}

func runGoTool(folder string, args []string) ([]byte, error) {
	cmd := exec.Command("go", args...)
	cmd.SetDir(folder)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Println("found exit error", ee)
		}
	}
	return out, err
}

func getTestEvents(output []byte) []testEvent {
	results := []testEvent{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		results = append(results, *parseTestEventLine(scanner.Bytes()))
	}
	return results
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
