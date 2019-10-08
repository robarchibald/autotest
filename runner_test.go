package autotest

import (
	"reflect"
	"testing"

	"github.com/EndFirstCorp/execfactory"
)

var testOutput = `{"Time":"2019-09-25T18:24:29.864601Z","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}
{"Time":"2019-09-25T18:24:29.864909Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}
{"Time":"2019-09-25T18:24:29.864953Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"stuff\n"}
{"Time":"2019-09-25T18:24:29.864977Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"--- PASS: TestHi (0.00s)\n"}
{"Time":"2019-09-25T18:24:29.864987Z","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":0}
{"Time":"2019-09-25T18:24:29.865004Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"PASS\n"}
{"Time":"2019-09-25T18:24:29.865088Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"ok  \tgithub.com/6degreeshealth/autotest/cmd\t0.006s\n"}
{"Time":"2019-09-25T18:24:29.865105Z","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Elapsed":0.006}`

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
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{})
	if err := RunTests("folder"); err != nil {
		t.Fatal(err)
	}
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{SeparateOutputErr: []byte("test"), SeparateOutputExitCode: 2},
	})
	if err := RunTests("folder"); err == nil || err.Error() != "Error running build\ntest" {
		t.Error("Expected to error at first runGoTool", err)
	}
}

func TestFail(t *testing.T) {
	t.Fatal("error")
}

func TestPrintEvents(t *testing.T) {
	printTestEvents(getTestEvents([]byte(testOutput)))
}

func TestPrintCoverage(t *testing.T) {
	printCoverage(getCoverage([]byte(coverageOutput)))
}

func TestRunGoTool(t *testing.T) {
	exec = execfactory.NewMockCreator([]execfactory.MockInstance{
		{SeparateOutputErr: []byte("test"), SeparateOutputOut: []byte("out"), SeparateOutputExitCode: 1},
	})
	if out, err, code := runGoTool("folder", nil); code != 1 || string(out) != "out" || string(err) != "test" {
		t.Error("Expected correct error", code)
	}
}

func TestGetTestEvents(t *testing.T) {
	if events := getTestEvents([]byte(testOutput)); len(events) != 2 || events[0].Package != "github.com/6degreeshealth/autotest/cmd" || events[0].Test != "TestHi" || events[1].Package != "github.com/6degreeshealth/autotest/cmd" || events[1].Test != "" {
		t.Error("expected to have parsed 2 lines", events)
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
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}`,
			&testEvent{Action: "run", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi"},
		},
		{
			"Output action",
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}`,
			&testEvent{Action: "output", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi", Output: "=== RUN   TestHi\n"},
		},
		{
			"Test pass action",
			`{"Time":"2019-09-25T18:24:00.000000Z","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":10}`,
			&testEvent{Elapsed: 10, Action: "pass", Package: "github.com/6degreeshealth/autotest/cmd", Test: "TestHi"},
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
