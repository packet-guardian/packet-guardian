package reports

import (
	"errors"
	"net/http"
	"sync"

	"github.com/usi-lfkeitel/packet-guardian/src/common"
)

type ReportFunc func(*common.Environment, http.ResponseWriter, *http.Request) error

type Report struct {
	Shortname string
	Fullname  string
	Call      ReportFunc
}

var (
	reportFuncs      = make(map[string]*Report)
	reportFuncsMutex sync.Mutex
)

func RegisterReport(shortname, fullname string, report ReportFunc) {
	reportFuncsMutex.Lock()
	reportFuncs[shortname] = &Report{
		Shortname: shortname,
		Fullname:  fullname,
		Call:      report,
	}
	reportFuncsMutex.Unlock()
}

func RenderReport(name string, w http.ResponseWriter, r *http.Request) error {
	report, ok := reportFuncs[name]
	if !ok {
		return errors.New("Report doesn't exist")
	}

	return report.Call(common.GetEnvironmentFromContext(r), w, r)
}

func GetReports() map[string]*Report {
	return reportFuncs
}

type ReportSorter []*Report

func (r ReportSorter) Len() int           { return len(r) }
func (r ReportSorter) Less(i, j int) bool { return r[i].Fullname < r[j].Fullname }
func (r ReportSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
