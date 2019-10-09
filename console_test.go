package autotest

import "testing"

var testOutput = `{"Time":"2019-09-25T18:24:29.864601Z","Action":"run","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi"}
{"Time":"2019-09-25T18:24:29.864909Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"=== RUN   TestHi\n"}
{"Time":"2019-09-25T18:24:29.864953Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"stuff\n"}
{"Time":"2019-09-25T18:24:29.864977Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Test":"TestHi","Output":"--- PASS: TestHi (0.00s)\n"}
{"Time":"2019-09-25T18:24:29.865004Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"PASS\n"}
{"Time":"2019-09-25T18:24:29.865088Z","Action":"output","Package":"github.com/6degreeshealth/autotest/cmd","Output":"ok  \tgithub.com/6degreeshealth/autotest/cmd\t0.006s\n"}
{"Time":"2019-09-25T18:24:29.865105Z","Action":"pass","Package":"github.com/6degreeshealth/autotest/cmd","Elapsed":0.110}`

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

var buildFailure = `# cover github.com/6degreeshealth/autotest
2019/10/08 17:51:44 cover: /Users/rob/Projects/autotest/console.go: /Users/rob/Projects/autotest/console.go:66:2: expected ';', found x (and 2 more errors)
FAIL    github.com/6degreeshealth/autotest [build failed]`

var buildFailure2 = `# github.com/6degreeshealth/autotest
../console.go:66:2: syntax error: unexpected x after top level declaration`

func TestPrintEvents(t *testing.T) {
	printTestEvents(getTestEvents([]byte(testOutput)), true)
}

func TestPrintCoverage(t *testing.T) {
	printCoverage(getCoverage([]byte(coverageOutput)))
}

func TestPrintBuildFailure(t *testing.T) {
	printBuildFailure([]byte(buildFailure))
}
