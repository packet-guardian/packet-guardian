package reports

import (
	"errors"
	"net/http"
	"sync"

	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/models/stores"
)

type ReportFunc func(*common.Environment, http.ResponseWriter, *http.Request, stores.StoreCollection) error

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

func RenderReport(name string, w http.ResponseWriter, r *http.Request, stores stores.StoreCollection) error {
	report, ok := reportFuncs[name]
	if !ok {
		return errors.New("Report doesn't exist")
	}

	return report.Call(common.GetEnvironmentFromContext(r), w, r, stores)
}

func GetReports() map[string]*Report {
	return reportFuncs
}

type ReportSorter []*Report

func (r ReportSorter) Len() int           { return len(r) }
func (r ReportSorter) Less(i, j int) bool { return r[i].Fullname < r[j].Fullname }
func (r ReportSorter) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
