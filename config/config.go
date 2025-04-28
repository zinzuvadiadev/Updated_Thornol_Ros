package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"strings"
	"time"
)

const (
	defaultLogLevel    = slog.LevelDebug
	defaultServiceName = "thornol"
	defaultVersion     = "1.0.0"
	defaultConfigDir   = "./"

	environmentConfKey = "GO_ENV"
	logLevelConfKey    = "LOG_LEVEL"
	serviceNameConfKey = "SERVICE_NAME"
	versionConfKey     = "VERSION"
	configDirConfKey   = "CONFIG_DIR"

	systemStatsIntervalConfKey   = "SYSTEM_STATS_INTERVAL"
	netStatsIntervalConfKey      = "NET_STATS_INTERVAL"
	speedTestIntervalConfKey     = "SPEED_TEST_INTERVAL"
	navStackMapFolderPathConfKey = "NAVSTACK_MAP_FOLDER_PATH"
	apiKeyConfKey                = "API_KEY"
	apiEndpointConfKey           = "API_ENDPOINT"
	orgIdConfKey                 = "ORG_ID"
)

type Config struct {
	// MQTTConf MQTTConfig
	Environment           string
	LogLevel              int
	ServiceName           string
	Version               string
	ConfigDir             string
	SystemStatsInterval   time.Duration
	NetStatsInterval      time.Duration
	SpeedTestInterval     time.Duration
	NavStackMapFolderPath string
	ApiKey                string
	ApiEndpoint           string
	OrgId                 string
	TopicMaps             TopicMaps
}

type MQTTConfig struct {
	DeviceID        string `json:"device_id"`
	DeviceName      string `json:"device_name"`
	Host            string `json:"host"`
	Port            string `json:"port"`
	MetadataTopic   string `json:"metadata_topic"`
	ObjectPushTopic string `json:"object_push_topic"`
	RootTopic       string `json:"root_topic"`

	TLSConfig         *tls.Config
	BaseTopic         string
	ShadowTopic       string
	RPCTopic          string
	LogsTopic         string
	DataTopic         string
	ShadowPingTopic   string
	ShadowReadTopic   string
	ShadowUpdateTopic string
	SysStatsTopic     string
	NetStatsTopic     string
	IfaceStatsTopic   string
	SpeedTestTopic    string

	//	Webrtc Topics
	RTCHealthCheckTopic        string
	RTCConnectTopic            string
	RTCDeviceCapabilitiesTopic string
}

type Link struct {
	NIC           string `json:"nic"`
	ISPName       string `json:"isp_name"`
	ISPSPOCEmail  string `json:"isp_spoc_email"`
	SPOCName      string `json:"isp_spoc_name"`
	ISPSPOCNumber string `json:"isp_spoc_number"`
	Plan          string `json:"plan"`
	Link          string `json:"link"`
	StaticIP      string `json:"static_ip"`
	Avail         string `json:"static_ip_available"`
}

type UCIConfig struct {
	Link []Link `json:"link"`
}

var Conf *Config
var MQTTConf *MQTTConfig
var UCIConf UCIConfig

type SubscriptionKey string
type PublicationKey string

const (
	Odometry           SubscriptionKey = "odometry"
	Map                SubscriptionKey = "map"
	LaserScan          SubscriptionKey = "laser_scan"
	Path               SubscriptionKey = "plan"
	LocalPath          SubscriptionKey = "local_plan"
	MapMetadata        SubscriptionKey = "map_metadata"
	NavigationFeedback SubscriptionKey = "navigation_feedback"
	GlobalCostmap      SubscriptionKey = "global_costmap"
	Soc                SubscriptionKey = "soc"
	Soh                SubscriptionKey = "soh"
	AMCLPose           SubscriptionKey = "amcl_pose"
)

var MandatoryTopics = []SubscriptionKey{
	Odometry,
	Map,
	LaserScan,
	Path,
	LocalPath,
	MapMetadata,
	NavigationFeedback,
	GlobalCostmap,
	Soc,
	Soh,
	AMCLPose,
}

const (
	CmdVel        PublicationKey = "cmd_vel"
	InitialPose   PublicationKey = "initialpose"
	WaypointsList PublicationKey = "waypoints_list"
	GoalPose      PublicationKey = "goal_pose"
)

type TopicConfig struct {
	Topic    string `json:"topic"`
	Throttle int    `json:"throttle"`
}

type TopicMaps struct {
	Subscriptions map[SubscriptionKey]TopicConfig `json:"subscriptions"`
	Publications  map[PublicationKey]TopicConfig  `json:"publications"`
}

func New() (*Config, error) {
	vars := &confVars{}

	environment := vars.mandatory(environmentConfKey)
	serviceName := vars.optional(serviceNameConfKey, defaultServiceName)
	version := vars.optional(versionConfKey, defaultVersion)
	configDir := vars.optional(configDirConfKey, defaultConfigDir)
	systemStatsInterval := vars.optionalDuration(systemStatsIntervalConfKey, 5*time.Second)
	netStatsInterval := vars.optionalDuration(netStatsIntervalConfKey, 5*time.Second)
	speedTestInterval := vars.optionalDuration(speedTestIntervalConfKey, 1*time.Hour)

	navStackMapFolderPath := vars.mandatory(navStackMapFolderPathConfKey)
	apiKey := vars.mandatory(apiKeyConfKey)
	apiEndpoint := vars.mandatory(apiEndpointConfKey)
	orgId := vars.mandatory(orgIdConfKey)
	config := &Config{
		// MQTTConf: *MQTTConf,
		Environment:           environment,
		ServiceName:           serviceName,
		Version:               version,
		ConfigDir:             configDir,
		SystemStatsInterval:   systemStatsInterval,
		NetStatsInterval:      netStatsInterval,
		SpeedTestInterval:     speedTestInterval,
		NavStackMapFolderPath: navStackMapFolderPath,
		ApiKey:                apiKey,
		ApiEndpoint:           apiEndpoint,
		OrgId:                 orgId,
	}

	// Load topic maps from CONFIG_DIR
	topicMapsPath := fmt.Sprintf("%s%s", configDir, "topic_maps.json")
	jsonSettings, err := os.ReadFile(topicMapsPath)
	if err != nil {
		// Provide default topic maps if file doesn't exist
		defaultTopicMaps := TopicMaps{
			Subscriptions: map[SubscriptionKey]TopicConfig{
				Odometry:           {Topic: "odom", Throttle: 100},
				Map:                {Topic: "map", Throttle: 0},
				LaserScan:          {Topic: "/scan", Throttle: 0},
				Path:               {Topic: "plan", Throttle: 0},
				LocalPath:          {Topic: "local_plan", Throttle: 0},
				MapMetadata:        {Topic: "map_metadata", Throttle: 0},
				NavigationFeedback: {Topic: "navigation_feedback", Throttle: 0},
				GlobalCostmap:      {Topic: "global_costmap/costmap", Throttle: 0},
				Soc:                {Topic: "soc", Throttle: 0},
				Soh:                {Topic: "soh", Throttle: 0},
				AMCLPose:           {Topic: "amcl_pose", Throttle: 0},
			},
			Publications: map[PublicationKey]TopicConfig{
				CmdVel:        {Topic: "/cmd_vel"},
				InitialPose:   {Topic: "/initialpose"},
				WaypointsList: {Topic: "/waypoints_list"},
				GoalPose:      {Topic: "/goal_pose"},
			},
		}
		config.TopicMaps = defaultTopicMaps
		slog.Info("Loaded default topic maps", "topic_maps", config.TopicMaps)
	} else {
		if err := json.Unmarshal(jsonSettings, &config.TopicMaps); err != nil {
			return nil, fmt.Errorf("failed to parse topic maps: %v", err)
		}

		// Check if all mandatory topics are present
		for _, topic := range MandatoryTopics {
			if _, ok := config.TopicMaps.Subscriptions[topic]; !ok {
				return nil, fmt.Errorf("mandatory topic %s is missing", topic)
			}
		}

		slog.Info("Loaded topic maps", "topic_maps", config.TopicMaps)
	}

	if err := vars.Error(); err != nil {
		return nil, fmt.Errorf("Config: environment variables: %v", err)
	}

	Conf = config

	return config, nil
}

func NewMQTTConfig(c *Config) (*MQTTConfig, error) {
	// load mqtt base config from file
	MQTTConf = &MQTTConfig{}
	jsonSettings, err := os.ReadFile(
		fmt.Sprintf("%s%s", c.ConfigDir, "connection_settings.json"),
	)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonSettings, &MQTTConf)
	if err != nil {
		return nil, err
	}

	// load root ca
	certpool := x509.NewCertPool()
	ca, err := os.ReadFile(
		fmt.Sprintf("%s%s", c.ConfigDir, "certs/root_ca_cert.pem"),
	)
	if err != nil {
		return nil, err
	}
	certpool.AppendCertsFromPEM(ca)

	// load client cert pair
	clientKeyPair, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s%s", c.ConfigDir, "certs/device_cert.pem"),
		fmt.Sprintf("%s%s", c.ConfigDir, "certs/device_private_key.pem"),
	)
	if err != nil {
		return nil, err
	}

	// create tls config
	tlsConfig := &tls.Config{
		RootCAs:            certpool,
		Certificates:       []tls.Certificate{clientKeyPair},
		ServerName:         MQTTConf.Host,
		InsecureSkipVerify: true,
	}
	MQTTConf.TLSConfig = tlsConfig

	MQTTConf.BaseTopic = fmt.Sprintf("%s%s", MQTTConf.RootTopic, MQTTConf.DeviceName)
	MQTTConf.ShadowTopic = fmt.Sprintf("%s/device-shadow", MQTTConf.BaseTopic)
	MQTTConf.RPCTopic = fmt.Sprintf("%s/reflection", MQTTConf.BaseTopic)
	MQTTConf.ObjectPushTopic = fmt.Sprintf("%s/object/+", MQTTConf.BaseTopic)
	MQTTConf.LogsTopic = fmt.Sprintf("%s/device-logs", MQTTConf.BaseTopic)
	MQTTConf.DataTopic = fmt.Sprintf("%s/device-data", MQTTConf.BaseTopic)
	MQTTConf.ShadowPingTopic = fmt.Sprintf("%s/p", MQTTConf.ShadowTopic)
	MQTTConf.ShadowReadTopic = fmt.Sprintf("%s/r", MQTTConf.ShadowTopic)
	MQTTConf.ShadowUpdateTopic = fmt.Sprintf("%s/u", MQTTConf.ShadowTopic)
	MQTTConf.SysStatsTopic = fmt.Sprintf("%s/system", MQTTConf.DataTopic)
	MQTTConf.NetStatsTopic = fmt.Sprintf("%s/network", MQTTConf.DataTopic)
	MQTTConf.IfaceStatsTopic = fmt.Sprintf("%s/iface-stats", MQTTConf.DataTopic)
	MQTTConf.SpeedTestTopic = fmt.Sprintf("%s/SpeedTest", MQTTConf.DataTopic)
	MQTTConf.RTCHealthCheckTopic = fmt.Sprintf("%s/rtc/healthcheck/r", MQTTConf.BaseTopic)
	MQTTConf.RTCDeviceCapabilitiesTopic = fmt.Sprintf("%s/rtc/capabilities/request", MQTTConf.BaseTopic)
	MQTTConf.RTCConnectTopic = fmt.Sprintf("%s/rtc/token/r", MQTTConf.BaseTopic)
	MQTTConf.MetadataTopic = fmt.Sprintf("%s/device-metadata/patch", MQTTConf.BaseTopic)
	return MQTTConf, nil
}

type confVars struct {
	missing   []string //name of the mandatory environment variable that are missing
	malformed []string //errors describing malformed environment varibale values
}

func (vars confVars) optional(key, fallback string) string {
	val := os.Getenv(key)

	if val == "" {
		return fallback
	}
	return val
}

func (vars *confVars) optionalInt(key string, fallback int) int {
	valStr := os.Getenv(key)

	if valStr == "" {
		return fallback
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("optional %s (value=%q) is not a number", key, valStr))
		return fallback
	}

	return val
}

func (vars *confVars) optionalBool(key string, fallback bool) bool {
	valStr := os.Getenv(key)

	if valStr == "" {
		return fallback
	}

	val, err := strconv.ParseBool(valStr)
	if err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("optional %s (value=%q) is not a boolean", key, valStr))
		return fallback
	}

	return val
}

func (vars *confVars) mandatory(key string) string {
	val := os.Getenv(key)

	if val == "" {
		vars.missing = append(vars.missing, key)
		return ""
	}

	return val
}

func (vars *confVars) optionalDuration(key string, fallback time.Duration) time.Duration {
	valStr := os.Getenv(key)

	if valStr == "" {
		return fallback
	}

	duration, err := time.ParseDuration(valStr)
	if err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("mandatory %s (value=%q) is not a Duration", key, valStr))
		return 0
	}

	return duration
}

func (vars *confVars) mandatoryInt(key string) int {
	valStr := vars.mandatory(key)

	val, err := strconv.Atoi(valStr)
	if err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("mandatory %s (value=%q) is not a boolean", key, valStr))
		return 0
	}

	return val
}

func (vars *confVars) mandatoryBool(key string) bool {
	valStr := vars.mandatory(key)

	val, err := strconv.ParseBool(valStr)
	if err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("mandatory %s (value=%q) is not a boolean", key, valStr))
		return false
	}

	return val
}

func (vars *confVars) mandatoryJSON(key string, value interface{}) {
	valStr := os.Getenv(key)

	if valStr == "" {
		vars.missing = append(vars.missing, key)
		return
	}

	if err := json.Unmarshal([]byte(valStr), value); err != nil {
		vars.malformed = append(vars.malformed, fmt.Sprintf("mandatory %s (value=%q) is not a JSON", key, valStr))
	}
}

func (vars confVars) Error() error {
	if len(vars.missing) > 0 {
		return fmt.Errorf("missing mandatory configurations: %s", strings.Join(vars.missing, ", "))
	}

	if len(vars.malformed) > 0 {
		return fmt.Errorf("malformed configurations: %s", strings.Join(vars.malformed, "; "))
	}
	return nil
}
