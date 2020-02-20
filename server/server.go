package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	toml "github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	dbm "github.com/tendermint/tm-db"

	"github.com/coinexchain/trade-server/core"
	"github.com/coinexchain/trade-server/rocksdb"
)

const (
	ReadTimeout  = 10
	WriteTimeout = 10
	WaitTimeout  = 10

	DexTopic = "coinex-dex"
	DbName   = "dex-trade"
)

type TradeServer struct {
	httpSvr  *http.Server
	hub      *core.Hub
	consumer Consumer
}

func NewServer(svrConfig *toml.Tree) *TradeServer {
	var (
		db        dbm.DB
		err       error
		hub       *core.Hub
		httpSvr   *http.Server
		consumer  Consumer
		wsManager = core.NewWebSocketManager()
	)
	if db, err = initDB(svrConfig); err != nil {
		log.WithError(err).Fatal("init db failed")
		return nil
	}
	if hub, err = initHub(svrConfig, db, wsManager); err != nil {
		log.WithError(err).Error("init hub failed")
		return nil
	}
	if httpSvr, err = initWebService(svrConfig, hub, wsManager); err != nil {
		log.WithError(err).Error("init web service failed")
		return nil
	}
	if consumer = NewConsumer(svrConfig, hub); consumer == nil {
		return nil
	}
	return &TradeServer{
		httpSvr:  httpSvr,
		consumer: consumer,
		hub:      hub,
	}
}

func initDB(svrConfig *toml.Tree) (dbm.DB, error) {
	var (
		db  dbm.DB
		err error
	)
	dataDir := svrConfig.GetDefault("data-dir", "data").(string)
	useRocksDB := svrConfig.GetDefault("use-rocksdb", false).(bool)
	if useRocksDB {
		db, err = rocksdb.NewRocksDB(DbName, dataDir)
	} else {
		db, err = newLevelDB(DbName, dataDir)
	}
	return db, err
}

func initHub(svrConfig *toml.Tree, db dbm.DB, wsManager *core.WebsocketManager) (*core.Hub, error) {
	interval := svrConfig.GetDefault("interval", int64(60)).(int64)
	keepRecent := svrConfig.GetDefault("keepRecent", int64(-1)).(int64)
	monitorInterval := svrConfig.GetDefault("monitorinterval", int64(0)).(int64)
	hub := core.NewHub(db, wsManager, interval, monitorInterval, keepRecent)
	if err := restoreHub(hub); err != nil {
		return nil, err
	}
	return hub, nil
}

func initWebService(svrConfig *toml.Tree, hub *core.Hub, wsManager *core.WebsocketManager) (*http.Server, error) {
	if err := checkHttpsOption(svrConfig); err != nil {
		log.WithError(err).Error("check https required cert file failed")
	}
	httpSvr, err := initHttpService(svrConfig, hub, wsManager)
	return httpSvr, err
}

func checkHttpsOption(svrConfig *toml.Tree) error {
	if httpsToggle, ok := svrConfig.GetDefault("https-toggle",
		false).(bool); ok && httpsToggle {
		certDir := svrConfig.GetDefault("cert-dir", "cert").(string)
		if _, err := os.Stat(certDir); err != nil {
			return fmt.Errorf("certificate path[%s] error", certDir)
		}
	}
	return nil
}

func initHttpService(svrConfig *toml.Tree, hub *core.Hub, wsManager *core.WebsocketManager) (*http.Server, error) {
	proxy := svrConfig.GetDefault("proxy", false).(bool)
	lcd := svrConfig.GetDefault("lcd", "").(string)
	router, err := registerHandler(hub, wsManager, proxy, lcd)
	if err != nil {
		return nil, err
	}
	httpSvr := &http.Server{
		Addr:         fmt.Sprintf(":%d", svrConfig.GetDefault("port", 8000).(int64)),
		Handler:      router,
		ReadTimeout:  ReadTimeout * time.Second,
		WriteTimeout: WriteTimeout * time.Second,
	}
	return httpSvr, nil
}

func (ts *TradeServer) Start(svrConfig *toml.Tree) {
	log.WithField("addr", ts.httpSvr.Addr).Info("Trade-Server start...")
	go ts.startHttpServer(svrConfig)
	go ts.consume()
}

func (ts *TradeServer) startHttpServer(svrConfig *toml.Tree) {
	var (
		certDir     = svrConfig.GetDefault("cert-dir", "cert").(string)
		httpsToggle = svrConfig.GetDefault("https-toggle", false).(bool)
	)
	if httpsToggle {
		certPath := certDir + "/" + serverCertName
		keyPath := certDir + "/" + serverKeyName
		if err := ts.httpSvr.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("https server listen and serve error")
			os.Exit(-1)
		}
		return
	}
	if err := ts.httpSvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.WithError(err).Error("http server listen and serve error")
		os.Exit(-1)
	}
}

func (ts *TradeServer) consume() {
	ts.consumer.Consume()
}

func (ts *TradeServer) Stop() {
	// stop http server
	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout*time.Second)
	defer cancel()
	if err := ts.httpSvr.Shutdown(ctx); err != nil {
		log.WithError(err).Error("http server shutdown failed")
	}

	// stop hub (before closing consumer)
	ts.hub.Close()

	// stop consumer
	ts.consumer.Close()
	log.Info("Server stop...")
}

func newLevelDB(name string, dir string) (db dbm.DB, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("couldn't create db: %v", r)
		}
	}()
	return dbm.NewDB(name, dbm.GoLevelDBBackend, dir), err
}

func restoreHub(hub *core.Hub) error {
	data := hub.LoadDumpData()
	if data == nil {
		log.Info("dump data does not exist, init new hub")
		return nil
	}
	hub4jo := &core.HubForJSON{}
	if err := json.Unmarshal(data, hub4jo); err != nil {
		return err
	}
	hub.Load(hub4jo)
	log.Info("restore hub finish")
	return nil
}
