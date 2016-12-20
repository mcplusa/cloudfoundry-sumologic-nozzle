package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/caching"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventRouting"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/firehoseclient"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/logging"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/sumoCFFirehose"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug                = true                                                   //debug", "Enable debug mode, print in console
	apiEndpoint          = kingpin.Flag("api-endpoint", "Sumo Endpoint").String() //"https://api.bosh-lite.com"
	sumoEndpoint         = kingpin.Flag("sumo-endpoint", "Sumo Endpoint").String()
	dopplerEndpoint      = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").String()
	subscriptionId       = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	user                 = "firehose_user"     //user created in CF, authorized to connect the firehose
	password             = "firehose_password" // password created along with the firehose_user
	skipSSLValidation    = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	keepAlive, errK      = time.ParseDuration("25s") //default
	wantedEvents         = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).Default("LogMessage").OverrideDefaultFromEnvar("EVENTS").String()
	boltDatabasePath     = "my.db"                   //default
	tickerTime, errT     = time.ParseDuration("60s") //Default
	eventsBatchSize      = kingpin.Flag("log-events-batch-size", "Log Events Batch Size").Default("10").Int()
	sumoPostMinimumDelay = kingpin.Flag("sumo-Post-Minimum-Delay", "Sumo Post Minimum Delay").Default("500ms").Duration()
)

var (
	version = "0.0.0"
)

func main() {
	//logging init
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	kingpin.Version(version)
	kingpin.Parse()

	runtime.GOMAXPROCS(1)

	logging.Info.Println("Configurations set:")
	logging.Info.Println("Api Endpoint: " + *apiEndpoint)
	logging.Info.Println("Sumo Endpoint: " + *sumoEndpoint)
	logging.Info.Println("Cloudfoundry Doppler Endpoint: " + *dopplerEndpoint)
	logging.Info.Println("Cloudfoundry Nozzle Subscription ID: " + *subscriptionId)
	logging.Info.Println("Cloudfoundry User: " + user)

	logging.Info.Printf("Events Batch Size: [%d]\n", eventsBatchSize)

	logging.Info.Println("Starting firehose-to-sumo " + version)

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          user,
		Password:          password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient, _ := cfclient.NewClient(&c)

	if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}

	//Creating Caching
	var cachingClient caching.Caching
	if caching.IsNeeded(*wantedEvents) {
		cachingClient = caching.NewCachingBolt(cfClient, boltDatabasePath)
	} else {
		cachingClient = caching.NewCachingEmpty()
	}

	logging.Info.Println("Creating queue")
	queue := eventQueue.NewQueue(make([]*events.Event, 100))
	loggingClientSumo := sumoCFFirehose.NewSumoLogicAppender(*sumoEndpoint, 1000, &queue, *eventsBatchSize, *sumoPostMinimumDelay)
	go loggingClientSumo.Start() //multi

	logging.Info.Println("Creating Events")
	events := eventRouting.NewEventRouting(cachingClient, *loggingClientSumo, &queue)
	err := events.SetupEventRouting(*wantedEvents)
	if err != nil {
		logging.Error.Fatal("Error setting up event routing: ", err)
		os.Exit(1)

	}

	// Parse extra fields from cmd call
	cachingClient.CreateBucket()
	//Let's Update the database the first time
	logging.Info.Printf("Start filling app/space/org cache.\n")
	apps := cachingClient.GetAllApp()
	logging.Info.Printf("Done filling cache! Found [%d] Apps \n", len(apps))

	//Let's start the goRoutine
	cachingClient.PerformPoollingCaching(tickerTime)

	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   cfClient.Endpoint.DopplerEndpoint,
		InsecureSSLSkipVerify:  *skipSSLValidation,
		IdleTimeoutSeconds:     keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
	}

	logging.Info.Printf("Connecting to Firehose... \n")

	firehoseClient := firehoseclient.NewFirehoseNozzle(cfClient, events, firehoseConfig)
	go firehoseClient.Start()

	defer firehoseClient.Start()

	defer cachingClient.Close()

}
