package application

import (
	"context"
	"fmt"
	"ox/pkg/olog"
	"ox/pkg/util/ocolor"
	"ox/pkg/util/ocycle"
	"ox/pkg/util/odebug"
	"ox/pkg/util/odefer"
	"ox/pkg/util/ogo"
	"sync"
	"time"

	"ox/pkg/component"
	job "ox/pkg/worker/ojob"

	"github.com/BurntSushi/toml"
	"ox/pkg/conf"

	//go-lint
	_ "ox/pkg/conf/datasource/file"
	_ "ox/pkg/conf/datasource/http"
	_ "ox/pkg/registry/etcdv3"

	"golang.org/x/sync/errgroup"
	"ox/pkg/ecode"
	"ox/pkg/flag"
	"ox/pkg/registry"
	"ox/pkg/server"
	"ox/pkg/signals"
	"ox/pkg/worker"
)

const (
	//StageAfterStop after app stop
	StageAfterStop uint32 = iota + 1
	//StageBeforeStop before app stop
	StageBeforeStop
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	cycle        *ocycle.Cycle
	smu          *sync.RWMutex
	initOnce     sync.Once
	startupOnce  sync.Once
	stopOnce     sync.Once
	servers      []server.Server
	workers      []worker.Worker
	jobs         map[string]job.Runner
	logger       *olog.Logger
	hooks        map[uint32]*odefer.DeferStack
	configParser conf.Unmarshaller
	disableMap   map[Disable]bool
	HideBanner   bool
	stopped      chan struct{}
	components   []component.Component
}

// New create a new Application instance
func New(fns ...func() error) (*Application, error) {
	app := &Application{}
	if err := app.Startup(fns...); err != nil {
		return nil, err
	}
	return app, nil
}

func DefaultApp() *Application {
	app := &Application{}
	app.initialize()
	return app
}

//init hooks
func (app *Application) initHooks(hookKeys ...uint32) {
	app.hooks = make(map[uint32]*odefer.DeferStack, len(hookKeys))
	for _, k := range hookKeys {
		app.hooks[k] = odefer.NewStack()
	}
}

//run hooks
func (app *Application) runHooks(k uint32) {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Clean()
	}
}

//RegisterHooks register a stage Hook
func (app *Application) RegisterHooks(k uint32, fns ...func() error) error {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Push(fns...)
		return nil
	}
	return fmt.Errorf("hook stage not found")
}

// initialize application
func (app *Application) initialize() {
	app.initOnce.Do(func() {
		//assign
		app.cycle = ocycle.NewCycle()
		app.smu = &sync.RWMutex{}
		app.servers = make([]server.Server, 0)
		app.workers = make([]worker.Worker, 0)
		app.jobs = make(map[string]job.Runner)
		app.logger = olog.OxLogger
		app.configParser = toml.Unmarshal
		app.disableMap = make(map[Disable]bool)
		app.stopped = make(chan struct{})
		app.components = make([]component.Component, 0)
		//private method
		app.initHooks(StageBeforeStop, StageAfterStop)

		app.parseFlags()
		app.printBanner()
	})
}

// // start up application
// // By default the startup composition is:
// // - parse config, watch, version flags
// // - load config
// // - init default biz logger, ox frame logger
// // - init procs
// func (app *Application) startup() (err error) {
// 	app.startupOnce.Do(func() {
// 		err = ogo.SerialUntilError(
// 			app.parseFlags,
// 			// app.printBanner,
// 			// app.loadConfig,
// 			// app.initLogger,
// 			// app.initMaxProcs,
// 			// app.initTracer,
// 			// app.initSentinel,
// 			// app.initGovernor,
// 		)()
// 	})
// 	return
// }

//Startup ..
func (app *Application) Startup(fns ...func() error) error {
	app.initialize()
	// if err := app.startup(); err != nil {
	// 	return err
	// }
	return ogo.SerialUntilError(fns...)()
}

// Defer ..
// Deprecated: use AfterStop instead
// func (app *Application) Defer(fns ...func() error) {
// 	app.AfterStop(fns...)
// }

// BeforeStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) BeforeStop(fns ...func() error) {
// 	app.RegisterHooks(StageBeforeStop, fns...)
// }

// AfterStop hook
// Deprecated: use RegisterHooks instead
// func (app *Application) AfterStop(fns ...func() error) {
// 	app.RegisterHooks(StageAfterStop, fns...)
// }

// Serve start server
func (app *Application) Serve(s ...server.Server) error {
	app.smu.Lock()
	defer app.smu.Unlock()
	app.servers = append(app.servers, s...)
	return nil
}

// Schedule ..
func (app *Application) Schedule(w worker.Worker) error {
	app.workers = append(app.workers, w)
	return nil
}

// Job ..
func (app *Application) Job(runner job.Runner) error {
	namedJob, ok := runner.(interface{ GetJobName() string })
	// job runner must implement GetJobName
	if !ok {
		return nil
	}
	jobName := namedJob.GetJobName()
	if flag.Bool("disable-job") {
		app.logger.Info("ox disable job", olog.FieldName(jobName))
		return nil
	}

	// start job by name
	jobFlag := flag.String("job")
	if jobFlag == "" {
		app.logger.Error("ox jobs flag name empty", olog.FieldName(jobName))
		return nil
	}

	if jobName != jobFlag {
		app.logger.Info("ox disable jobs", olog.FieldName(jobName))
		return nil
	}
	app.logger.Info("ox register job", olog.FieldName(jobName))
	app.jobs[jobName] = runner
	return nil
}

// SetRegistry set customize registry
// Deprecated, please use registry.DefaultRegisterer instead.
func (app *Application) SetRegistry(reg registry.Registry) {
	registry.DefaultRegisterer = reg
}

// SetGovernor set governor addr (default 127.0.0.1:0)
// Deprecated
//func (app *Application) SetGovernor(addr string) {
//	app.governorAddr = addr
//}

// Run run application
func (app *Application) Run(servers ...server.Server) error {
	app.smu.Lock()
	app.servers = append(app.servers, servers...)
	app.smu.Unlock()

	app.waitSignals() //start signal listen task in goroutine
	defer app.clean()

	// todo jobs not graceful
	app.startJobs()

	// start servers and govern server
	app.cycle.Run(app.startServers)
	// start workers
	app.cycle.Run(app.startWorkers)

	//blocking and wait quit
	if err := <-app.cycle.Wait(); err != nil {
		app.logger.Error("ox shutdown with error", olog.FieldMod(ecode.ModApp), olog.FieldErr(err))
		return err
	}
	app.logger.Info("shutdown ox, bye!", olog.FieldMod(ecode.ModApp))
	return nil
}

//clean after app quit
func (app *Application) clean() {
	_ = olog.DefaultLogger.Flush()
	_ = olog.OxLogger.Flush()
}

// Stop application immediately after necessary cleanup
func (app *Application) Stop() (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(StageBeforeStop)

		//stop servers
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(s.Stop)
			}(s)
		}
		app.smu.RUnlock()

		//stop workers
		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(StageAfterStop)
		app.cycle.Close()
	})
	return
}

// GracefulStop application after necessary cleanup
func (app *Application) GracefulStop(ctx context.Context) (err error) {
	app.stopOnce.Do(func() {
		app.stopped <- struct{}{}
		app.runHooks(StageBeforeStop)

		//stop servers
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(func() error {
					return s.GracefulStop(ctx)
				})
			}(s)
		}
		app.smu.RUnlock()

		//stop workers
		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(StageAfterStop)
		app.cycle.Close()
	})
	return err
}

// waitSignals wait signal
func (app *Application) waitSignals() {
	app.logger.Info("init listen signal", olog.FieldMod(ecode.ModApp), olog.FieldEvent("init"))
	signals.Shutdown(func(grace bool) { //when get shutdown signal
		//todo: support timeout
		if grace {
			app.GracefulStop(context.TODO())
		} else {
			app.Stop()
		}
	})
}

// func (app *Application) initGovernor() error {
// 	if app.isDisable(DisableDefaultGovernor) {
// 		app.logger.Info("defualt governor disable", olog.FieldMod(ecode.ModApp))
// 		return nil
// 	}

// 	config := governor.StdConfig("governor")
// 	if !config.Enable {
// 		return nil
// 	}
// 	return app.Serve(config.Build())
// }

func (app *Application) startServers() error {
	var eg errgroup.Group
	var ctx, cancel = context.WithTimeout(context.Background(), time.Second*10)
	go func() {
		<-app.stopped
		cancel()
	}()
	// start multi servers
	for _, s := range app.servers {
		s := s
		eg.Go(func() (err error) {
			registry.DefaultRegisterer.RegisterService(ctx, s.Info())
			defer registry.DefaultRegisterer.UnregisterService(ctx, s.Info())
			app.logger.Info("start server", olog.FieldMod(ecode.ModApp), olog.FieldEvent("init"), olog.FieldName(s.Info().Name), olog.FieldAddr(s.Info().Label()), olog.Any("scheme", s.Info().Scheme))
			defer app.logger.Info("exit server", olog.FieldMod(ecode.ModApp), olog.FieldEvent("exit"), olog.FieldName(s.Info().Name), olog.FieldErr(err), olog.FieldAddr(s.Info().Label()))
			err = s.Serve()
			return
		})
	}
	return eg.Wait()
}

func (app *Application) startWorkers() error {
	var eg errgroup.Group
	// start multi workers
	for _, w := range app.workers {
		w := w
		eg.Go(func() error {
			return w.Run()
		})
	}
	return eg.Wait()
}

// todo handle error
func (app *Application) startJobs() error {
	if len(app.jobs) == 0 {
		return nil
	}
	var jobs = make([]func(), 0)
	//warp jobs
	for name, runner := range app.jobs {
		jobs = append(jobs, func() {
			app.logger.Info("job run begin", olog.FieldName(name))
			defer app.logger.Info("job run end", olog.FieldName(name))
			// runner.Run panic 错误在更上层抛出
			runner.Run()
		})
	}
	ogo.Parallel(jobs...)()
	return nil
}

//parseFlags init
func (app *Application) parseFlags() error {
	if app.isDisable(DisableParserFlag) {
		app.logger.Info("parseFlags disable", olog.FieldMod(ecode.ModApp))
		return nil
	}
	// flag.Register(&flag.StringFlag{
	// 	Name:    "config",
	// 	Usage:   "--config",
	// 	EnvVar:  "OX_CONFIG",
	// 	Default: "",
	// 	Action:  func(name string, fs *flag.FlagSet) {},
	// })

	// flag.Register(&flag.BoolFlag{
	// 	Name:    "version",
	// 	Usage:   "--version, print version",
	// 	Default: false,
	// 	Action: func(string, *flag.FlagSet) {
	// 		pkg.PrintVersion()
	// 		os.Exit(0)
	// 	},
	// })

	// flag.Register(&flag.StringFlag{
	// 	Name:    "host",
	// 	Usage:   "--host, print host",
	// 	Default: "127.0.0.1",
	// 	Action:  func(string, *flag.FlagSet) {},
	// })
	return flag.Parse()
}

//loadConfig init
// func (app *Application) loadConfig() error {
// 	if app.isDisable(DisableLoadConfig) {
// 		app.logger.Info("load config disable", olog.FieldMod(ecode.ModConfig))
// 		return nil
// 	}

// 	var configAddr = flag.String("config")
// 	provider, err := manager.NewDataSource(configAddr)
// 	if err != manager.ErrConfigAddr {
// 		if err != nil {
// 			app.logger.Panic("data source: provider error", olog.FieldMod(ecode.ModConfig), olog.FieldErr(err))
// 		}

// 		if err := conf.LoadFromDataSource(provider, app.configParser); err != nil {
// 			app.logger.Panic("data source: load config", olog.FieldMod(ecode.ModConfig), olog.FieldErrKind(ecode.ErrKindUnmarshalConfigErr), olog.FieldErr(err))
// 		}
// 	} else {
// 		app.logger.Info("no config... ", olog.FieldMod(ecode.ModConfig))
// 	}
// 	return nil
// }

//initLogger init
// func (app *Application) initLogger() error {
// 	if conf.Get(olog.ConfigEntry("default")) != nil {
// 		olog.DefaultLogger = olog.RawConfig(constant.ConfigPrefix + ".logger.default").Build()
// 	}
// 	olog.DefaultLogger.AutoLevel(constant.ConfigPrefix + ".logger.default")

// 	if conf.Get(constant.ConfigPrefix+".logger.ox") != nil {
// 		olog.OxLogger = olog.RawConfig(constant.ConfigPrefix + ".logger.ox").Build()
// 	}
// 	olog.OxLogger.AutoLevel(constant.ConfigPrefix + ".logger.ox")

// 	return nil
// }

//initTracer init
// func (app *Application) initTracer() error {
// 	// init tracing component jaeger
// 	if conf.Get("ox.trace.jaeger") != nil {
// 		var config = jaeger.RawConfig("ox.trace.jaeger")
// 		trace.SetGlobalTracer(config.Build())
// 	}
// 	return nil
// }

//initSentinel init
// func (app *Application) initSentinel() error {
// 	// init reliability component sentinel
// 	if conf.Get("ox.reliability.sentinel") != nil {
// 		app.logger.Info("init sentinel")
// 		return sentinel.RawConfig("ox.reliability.sentinel").Build()
// 	}
// 	return nil
// }

//initMaxProcs init
// func (app *Application) initMaxProcs() error {
// 	if maxProcs := conf.GetInt("maxProc"); maxProcs != 0 {
// 		runtime.GOMAXPROCS(maxProcs)
// 	} else {
// 		if _, err := maxprocs.Set(); err != nil {
// 			app.logger.Panic("auto max procs", olog.FieldMod(ecode.ModProc), olog.FieldErrKind(ecode.ErrKindAny), olog.FieldErr(err))
// 		}
// 	}
// 	app.logger.Info("auto max procs", olog.FieldMod(ecode.ModProc), olog.Int64("procs", int64(runtime.GOMAXPROCS(-1))))
// 	return nil
// }

func (app *Application) isDisable(d Disable) bool {
	b, ok := app.disableMap[d]
	if !ok {
		return false
	}
	return b
}

//printBanner init
func (app *Application) printBanner() error {
	if app.HideBanner {
		return nil
	}

	if odebug.IsTestingMode() {
		return nil
	}

	const banner = `
  ___   __  __
 / _ \  \ \/ /
| (_) |  >  < 
 \___/  /_/\_\

 Welcome to ox, starting application ...
`
	fmt.Println(ocolor.Green(banner))
	return nil
}
