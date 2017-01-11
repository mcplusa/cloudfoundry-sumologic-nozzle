# cloudfoundry-sumologic-nozzle

This Nozzle aggregates all the events from the _Firehose_ feature in Cloud Foundry towards Sumo Logic

## Getting Started

### Options of use

```
usage: main [<flags>]

Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --api-endpoint=API-ENDPOINT    CF API Endpoint
  --sumo-endpoint=SUMO-ENDPOINT  Sumo Logic Endpoint
  --subscription-id="firehose"   Cloud Foundry ID for the subscription.
  --cloudfoundry-user=CLOUDFOUNDRY-USER  
                                 Cloud Foundry User
  --cloudfoundry-password=CLOUDFOUNDRY-PASSWORD  
                                 Cloud Foundry Password
  --events="LogMessage"          Comma separated list of events you would like. Valid options are ContainerMetric, CounterEvent, Error, HttpStart,HttpStartStop, HttpStop, LogMessage, ValueMetric
  --nozzle-polling-period=15s    Nozzle Polling Period
  --log-events-batch-size=LOG-EVENTS-BATCH-SIZE  
                                 Log Events Batch Size
  --sumo-post-minimum-delay=SUMO-POST-MINIMUM-DELAY  
                                 Sumo Logic HTTP Post Minimum Delay
  --sumo-category=""             Sumo Logic Category
  --sumo-name=""                 Sumo Logic Name
  --sumo-host=""                 Sumo Logic Host
  --verbose-log-messages         Allow Verbose Log Messages
  --custom-metadata=""           Custom Metadata
  --version                      Show application version.
```

There are two ways to run this Nozzle:

1. Run as standalone app
2. Run as a tile in Pivotal Cloud Foundry

### Run as standalone app

```
godep go run main.go --sumo-endpoint=https://sumo-endpoint --api-endpoint=https://api.endpoint --cloudfoundry-user=some_user --cloudfoundry-password=some_password --sumo-post-minimum-delay=200ms --custom-metadata=Key1:Value1,Key2:Value2,Key3:Value3 --log-events-batch-size=200 --events=LogMessage, ValueMetric --verbose-log-messages
```

If everything goes right, you should see in your terminal the _Nozzle's Logs_ and, in the __Sumo Logic endpoint__ (defined in the _--sumo-endpoint_ flag) you should see the logs according the events you choose (_'LogMessage'_ and _'ValueMetric'_ with _verbose_ in this case).

### Run as a tile in Pivotal Cloud Foundry

**Pivotal Cloud Foundry** (PCF) has a Tile Generator tool which will help you to deploy this Nozzle in PCF allowing an easy configuration of it.

The tile configuration is handled in the 'tile.yml' file. (If you want to modify this file, is worth to mention that it is directly related to the Nozzle flags mentioned before).

#### Steps to run as this Nozzle as a tile in PCF:

 ##### Step 1 - Install the tile-generator python package

* Follow the Official Pivotal Instructions: http://docs.pivotal.io/tiledev/tile-generator.html#how-to
(only until half of the _step 3_, DON'T DO 'tile init', only cd into the 'cloudfoundry-sumologic-nozzle' folder)

 ##### Step 2 - Check the tile file
* If you want to add more settings to the tile or remove some. Check the Official Pivotal Documentation for more options http://docs.pivotal.io/tiledev/tile-generator.html#define

 ##### Step 3 - Prepare your code:
* Zip your entire code and place the zip file into the root directory of the project for which you wish to create a tile. For this tile use this command: (you should do this in a new terminal window)

    ```
    zip -r sumo-logic-nozzle.zip bitbucket-pipelines.yml caching/ ci/ eventQueue/ eventRouting/ events/ firehoseclient/ glide.yaml glide.lock Godeps/ LICENSE logging/ main.go manifest.yml event.db Procfile sumoCFFirehose/ utils/ vendor/
    ```
 ##### Step 4 - Build tile file
* go to the 'tile-generator' terminal window and run

    ```
    $ tile build
    ```
 ##### Step 4 - Install the tile in Pivotal Cloud Foundry
* Login with proper credentials into the OPS Manager and import the .pivotal file created above and wait.
* Then add it to the Installation Dashboard to configure it. You should able to configure the settings created in the tile file.
* Update the changes and you should start to see some logs in the Sumo Logic Endpoint defined.

## Authors

mcplusa.com

## Related Sources

* Firehose-to-syslog Nozzle
