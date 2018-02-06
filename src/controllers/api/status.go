package api

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/packet-guardian/packet-guardian/src/common"
	"github.com/packet-guardian/packet-guardian/src/db"
	"github.com/packet-guardian/packet-guardian/src/models"
)

type Status struct {
	e *common.Environment
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

	dbVer := s.e.DB.SchemaVersion()
	dbStatus := "ok"
	if dbVer != db.DBVersion {
		dbStatus = "warning"
	}

	data := map[string]interface{}{
		"database_version": dbVer,
		"database_status":  dbStatus,
		"database_type":    s.e.DB.Driver,
	}

	common.NewAPIResponse("", data).WriteResponse(w, http.StatusOK)
}
