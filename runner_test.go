package autotest

import (
	"testing"
)

func TestGetTestEvents(t *testing.T) {
	output := `{"Time":"2019-09-25T18:24:29.864601-07:00","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}
	{"Time":"2019-09-25T18:24:29.864909-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}
	{"Time":"2019-09-25T18:24:29.864953-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"stuff\n"}
	{"Time":"2019-09-25T18:24:29.864977-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"--- PASS: TestHi (0.00s)\n"}
	{"Time":"2019-09-25T18:24:29.864987-07:00","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Elapsed":0}
	{"Time":"2019-09-25T18:24:29.865004-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"PASS\n"}
	{"Time":"2019-09-25T18:24:29.865088-07:00","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"ok  \tgithub.com/6degreeshealth/autotest/cmd\t0.006s\n"}
	{"Time":"2019-09-25T18:24:29.865105-07:00","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Elapsed":0.006}`
	if events := getTestEvents([]byte(output)); len(events) != 8 {
		t.Error("expected to have parsed 8 lines", len(events))
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

	if name, percent := parseNameAndPercent("invalid"); name != "invalid" || percent != float32(0) {
		t.Error("Expected correct values", name, percent)
	}
}
