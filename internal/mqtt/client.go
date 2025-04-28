package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	Conf "thornol/config"

	"thornol/internal/livekit"
	PM "thornol/internal/protos"

	bridge "github.com/golain-io/mqtt-bridge"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pion/mediadevices"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/protobuf/proto"
)

type Capabilities struct {
	AudioTracks []mediadevices.Track `json:"audio_tracks"`
	VideoTracks []mediadevices.Track `json:"video_tracks"`
}

type RTCDeviceCapabilitiesResponse struct {
	DeviceId           string       `json:"device_id"`
	DeviceCapabilities Capabilities `json:"device_capabilities"`
}

type GolainClient struct {
	Client                mqtt.Client
	BridgeClient          *bridge.MQTTNetBridge
	ShadowHandler         func([]byte) error
	OnConnectHandler      func()
	ObjectPushHandler     func(topic string, payload []byte) error
	NavStackMapFolderPath string
	ApiKey                string
	ApiEndpoint           string
	OrgId                 string
}

var DeviceClient *GolainClient

func NewMQTTClient(C *Conf.MQTTConfig) (*GolainClient, error) {
	if DeviceClient != nil {
		return DeviceClient, nil
	}
	DeviceClient = &GolainClient{
		ShadowHandler: defaultShadowHandler,
	}

	options := mqtt.NewClientOptions()
	options.TLSConfig = C.TLSConfig
	options.AddBroker(fmt.Sprintf("%s://%s:%s", "ssl", C.Host, C.Port))
	// options.AddBroker(fmt.Sprintf("tcp://%s:%s", C.Host, C.Port))
	options.SetClientID(C.DeviceName)
	options.AutoReconnect = false
	// options.OnConnect = defaultConnectHandler
	options.OnConnectionLost = connectionLostHandler
	options.OnConnect = defaultConnectHandler
	// options.OnConnectionLost = defaultConnectionLostHandler

	DeviceClient.Client = mqtt.NewClient(options)

	// Start a goroutine for connection attempts with backoff
	go func() {
		backoff := time.Second
		for {
			token := DeviceClient.Client.Connect()
			if token.Wait() && token.Error() != nil {
				slog.Error("MQTT connection error", "error", token.Error())
				// Random backoff between 1 and 2 times the current backoff
				jitter := time.Duration(rand.Float64() * float64(backoff))
				sleepTime := backoff + jitter
				slog.Info("Retrying connection", "delay", sleepTime)
				time.Sleep(sleepTime)
				// Increase backoff, cap at 1 minute
				backoff = time.Duration(float64(backoff) * 1.5)
				if backoff > time.Minute {
					backoff = time.Minute
				}
			} else {
				slog.Info("MQTT client Connected")
				return
			}
		}
	}()

	return DeviceClient, nil
}

func defaultConnectHandler(client mqtt.Client) {
	if DeviceClient.OnConnectHandler != nil {
		slog.Info("Running on connect handler")
		DeviceClient.OnConnectHandler()
	}
	slog.Info("MQTT Client Connected")
	if err := DeviceClient.defaultSubscriptions(); err != nil {
		slog.Error("Failed to set default subscriptions", "error", err)
	}

}

func defaultConnectionLostHandler(client mqtt.Client, err error) {
	slog.Info("MQTT Client Connection Lost")
}

func connectionLostHandler(client mqtt.Client, err error) {
	slog.Error("MQTT client Connection Lost", "error", err)
	// Trigger reconnection
	go func() {
		backoff := time.Second
		for {
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				slog.Error("MQTT reconnection error", "error", token.Error())
				jitter := time.Duration(rand.Float64() * float64(backoff))
				sleepTime := backoff + jitter
				slog.Info("Retrying connection", "delay", sleepTime)
				time.Sleep(sleepTime)
				backoff = time.Duration(float64(backoff) * 1.5)
				if backoff > time.Minute {
					backoff = time.Minute
				}
			} else {
				if client.IsConnected() && client.IsConnectionOpen() {
					slog.Info("MQTT client Reconnected")
					return
				} else {
					client.Disconnect(0)
				}
			}
		}
	}()
}

func defaultShadowHandler([]byte) error {
	slog.Warn("Shadow handler not set")
	return nil
}

func (c *GolainClient) shadowUpdateHandler(client mqtt.Client, message mqtt.Message) {
	slog.Info("Shadow update received", "topic", message.Topic(), "payload", string(message.Payload()))
	c.ShadowHandler(message.Payload())
}

func (c *GolainClient) RegisterBridge(bridgeID, bridgeType string) error {
	token := c.Client.Publish(fmt.Sprintf("%s/register/bridge", Conf.MQTTConf.BaseTopic), 2, false, []byte(fmt.Sprintf("%s:%s", bridgeID, bridgeType)))
	token.WaitTimeout(5 * time.Second)
	return token.Error()
}

func (c *GolainClient) livekitConnectHandler(client mqtt.Client, message mqtt.Message) {

	data := &livekit.RTCJoinRoomRequest{}
	err := json.Unmarshal(message.Payload(), data)
	if err != nil {
		otelzap.L().Sugar().Error("Not able to unmarshal payload", err.Error())
	}
	fmt.Printf("Data: %v\n", data)
	started := make(chan bool)
	go func() {
		if err := livekit.JoinRoom(data, started); err != nil {
			slog.Error("Failed to join room", "error", err)
		}
	}()

	if <-started {
		slog.Info("Room started")
		if err := c.Publish(fmt.Sprintf("%s/rtc/device/room/r", Conf.MQTTConf.BaseTopic), message.Payload()); err != nil {
			otelzap.L().Sugar().Error("Error publishing livekit connect", err.Error())
		}
	}
}

func (c *GolainClient) processDeviceCapabilitesCheck(client mqtt.Client, message mqtt.Message) {
	fmt.Printf("Processing device capabilities check")
	data := &livekit.RTCDeviceCapabilitiesRequest{}
	err := json.Unmarshal(message.Payload(), data)
	if err != nil {
		otelzap.L().Sugar().Error("Not able to unmarshal payload", err.Error())
	}
	capabilties, err := livekit.RTCDeviceCapabilitiesCheck(data)

	if err != nil {
		otelzap.L().Sugar().Error("Error in device capabilities check", err.Error())
		return
	}
	payload, err := json.Marshal(capabilties)
	if err != nil {
		otelzap.L().Sugar().Error("Not able to marshal payload", err.Error())
		return
	}

	if err := c.Publish(fmt.Sprintf("%s/rtc/capabilities/response", Conf.MQTTConf.BaseTopic), payload); err != nil {
		otelzap.L().Sugar().Error("Error publishing device capabilities", err.Error())
	}
}

func (c *GolainClient) processObjectPush(client mqtt.Client, message mqtt.Message) {
	slog.Info("Object push received", "topic", message.Topic(), "payload", string(message.Payload()))

	if c.ObjectPushHandler != nil {
		c.ObjectPushHandler(message.Topic(), message.Payload())
	}

}

func (c *GolainClient) Publish(topic string, payload []byte) error {
	token := c.Client.Publish(topic, 1, true, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *GolainClient) Subscribe(topic string, callback mqtt.MessageHandler) error {
	token := c.Client.Subscribe(topic, 1, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *GolainClient) SubscribeWithQoS(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := c.Client.Subscribe(topic, qos, callback)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (c *GolainClient) PublishLog(l PM.PLog, ctx context.Context) error {
	b, err := proto.Marshal(&l)
	if err != nil {
		return err
	}
	if err := c.Publish(Conf.MQTTConf.LogsTopic, b); err != nil {
		return err
	}
	return nil
}

func (c *GolainClient) defaultSubscriptions() error {
	// subscribe to shadow updates
	fmt.Printf("Subscribing to shadow update topic: %s\n", Conf.MQTTConf.ShadowReadTopic)
	if err := c.Subscribe(Conf.MQTTConf.ShadowReadTopic, c.shadowUpdateHandler); err != nil {
		return err
	}
	fmt.Printf("Publishing to shadow ping topic: %s\n", Conf.MQTTConf.ShadowPingTopic)
	// ping for update
	if err := c.Publish(Conf.MQTTConf.ShadowPingTopic, []byte("1")); err != nil {
		slog.Error("Error pinging for shadow update")
	}
	// if err := c.Subscribe(Conf.MQTTConf.RTCHealthCheckTopic, c.healthCheckHandler); err != nil {
	// 	return err
	// }
	fmt.Printf("Subscribing to RTC connect: %s\n", Conf.MQTTConf.RTCConnectTopic)
	if err := c.Subscribe(Conf.MQTTConf.RTCConnectTopic, c.livekitConnectHandler); err != nil {
		return err
	}
	if err := c.Subscribe(Conf.MQTTConf.RTCDeviceCapabilitiesTopic, c.processDeviceCapabilitesCheck); err != nil {
		return err
	}

	fmt.Printf("Subscribing to object push topic: %s\n", Conf.MQTTConf.ObjectPushTopic)
	if err := c.SubscribeWithQoS(Conf.MQTTConf.ObjectPushTopic, 2, c.processObjectPush); err != nil {
		return err
	}

	return nil
}
