package autotest

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/EndFirstCorp/execfactory"
)

var testOutput = `{"Time":"2019-09-25T18:24:29.864601-07:00","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}
{"Time":"2019-09-25T18:24:29.864909-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}
{"Time":"2019-09-25T18:24:29.864953-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"stuff\n"}
{"Time":"2019-09-25T18:24:29.864977-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"--- PASS: TestHi (0.00s)\n"}
{"Time":"2019-09-25T18:24:29.864987-07:00","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":0}
{"Time":"2019-09-25T18:24:29.865004-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"PASS\n"}
{"Time":"2019-09-25T18:24:29.865088-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"ok  \tgithub.com/6degreeshealth/autotest/cmd\t0.006s\n"}
{"Time":"2019-09-25T18:24:29.865105-07:00","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Elapsed":0.006}`

var coverageOutput = `github.com/6degreeshealth/autotest/runner.go:37:	RunTests		100.0%
github.com/6degreeshealth/autotest/runner.go:58:	runGoTool		100.0%
github.com/6degreeshealth/autotest/runner.go:64:	getTestEvents		100.0%
github.com/6degreeshealth/autotest/runner.go:73:	parseTestEventLine	100.0%
github.com/6degreeshealth/autotest/runner.go:84:	getCoverage		100.0%
github.com/6degreeshealth/autotest/runner.go:99:	parseCoverageLine	100.0%
github.com/6degreeshealth/autotest/runner.go:109:	getFilename		100.0%
github.com/6degreeshealth/autotest/runner.go:113:	parseNameAndPercent	100.0%
github.com/6degreeshealth/autotest/runner.go:123:	parsePercent		83.3%
total:							(statements)		35.5%`

func TestRunTests(t *testing.T) {
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{CombinedOutputVal: []byte(testOutput)},
		{CombinedOutputVal: []byte(coverageOutput)},
	})
	if err := RunTests("folder"); err != nil {
		t.Fatal(err)
	}
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{CombinedOutputErr: errors.New("test")},
	})
	if err := RunTests("folder"); err == nil || err.Error() != "test" {
		t.Error("Expected to error at first runGoTool", err)
	}
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{CombinedOutputVal: []byte{}},
		{CombinedOutputErr: errors.New("test")},
	})
	if err := RunTests("folder"); err == nil || err.Error() != "test" {
		t.Error("Expected to error at second runGoTool", err)
	}
}

func TestRunToTool(t *testing.T) {
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{CombinedOutputErr: errors.New("test")},
		{CombinedOutputVal: []byte("hello")},
		{CombinedOutputErr: &execfactory.ExitError{}},
	})
	if _, err := runGoTool("folder", nil); err == nil || err.Error() != "test" {
		t.Error("Expected correct error", err)
	}
	if out, err := runGoTool("folder", nil); err != nil || string(out) != "hello" {
		t.Error("Expected correct error", err, string(out))
	}
}

func TestGetTestEvents(t *testing.T) {
	if events := getTestEvents([]byte(testOutput)); len(events) != 8 {
		t.Error("expected to have parsed 8 lines", len(events))
	}
}

func TestParseTestEventLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want *testEvent
	}{
		{
			"Run action",
			`{"Time":"2019-09-25T18:24:00.000000-07:00","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.Local), Action: "run", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi"},
		},
		{
			"Output action",
			`{"Time":"2019-09-25T18:24:00.000000-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.Local), Action: "output", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi", Output: "=== RUN   TestHi\n"},
		},
		{
			"Test pass action",
			`{"Time":"2019-09-25T18:24:00.000000-07:00","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":0}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.Local), Action: "pass", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi", Elapsed: 0},
		},
		{
			"Bare line",
			"bare line",
			&testEvent{Output: "bare line"},
		},
		{
			"Bogus JSON",
			`{{}`,
			&testEvent{Output: "{{}"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTestEventLine([]byte(tt.line)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTestEventLine() = \n%v, want \n%v", got, tt.want)
			}
		})
	}
}

func TestGetCoverage(t *testing.T) {
	lines := `line1
	line2
	github.com/6degreeshealth/autotest/cmd/hi.go:3:		me		95.4%`
	if numLines := len(getCoverage([]byte(lines))); numLines != 3 {
		t.Error("expected 3 coverage lines", numLines)
	}
}

func TestParseCoverageLine(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantFilename string
		wantFuncName string
		wantPercent  float32
	}{
		{name: "Valid", line: "github.com/6degreeshealth/autotest/cmd/hi.go:3:		me		95.4%", wantFilename: "hi.go", wantFuncName: "me", wantPercent: float32(95.4)},
		{name: "Total", line: "total:							(statements)	50.0%", wantFilename: "total", wantFuncName: "(statements)", wantPercent: float32(50.0)},
		{name: "Invalid", line: "test", wantFilename: "test", wantFuncName: "", wantPercent: float32(0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilename, gotFuncName, gotPercent := parseCoverageLine(tt.line)
			if gotFilename != tt.wantFilename {
				t.Errorf("parseCoverageLine() got filename = %v, want %v", gotFilename, tt.wantFilename)
			}
			if gotFuncName != tt.wantFuncName {
				t.Errorf("parseCoverageLine() got funcName = %v, want %v", gotFuncName, tt.wantFuncName)
			}
			if gotPercent != tt.wantPercent {
				t.Errorf("parseCoverageLine() got percent = %v, want %v", gotPercent, tt.wantPercent)
			}
		})
	}
}
func TestParseNameAndPercent(t *testing.T) {
	if name, percent := parseNameAndPercent("me		95.4%"); name != "me" || percent != float32(95.4) {
		t.Error("Expected correct values", name, percent)
	}
	if name, percent := parseNameAndPercent("invalidWithoutTab"); name != "invalidWithoutTab" || percent != float32(0) {
		t.Error("Expected correct values", name, percent)
	}
	if name, percent := parseNameAndPercent("invalid	"); name != "invalid" || percent != float32(0) {
		t.Error("Expected correct values", name, percent)
	}
}
