package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	Config "thornol/config"

	// "thornol/internal/bridge"
	bridge "github.com/golain-io/mqtt-bridge"

	mission_executor "thornol/internal/mission_executor"
	mqtt "thornol/internal/mqtt"
	PM "thornol/internal/protos"
	"thornol/internal/server"
	shadow "thornol/internal/shadow"
	utils "thornol/internal/utils"
)

var mqttClient mqtt.GolainClient
var bridgeClient *bridge.MQTTNetBridge

func objectPushHandler(topic string, payload []byte) error {

	topicElements := strings.Split(topic, "/")
	lastTopicElement := topicElements[len(topicElements)-1]

	switch lastTopicElement {
	case "map_collection":
		var objPayload utils.ObjectPushPayload
		if err := json.Unmarshal(payload, &objPayload); err != nil {
			slog.Error("Failed to unmarshal object push payload", "error", err)
			return err
		}

		slog.Info("Received object push payload", "payload", objPayload)
		// Download and copy map files to the navigation stack folder
		if err := utils.CopyMapFiles(&objPayload, mqttClient.NavStackMapFolderPath); err != nil {
			slog.Error("Failed to copy map files", "error", err)
			return err
		}

		slog.Info("Successfully copied map files",
			"object_id", objPayload.ObjectID,
			"destination", mqttClient.NavStackMapFolderPath)

		if mapID, ok := objPayload.Metadata["map_id"].(string); ok && mapID != "" {
			shadow.GlobalShadow.MapId = mapID
			// Save metadata to golain.json
			if err := utils.SaveMapMetadata(mapID, mqttClient.NavStackMapFolderPath); err != nil {
				slog.Error("Failed to save map metadata", "error", err)
			}
		}
		shadow.GlobalShadow.MapFound = true
		shadow.GlobalShadow.UpdateShadow()

	default:
		slog.Warn("Unknown topic", "topic", topic)
	}
	return nil
}

func StartMQTTClient(conf *Config.Config) {
	mqttConf, err := Config.NewMQTTConfig(conf)
	if err != nil {
		slog.Error("Error loading mqtt configuration: %v", err)
		os.Exit(1)
	}
	client, err := mqtt.NewMQTTClient(mqttConf)
	if err != nil {
		slog.Error("Error creating mqtt client: %v", err)
		os.Exit(1)
	}
	client.NavStackMapFolderPath = conf.NavStackMapFolderPath
	client.ApiKey = conf.ApiKey
	client.ApiEndpoint = conf.ApiEndpoint
	client.OrgId = conf.OrgId
	client.ShadowHandler = func(buffer []byte) error {
		slog.Info("received shadow update", "buffer", buffer)
		return shadow.GlobalShadow.ShadowCallback(buffer)
	}

	client.OnConnectHandler = func() {
		mission_executor.Init(client)
		mission_executor.GlobalScheduler.Start()
		bridgeClient = server.StartRPCService(mqttClient)
	}
	client.ObjectPushHandler = objectPushHandler
	mqttClient = *client
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "debug_subscriptions":
			// Load env file
			if godotenv.Load(".env") != nil {
				slog.Warn("Error while loading environment variables, using pre-existing ones")
			}

			// Load configuration
			conf, err := Config.New()
			if err != nil {
				slog.Error("Error loading configuration: %v", err)
				os.Exit(1)
			}

			// Pass configuration to run function
			if err := run(conf); err != nil {
				slog.Error("Error running subscription debug: %v", err)
				os.Exit(1)
			}
			return
		}
	}
	thornol_main()
}

func thornol_main() {
	// load env file
	if godotenv.Load(".env") != nil {
		slog.Warn("Error while loading environment variables, using pre-existing ones")
	}

	// Check if necessary environment variables were loader/pre existed
	if _, ok := os.LookupEnv("GO_ENV"); !ok {
		slog.Error("Error while loading environment variables!")
	}

	conf, err := Config.New()
	if err != nil {
		slog.Error("Error loading configuration: %v", err)
		os.Exit(1)
	}

	// setup default logger
	slog.SetDefault(
		slog.New(slog.NewTextHandler(
			io.Writer(os.Stderr),
			&slog.HandlerOptions{
				Level: slog.Level(conf.LogLevel),
			},
		),
		),
	)

	shadow.GlobalShadow.SetFirstShadowUpdateCallback(nil)

	StartMQTTClient(conf)

	shadow.GlobalShadow.SetUpdateShadowCallback(func(current, prev *PM.Shadow) {
		slog.Info("shadow updated", "current", current, "previous", prev)
	})

	defer mission_executor.GlobalScheduler.Stop()

	defer bridgeClient.Close()

	time.Sleep(2 * time.Second)
	// fmt.Println("Publishing to reflection topic")
	// mqttClient.Publish(fmt.Sprintf("%s/reflection/thornol", Config.MQTTConf.BaseTopic), []byte("Hello from thornol"))

	// mqttClient.RegisterBridge("thornol", "grpc")

	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	}()
	<-c
	os.Exit(1)
}
