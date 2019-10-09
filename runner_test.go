package autotest

import (
	"reflect"
	"testing"
	"time"

	"github.com/EndFirstCorp/execfactory"
)

func TestRunTests(t *testing.T) {
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{})
	RunTests("folder")
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{SimpleOutputOut: []byte("test"), SimpleOutputExitCode: 2},
	})
	RunTests("folder")
}

func TestRunGoTool(t *testing.T) {
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{SimpleOutputOut: []byte("test"), SimpleOutputExitCode: 1},
	})
	if out, code := runGoTool("folder", nil); code != 1 || string(out) != "test" {
		t.Error("Expected correct error", code)
	}
}

func TestGetTestEvents(t *testing.T) {
	if events, _ := getTestEvents([]byte(testOutput)); len(events) != 2 || events[0].Package != "github.com/6degreeshealth/autotest/cmd" || events[0].Test != "TestHi" || events[1].Package != "github.com/6degreeshealth/autotest/cmd" || events[1].Test != "" {
		t.Error("expected to have parsed 2 lines", events)
	}
}

func TestParseTestEventLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantTest *testEvent
		wantOk   bool
	}{
		{
			"Run action",
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.UTC), Action: "run", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi"}, true,
		},
		{
			"Output action",
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.UTC), Action: "output", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi", Output: "=== RUN   TestHi\n"}, true,
		},
		{
			"Test pass action",
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":10}`,
			&testEvent{Time: time.Date(2019, 9, 25, 18, 24, 0, 0, time.UTC), Elapsed: 10, Action: "pass", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi"}, true,
		},
		{"Bare line", "bare line", &testEvent{}, false},
		{"Bogus JSON", `{{}`, &testEvent{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTest, gotOk := parseTestEventLine([]byte(tt.line))
			if !reflect.DeepEqual(gotTest, tt.wantTest) {
				t.Errorf("parseTestEventLine() = \n%v, wantTest \n%v", gotTest, tt.wantTest)
			}
			if gotOk != tt.wantOk {
				t.Errorf("parseTestEventLine() = \n%v, wantOk \n%v", gotOk, tt.wantOk)
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
