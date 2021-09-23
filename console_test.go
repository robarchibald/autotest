package autotest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/assert"
)

var buildFailure = `# cover github.com/robarchibald/autotest
2019/10/08 17:51:44 cover: autotest/console.go: autotest/console.go:66:2: expected ';', found x (and 2 more errors)
FAIL    github.com/robarchibald/autotest [build failed]`

func TestPrintTest(t *testing.T) {
	header := "\n----------------------------------- folderName -----------------------------------\n"
	tests := []struct {
		name     string
		result   *TestResult
		wantText string
	}{
		// header should be 80 characters wide plus 2 spaces
		{"just header",
			&TestResult{Folder: "folderName"},
			header},
		{"error",
			&TestResult{Folder: "folderName", Error: fmt.Errorf("file:3:4:fail")},
			header + fmt.Sprintf("Error in file at line %s, column %s\n%s\n", aurora.Blue("3"), aurora.Blue("4"), aurora.Red("fail"))},
		{"status",
			&TestResult{Folder: "folderName", Status: []TestStatus{{Elapsed: 1.23, Package: "pkg", Test: "TestMe", TestResult: "fail", Output: ""}}},
			header + "       " + aurora.Blue("--- Test Results ---").String() + "\n" +
				getColumns([]string{"Time  ", "Package", "Test     ", "Status"}) + "\n" +
				aurora.Red(formatFloat(1.23, 2)+"s ").String() + " pkg  " +
				aurora.BrightWhite("TestMe   ").String() + " " + aurora.Red("FAIL").String() + " \n",
		},
		{"coverage",
			&TestResult{Folder: "folderName", Coverage: []FunctionCoverage{{Filename: "file", Function: "func", LineNumber: 1, CoveragePercent: float32(1.12)}}},
			header + "    " + aurora.Blue("--- Code Coverage ---").String() + "\n" +
				getColumns([]string{"Filename", "Function", "Coverage"}) + "\n" +
				"file func " + aurora.BrightRed("1.1%").String() + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &fakePrinter{}
			Println = p.Println
			Printf = p.Printf
			Print = p.Print

			PrintTest(tt.result)
			assert.Equal(t, tt.wantText, p.printed.String())
		})
	}
}

func getColumns(columns []string) string {
	var buf strings.Builder
	for _, column := range columns {
		buf.WriteString(fmt.Sprint(aurora.Gray(15, column+" ")))
	}
	return buf.String()
}

func TestPrintBuildFailure(t *testing.T) {
	p := &fakePrinter{}
	Println = p.Println
	Printf = p.Printf
	Print = p.Print
	printBuildFailure(fmt.Errorf(buildFailure))
	assert.Equal(t,
		fmt.Sprintf("Error in autotest/console.go at line %s, column %s\n%s\n", aurora.Blue("66"), aurora.Blue("2"), aurora.Red("expected ';', found x (and 2 more errors)")),
		p.printed.String())
}

type fakePrinter struct {
	printed strings.Builder
}

func (p *fakePrinter) Println(a ...interface{}) (n int, err error) {
	return p.printed.WriteString(fmt.Sprintln(a...))
}
func (p *fakePrinter) Printf(format string, a ...interface{}) (n int, err error) {
	return p.printed.WriteString(fmt.Sprintf(format, a...))
}
func (p *fakePrinter) Print(a ...interface{}) (n int, err error) {
	return p.printed.WriteString(fmt.Sprint(a...))
}
