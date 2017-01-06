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
	apiEndpoint  = kingpin.Flag("api-endpoint", "CF API Endpoint").OverrideDefaultFromEnvar("API_ENDPOINT").String()
	sumoEndpoint = kingpin.Flag("sumo-endpoint", "Sumo Logic Endpoint").OverrideDefaultFromEnvar("SUMO_ENDPOINT").String()
	//dopplerEndpoint      = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").String()
	subscriptionId       = kingpin.Flag("subscription-id", "Cloud Foundry ID for the subscription.").Default("firehose").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	user                 = kingpin.Flag("cloudfoundry-user", "Cloud Foundry User").OverrideDefaultFromEnvar("CLOUDFOUNDRY_USER").String()             //user created in CF, authorized to connect the firehose
	password             = kingpin.Flag("cloudfoundry-password", "Cloud Foundry Password").OverrideDefaultFromEnvar("CLOUDFOUNDRY_PASSWORD").String() // password created along with the firehose_user                                                                                                           //kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	keepAlive, errK      = time.ParseDuration("25s")                                                                                                  //default
	wantedEvents         = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).Default("Error, ContainerMetric, HttpStart, HttpStop, HttpStartStop, LogMessage, ValueMetric, CounterEvent").OverrideDefaultFromEnvar("EVENTS").String()
	boltDatabasePath     = "my.db" //TODO remove once database code is removed
	tickerTime           = kingpin.Flag("nozzle-polling-period", "Nozzle Polling Period").Default("15s").OverrideDefaultFromEnvar("NOZZLE_POLLING_PERIOD").Duration()
	eventsBatchSize      = kingpin.Flag("log-events-batch-size", "Log Events Batch Size").OverrideDefaultFromEnvar("LOG_EVENTS_BATCH_SIZE").Int()
	sumoPostMinimumDelay = kingpin.Flag("sumo-post-minimum-delay", "Sumo Logic HTTP Post Minimum Delay").OverrideDefaultFromEnvar("SUMO_POST_MINIMUM_DELAY").Duration()
	sumoCategory         = kingpin.Flag("sumo-category", "Sumo Logic Category").Default("").OverrideDefaultFromEnvar("SUMO_CATEGORY").String()
	sumoName             = kingpin.Flag("sumo-name", "Sumo Logic Name").Default("").OverrideDefaultFromEnvar("SUMO_NAME").String()
	sumoClient           = kingpin.Flag("sumo-client", "Sumo Logic Client").Default("").OverrideDefaultFromEnvar("SUMO_CLIENT").String()
)

var (
	version = "0.1"
)

func main() {
	//logging init
	logging.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	kingpin.Version(version)
	kingpin.Parse()

	runtime.GOMAXPROCS(1)

	logging.Info.Println("Set Configurations:")
	logging.Info.Println("CF API Endpoint: " + *apiEndpoint)
	logging.Info.Println("Sumo Logic Endpoint: " + *sumoEndpoint)
	//logging.Info.Println("Cloud foundry Doppler Endpoint: " + *dopplerEndpoint) //TODO
	logging.Info.Println("Cloud Foundry Nozzle Subscription ID: " + *subscriptionId)
	logging.Info.Println("Cloud Foundry User: " + *user)

	logging.Info.Printf("Log Events Batch Size: [%d]\n", *eventsBatchSize)
	logging.Info.Println("Starting Sumo Logic Nozzle " + version)

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: true,
	}
	cfClient, _ := cfclient.NewClient(&c)

	/*if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}*/ //TODO

	//Creating Caching
	var cachingClient caching.Caching
	if caching.IsNeeded(*wantedEvents) {
		cachingClient = caching.NewCachingBolt(cfClient, boltDatabasePath)
	} else {
		cachingClient = caching.NewCachingEmpty()
	}

	logging.Info.Println("Creating queue")
	queue := eventQueue.NewQueue(make([]*events.Event, 100))
	loggingClientSumo := sumoCFFirehose.NewSumoLogicAppender(*sumoEndpoint, 5000, &queue, *eventsBatchSize, *sumoPostMinimumDelay, *sumoCategory, *sumoName, *sumoClient)
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
	//TODO log out each apps name
	logging.Info.Println("Apps founded: ")
	for i := 0; i < len(apps); i++ {
		logging.Info.Printf("[%d] "+apps[i].Name, i+1)
	}
	//Let's start the goRoutine
	cachingClient.PerformPoollingCaching(*tickerTime)

	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   cfClient.Endpoint.DopplerEndpoint,
		InsecureSSLSkipVerify:  true,
		IdleTimeoutSeconds:     keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
	}

	logging.Info.Printf("Connecting to Firehose... \n")

	firehoseClient := firehoseclient.NewFirehoseNozzle(cfClient, events, firehoseConfig)
	go firehoseClient.Start()

	defer firehoseClient.Start()

	defer cachingClient.Close()

}
