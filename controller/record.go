package controller

import (
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/at"
	"lazybones-crossing-go/service"
)

// RestController
type recordController struct {
	at.RestController
	recordService service.RecordService
}

func init() {
	app.Register(newRecordController)
}

func newRecordController(recordService service.RecordService) *recordController {
	return &recordController{
		recordService: recordService,
	}
}
