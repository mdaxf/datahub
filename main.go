package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/mdaxf/datahub/dapr"
	//	"github.com/mdaxf/datahub/mqtt"
	dbconn "github.com/mdaxf/datahub/database"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/integration/mqttclient"
	"github.com/mdaxf/iac/integration/signalr"
	"github.com/mdaxf/iac/logger"
	ssrv "github.com/mdaxf/signalrsrv/signalr"
)

type Configuration struct {
	Documentdb    Documentdb             `json:"documentdb"`
	Log           map[string]interface{} `json:"log"`
	Database      map[string]interface{} `json:"database"`
	SingalRConfig map[string]interface{} `json:"singalrconfig"`
}
type Documentdb struct {
	DatabaseType       string `json:"type"`
	DatabaseConnection string `json:"connection"`
	DatabaseName       string `json:"database"`
}

var wg sync.WaitGroup

var DocDBconn *documents.DocDB

var SignalRClient ssrv.Client

var ilog logger.Log

func main() {
	var err error
	config := getConfiguration()
	logger.Init(config.Log)

	ilog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Data Hub"}
	ilog.Debug("initialize Data Hub")

	ilog.Debug(fmt.Sprintf("configuration: %v", logger.ConvertJson(config)))

	ilog.Debug(fmt.Sprintf("Database connection config:%v", config.Database))

	err = dbconn.Initialize(config.Database)
	if err != nil {
		ilog.Error(fmt.Sprintf("failed to initialize Database: %v", err))
	}

	ilog.Debug(fmt.Sprintf("%v", config.Documentdb))
	// ...

	ilog.Debug(fmt.Sprintf("initialize Database with %s, %s, %s", config.Documentdb.DatabaseType, config.Documentdb.DatabaseConnection, config.Documentdb.DatabaseName))

	DocDBconn, err = documents.InitMongDB(config.Documentdb.DatabaseConnection, config.Documentdb.DatabaseName)
	if err != nil {
		ilog.Error(fmt.Sprintf("failed to initialize Database: %v", err))
	}

	if dbconn.DB != nil {
		defer dbconn.DB.Close()
	}

	initializesignalRClient(config.SingalRConfig)

	//initializeDaprClient()

	initializeMqttClient()

	// Close the DocDB connection
	if DocDBconn.MongoDBClient != nil {
		defer DocDBconn.MongoDBClient.Disconnect(context.Background())
	}

	wg.Wait()
}

func getConfiguration() Configuration {

	data, err := ioutil.ReadFile("configuration.json")
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to read configuration file: %v", err))

	}
	fmt.Println(fmt.Sprintf("configuration file: %s", string(data)))
	var config Configuration
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to unmarshal the configuration file: %v", err))

	}
	fmt.Println(fmt.Sprintf("configuration: %v", logger.ConvertJson(config)))
	return config
}

func initializesignalRClient(config map[string]interface{}) {
	wg.Add(1)
	var err error
	ilog.Debug("initialize IAC Message Bus Client")
	go func() {
		defer wg.Done()
		SignalRClient, err = signalr.Connect(config)
		if err != nil {
			//	fmt.Errorf("Failed to connect to IAC Message Bus: %v", err)
			ilog.Error(fmt.Sprintf("Failed to connect to IAC Message Bus: %v", err))
		}
		fmt.Println("IAC Message Bus: ", SignalRClient)
		ilog.Debug(fmt.Sprintf("IAC Message Bus: %v", SignalRClient))
	}()
}

func initializeDaprClient() {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Dapr"}
	ilog.Debug("initialize Dapr Client")
	wg.Add(1)
	go func() {
		defer wg.Done()

		ilog.Debug("initialize Dapr Client")
		dapr.StartDaprServer()

	}()
}

func initializeMqttClient() {
	ilog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Mqtt"}
	ilog.Debug("initialize MQTT Client")

	wg.Add(1)
	go func() {
		defer wg.Done()

		ilog.Debug("initialize MQTT Client")

		data, err := ioutil.ReadFile("mqttconfig.json")
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to read configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("MQTT Clients configuration file: %s", string(data)))
		var mqttconfig mqttclient.MqttConfig
		err = json.Unmarshal(data, &mqttconfig)
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to unmarshal the configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("MQTT Clients configuration: %v", logger.ConvertJson(mqttconfig)))

		for _, mqttcfg := range mqttconfig.Mqtts {
			ilog.Debug(fmt.Sprintf("MQTT Client configuration: %s", logger.ConvertJson(mqttcfg)))
			ilog.Debug(fmt.Sprintf("parameters: %s, %s, %s", dbconn.DB, DocDBconn, SignalRClient))
			mqttclient.NewMqttClientbyExternal(mqttcfg, dbconn.DB, DocDBconn, SignalRClient)
		}

	}()

}
