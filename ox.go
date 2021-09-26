package ox

import "ox/pkg/application"

var (
	//StageAfterStop after app stop
	StageAfterStop uint32 = application.StageAfterStop
	//StageBeforeStop before app stop
	StageBeforeStop = application.StageBeforeStop
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application = application.Application

var New = application.New
var DefaultApp = application.DefaultApp
