package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/caching"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventQueue"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/eventRouting"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/firehoseclient"
	"bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/sumoCFFirehose"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug             = true //debug", "Enable debug mode, print in console
	apiEndpoint       = "https://api.bosh-lite.com"
	sumoEndpoint      = kingpin.Flag("sumo-endpoint", "Sumo Endpoint").String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	user              = "firehose_user"     //user created in CF, authorized to connect the firehose
	password          = "firehose_password" // password created along with the firehose_user
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	keepAlive, errK   = time.ParseDuration("25s") //default
	wantedEvents      = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).Default("LogMessage").OverrideDefaultFromEnvar("EVENTS").String()
	boltDatabasePath  = "my.db"                   //default
	tickerTime, errT  = time.ParseDuration("60s") //Default
)

var (
	version = "0.0.0"
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	fmt.Println("this is the sumo endpoint")
	fmt.Println(sumoEndpoint)

	//Creating queue
	queue := eventQueue.NewQueue(make([]*eventQueue.Node, 100))
	loggingClientSumo := sumoCFFirehose.NewSumoLogicAppender(*sumoEndpoint, 1000, *queue)

	fmt.Printf("Starting firehose-to-sumo %s \n", version)

	c := cfclient.Config{
		ApiAddress:        apiEndpoint,
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

	//Creating Events
	events := eventRouting.NewEventRouting(cachingClient, *loggingClientSumo, *queue)
	err := events.SetupEventRouting(*wantedEvents)
	if err != nil {
		log.Fatal("Error setting up event routing: ", err)
		os.Exit(1)

	}

	// Parse extra fields from cmd call
	cachingClient.CreateBucket()
	//Let's Update the database the first time
	fmt.Printf("Start filling app/space/org cache.\n")
	apps := cachingClient.GetAllApp()
	fmt.Printf("Done filling cache! Found [%d] Apps \n", len(apps))

	//Let's start the goRoutine
	cachingClient.PerformPoollingCaching(tickerTime)

	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   cfClient.Endpoint.DopplerEndpoint,
		InsecureSSLSkipVerify:  *skipSSLValidation,
		IdleTimeoutSeconds:     keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
	}

	if /*loggingClientSumo.Connect() ||*/ debug {

		fmt.Printf("Connected to Server! Connecting to Firehose... \n")

		firehoseClient := firehoseclient.NewFirehoseNozzle(cfClient, events, firehoseConfig)
		err = firehoseClient.Start()
		fmt.Printf("I created the Firehose... \n")
		if err != nil {
			fmt.Printf("Failed connecting to Firehose...Please check settings and try again! \n") //Log error

		} else {
			fmt.Printf("Firehose Subscription Succesfull! Routing events... \n")
		}

	} else {
		fmt.Printf("Failed connecting to the Fluentd Server...Please check settings and try again! \n") //Log error
	}

	defer cachingClient.Close()
}
