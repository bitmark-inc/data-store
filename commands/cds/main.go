package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	bitmarksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/data-store/cds"
	"github.com/bitmark-inc/data-store/store"
	"github.com/bitmark-inc/data-store/web"
)

var (
	server *web.Server
)

func initLog() {
	logLevel, err := log.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(logLevel)
	}

	log.SetOutput(os.Stdout)

	log.SetFormatter(&prefixed.TextFormatter{
		ForceFormatting: true,
		FullTimestamp:   true,
	})
}

func loadConfig(file string) {
	// Config from file
	viper.SetConfigType("yaml")
	if file != "" {
		viper.SetConfigFile(file)
	}

	viper.AddConfigPath("/.config/")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("No config file. Read config from env.")
		viper.AllowEmptyEnv(false)
	}

	// Config from env if possible
	viper.AutomaticEnv()
	viper.SetEnvPrefix("ds")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func main() {
	var configFile string

	initialCtx, cancelInitialization := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info("Server is preparing to shutdown")

		if initialCtx != nil && cancelInitialization != nil {
			log.Info("Cancelling initialization")
			cancelInitialization()
			<-initialCtx.Done()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if server != nil {
			log.Info("Shutdown mobile api server")
			if err := server.Shutdown(ctx); err != nil {
				log.Error("Server Shutdown:", err)
			}
		}

		os.Exit(1)
	}()

	flag.StringVar(&configFile, "c", "./config.yaml", "[optional] path of configuration file")
	flag.Parse()

	loadConfig(configFile)

	initLog()

	// Init Bitmark SDK
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	bitmarksdk.Init(&bitmarksdk.Config{
		Network:    bitmarksdk.Network(viper.GetString("bitmarksdk.network")),
		APIToken:   viper.GetString("bitmarksdk.token"),
		HTTPClient: httpClient,
	})
	log.WithField("prefix", "init").Info("Initialized bitmark sdk")

	opts := options.Client().ApplyURI(viper.GetString("mongo.conn"))
	opts.SetMaxPoolSize(viper.GetUint64("mongo.pool"))
	mongoClient, err := mongo.NewClient(opts)
	if nil != err {
		log.Panicf("create mongo client with error: %s", err)
	}

	if err := mongoClient.Connect(context.Background()); nil != err {
		log.Panicf("connect mongo database with error: %s", err)
	}

	acct, err := account.FromSeed(viper.GetString("server.bitmark_account_seed"))
	if err != nil {
		log.Panic(err)
	}

	rootKey, err := hex.DecodeString(viper.GetString("server.macaroon_root_key"))
	if err != nil {
		log.Panic(err)
	}

	cds := cds.New(store.NewMongodbDataPool(mongoClient, viper.GetString("server.store_prefix")))

	// Init http server
	server = web.NewServer(viper.GetBool("server.tracing"), acct.(*account.AccountV2), viper.GetString("server.endpoint"), rootKey)
	server.Middleware(server.DumpRequest)
	server.Route("PUT", "/poi_rating/:poi_id", server.CheckMacaroon(), cds.SetPOIRating())
	server.Route("GET", "/poi_rating/:poi_id", server.CheckMacaroon(), cds.GetPOISummarizedRatings)
	server.Route("GET", "/poi_rating", server.CheckMacaroon(), cds.GetPOISummarizedRatings)
	server.Route("POST", "/symptom-daily-reports", server.CheckMacaroon(), cds.AddSymptomDailyReports)
	server.Route("GET", "/report-items", server.CheckMacaroon(), cds.GetSymptomReportItems)
	server.Route("GET", "/data/export", server.CheckMacaroon(), cds.ExportData)
	log.WithField("prefix", "init").Info("Initialized http server")

	// Remove initial context
	initialCtx = nil
	cancelInitialization = nil

	log.Fatal(server.Run(":" + viper.GetString("server.port")))
}
