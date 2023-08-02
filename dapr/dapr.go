package dapr

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/controllers/trans"
	"github.com/mdaxf/iac/logger"
)

func StartDaprServer() {
	iLog := &logger.Log{ModuleName: logger.API, User: "System", ControllerName: "Dapr"}
	iLog.Debug("Start Dapr Server")

	configFile := "daprconfig.json"

	// Read the configuration file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading configuration file: %v", err))
		//log.Fatalf("Error reading configuration file: %v", err)
		return
	}

	var topicsConfig struct {
		Host   string `json:"host"`
		Topics []struct {
			Name    string `json:"name"`
			Handler string `json:"handler"`
		} `json:"topics"`
	}

	if err := json.Unmarshal(configData, &topicsConfig); err != nil {
		iLog.Error(fmt.Sprintf("Error parsing configuration file: %v", err))
		//log.Fatalf("Error parsing configuration file: %v", err)
		return
	}
	if topicsConfig.Host == "" {
		iLog.Error("Missing host in configuration file")
		//log.Fatal("Missing host in configuration file")
		return
	}
	// Start listening to topics and execute handlers
	for _, topic := range topicsConfig.Topics {
		go func(topicName, handlerName string) {
			url := topicsConfig.Host + topicName
			iLog.Debug(fmt.Sprintf("Listening to topic %s", topicName))
			for {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					iLog.Error(fmt.Sprintf("Error while creating request for topic %s: %v", topicName, err))
					//log.Printf("Error while creating request for topic %s: %v", topicName, err)
					time.Sleep(10 * time.Second)
					continue
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					iLog.Error(fmt.Sprintf("Error while receiving message from topic %s: %v", topicName, err))
					//log.Printf("Error while receiving message from topic %s: %v", topicName, err)
					continue
				}
				defer resp.Body.Close()
				iLog.Debug(fmt.Sprintf("Received message from topic %s :%s", topicName, resp.Body))
				var data []byte
				if resp.StatusCode == http.StatusOK {
					data, err = io.ReadAll(resp.Body)
					if err != nil {
						iLog.Error(fmt.Sprintf("Error while reading message from topic %s: %v", topicName, err))
						//log.Printf("Error while reading message from topic %s: %v", topicName, err)
						continue
					}
					iLog.Debug(fmt.Sprintf("Received message from topic %s: %s", topicName, data))
					//log.Printf("Received message from topic %s: %s", topicName, data)
					// Call the corresponding handler function
					_, err = callHandler(handlerName, data)
					if err != nil {
						iLog.Error(fmt.Sprintf("Error while executing handler %s for topic %s: %v", handlerName, topicName, err))
						//log.Printf("Error while executing handler %s for topic %s: %v", handlerName, topicName, err)
					}
				}
			}
		}(topic.Name, topic.Handler)
	}

	// Block the main goroutine to keep the program running
	select {}
}

// callHandler is a helper function to call the appropriate handler function
func callHandler(handlerName string, data []byte) (map[string]interface{}, error) {

	var jdata map[string]interface{}

	err := json.Unmarshal(data, &jdata)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	tr := &trans.TranCodeController{}
	outputs, err := tr.Execute(handlerName, jdata)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return outputs, nil
}
