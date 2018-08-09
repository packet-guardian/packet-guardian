package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/db"
	"github.com/packet-guardian/packet-guardian/src/models"
)

type Status struct {
	e *common.Environment
}

type StatusResp struct {
	Application *ApplicationStatusResp `json:"application"`
	Database    *DatabaseStatusResp    `json:"database"`
	GoRoutines  *GoRoutineStatusResp   `json:"go_routines"`
	Memory      *MemoryStatusResp      `json:"memory"`
}

type ApplicationStatusResp struct {
	Version string `json:"version"`
}

type DatabaseStatusResp struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
	Type    string `json:"type"`
}

type GoRoutineStatusResp struct {
	RoutineNum int `json:"routine_num"`
}

type MemoryStatusResp struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"total_alloc"`
	Sys          uint64 `json:"sys"`
	Mallocs      uint64 `json:"mallocs"`
	Frees        uint64 `json:"frees"`
	PauseTotalNs uint64 `json:"pause_total_ns"`
	NumGC        uint32 `json:"num_gc"`
	HeapObjects  uint64 `json:"head_objects"`
	LastGC       string `json:"last_gc"`
}

func NewStatusController(e *common.Environment) *Status {
	return &Status{e: e}
}

func (s *Status) GetStatus(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionUser := models.GetUserFromContext(r)

	if !sessionUser.Can(models.ViewDebugInfo) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data := &StatusResp{
		Application: s.applicationStatus(),
		Database:    s.databaseStatus(),
		GoRoutines:  s.goRoutineStatus(),
		Memory:      s.memoryStatus(),
	}

	common.NewAPIResponse("", data).WriteResponse(w, http.StatusOK)
}

func (s *Status) applicationStatus() *ApplicationStatusResp {
	return &ApplicationStatusResp{
		Version: common.SystemVersion,
	}
}

func (s *Status) databaseStatus() *DatabaseStatusResp {
	dbVer := s.e.DB.SchemaVersion()
	dbStatus := "ok"
	if dbVer != db.DBVersion {
		dbStatus = "warning"
	}

	return &DatabaseStatusResp{
		Version: dbVer,
		Status:  dbStatus,
		Type:    s.e.DB.Driver,
	}
}

func (s *Status) goRoutineStatus() *GoRoutineStatusResp {
	return &GoRoutineStatusResp{
		RoutineNum: runtime.NumGoroutine(),
	}
}

func (s *Status) memoryStatus() *MemoryStatusResp {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	return &MemoryStatusResp{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		PauseTotalNs: m.PauseTotalNs,
		NumGC:        m.NumGC,
		HeapObjects:  m.HeapObjects,
		LastGC:       time.Unix(0, int64(m.LastGC)).Format(time.RFC3339),
	}
}
