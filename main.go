package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/mdaxf/datahub/dapr"
	"github.com/mdaxf/datahub/mqtt"
	"github.com/mdaxf/iac/logger"
)

var wg sync.WaitGroup

func main() {
	logger.Init()

	initializeDaprClient()

	initializeMqttClient()

	wg.Wait()
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
		var mqttconfig mqtt.MqttConfig
		err = json.Unmarshal(data, &mqttconfig)
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to unmarshal the configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("MQTT Clients configuration: %v", logger.ConvertJson(mqttconfig)))

		for _, mqttcfg := range mqttconfig.Mqtts {
			ilog.Debug(fmt.Sprintf("MQTT Client configuration: %s", logger.ConvertJson(mqttcfg)))
			mqtt.NewMqttClient(mqttcfg)

		}

	}()

}
