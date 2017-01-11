# cloudfoundry-sumologic-nozzle

This Nozzle aggregates all the events from the _Firehose_ feature in Cloud Foundry towards Sumo Logic

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
  --events="LogMessage"          Comma separated list of events you would like. Valid options are ContainerMetric,
                                 CounterEvent, Error, HttpStart, HttpStartStop, HttpStop, LogMessage, ValueMetric
  --nozzle-polling-period=15s    Nozzle Polling Period
  --log-events-batch-size=200    Log Events Batch Size to send to Sumo
  --sumo-post-minimum-delay=200ms  
                                 Sumo Logic HTTP Post Minimum Delay
  --sumo-category=""             Sumo Logic Category
  --sumo-name=""                 Sumo Logic Name
  --sumo-host=""                 Sumo Logic Host
  --verbose-log-messages         Allow Verbose Log Messages
  --custom-metadata=""           Use this flag for addingCustom Metadata (key1:value1,key2:value2, etc...)
  --include-only-matching-filter=""  
                                 Adds an 'Include only' filter to Events content (key1:value1,key2:value2, etc...)
  --exclude-always-matching-filter=""  
                                 Adds an 'Exclude always' filter to Events content (key1:value1,key2:value2,
                                 etc...)
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


### Filtering Option

asked. It works this way:
* case 1:
Include-Only filter="" (Empty)
Exclude-Always filter="" (Empty)
in this case, all the events will be sent to Sumo Logic

* case 2:
Include-Only filter="" (Empty)
Exclude-Always filter= "source_type:other,origin:rep",
in this case, all the events that contains a source-type:other field OR an origin:rep field will be not sent to Sumo Logic

* case 3:
Include-Only filter="job:diego_cell,source_type:other"
Exclude-Always filter="" (Empty)
in this case, Only the events that contains a job:diego-cell field OR a source-type:other field will be sent to Sumo Logic

* case 4:
Include-Only filter="job:diego_cell,source_type:other"
Exclude-Always filter="source_type:other,origin:rep"

in this case, all the events that contains a job:diego-cell field OR a source-type:other field will be sent to Sumo Logic
AND also
all the events that contains a source-type:other field OR an origin:rep field will be not sent to Sumo Logic
if an event share both filters (contains a Include-Only filter field and a Exclude-Always filter field), Only the *Include-Only filter will be considered*

The correct way of using those flags will be something like this:

```
godep go run main.go --sumo-endpoint=https://sumo-endpoint --api-endpoint=https://api.endpoint --cloudfoundry-user=some_user --cloudfoundry-password=some_password --sumo-post-minimum-delay=200ms --log-events-batch-size=200 --events=LogMessage, ValueMetric   --include-only-matching-filter=job:diego_cell,source_type:app --exclude-always-matching-filter=source_type:other,unit:count
```


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
