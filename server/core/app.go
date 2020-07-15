package core

import (
	"github.com/bleenco/abstruse/pkg/etcd/embed"
	"github.com/bleenco/abstruse/server/config"
	"github.com/bleenco/abstruse/server/db/repository"
	"github.com/bleenco/abstruse/server/ws"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

// App represents main server application which handles all the services.
type App struct {
	cfg       *config.Config
	logger    *zap.SugaredLogger
	etcd      *embed.Server
	client    *clientv3.Client
	ws        *ws.Server
	scheduler *Scheduler
	repo      repository.Repo
	stopch    chan struct{}
}

// NewApp returns new instance of main application.
func NewApp() *App {
	logger := Log.With(zap.String("type", "app")).Sugar()
	etcd := embed.NewServer(Config, Log)
	ws := ws.NewServer(Config, Log)
	app := &App{
		cfg:    Config,
		logger: logger,
		etcd:   etcd,
		ws:     ws,
		repo:   repository.NewRepo(),
		stopch: make(chan struct{}),
	}
	app.scheduler = NewScheduler(Log, app)
	return app
}

// Run starts the main application with all services.
func (a *App) Run() error {
	var err error
	errch := make(chan error)

	go func() {
		if err := a.ws.Run(); err != nil {
			errch <- err
		}
	}()

	a.client, err = a.startEtcd()
	if err != nil {
		errch <- err
	}

	go func() {
		if err := a.scheduler.Run(a.client); err != nil {
			errch <- err
		}
	}()

	go func() {
		if err := a.watchWorkers(); err != nil {
			errch <- err
		}
	}()

	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	if err := a.TriggerBuild(1, 4); err != nil {
	// 		fmt.Println(err)
	// 	}
	// }()

	select {
	case err := <-errch:
		return err
	case <-a.stopch:
		return nil
	}
}

// Stop stops the application.
func (a *App) Stop() {
	a.stopch <- struct{}{}
}

// RestartEtcd restarts etcd server.
func (a *App) RestartEtcd() error {
	var err error
	a.etcd.Stop()
	a.etcd = embed.NewServer(Config, Log)
	a.client, err = a.startEtcd()

	return err
}

// SaveAuthConfig saves new authentication configuration.
func (a *App) SaveAuthConfig(cfg *config.Auth) error {
	return saveAuthConfig(cfg)
}

// SaveDBConfig saves new database configuration.
func (a *App) SaveDBConfig(cfg *config.Db) error {
	return saveDBConfig(cfg)
}

// SaveEtcdConfig saves new etcd configuration.
func (a *App) SaveEtcdConfig(cfg *config.Etcd) error {
	return saveEtcdConfig(cfg)
}

func (a *App) startEtcd() (*clientv3.Client, error) {
	if err := a.etcd.Run(); err != nil {
		return nil, err
	}

	return a.etcd.GetClient()
}
