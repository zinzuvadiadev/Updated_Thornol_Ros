package mission_executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	std_msgs "thornol/internal/rclgo/generated/std_msgs/msg"

	"net/http"
	"thornol/internal/mqtt"
	shadow "thornol/internal/shadow"

	PM "thornol/internal/protos"

	"github.com/tiiuae/rclgo/pkg/rclgo"
)

type MissionDetails struct {
	MapFile   string     `json:"map_file"`
	Waypoints []Waypoint `json:"waypoints"`
	Actions   []Action   `json:"actions"` // TODO: Implement actions
}

type Waypoint struct {
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
	QX float64 `json:"qx"`
	QY float64 `json:"qy"`
	QZ float64 `json:"qz"`
	QW float64 `json:"qw"`
}

type FeedbackType string

const (
	FeedbackType_WAYPOINT_REACHED FeedbackType = "waypoint_reached"
	FeedbackType_WAYPOINT_FAILED  FeedbackType = "waypoint_failed"
)

type Feedback struct {
	Type            FeedbackType `json:"type"`
	CurrentWaypoint int          `json:"current_waypoint"`
	TotalWaypoints  int          `json:"total_waypoints"`
	Timestamp       int64        `json:"timestamp"`
}

type Action struct {
	// TODO: Define action types and parameters
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

type MissionExecutor struct {
	Node             *rclgo.Node
	WaypointsPub     *std_msgs.StringPublisher
	FeedbackSub      *std_msgs.StringSubscription
	PendingWaypoints int
	FeedbackChan     chan FeedbackType
	ApiKey           string
	WaitSet          *rclgo.WaitSet
}

type APIMetaData struct {
	Sequence []SequenceItem `json:"sequence"`
}

// API response structures
type APIMission struct {
	MetaData APIMetaData `json:"meta_data"`
}

type APIResponse struct {
	Mission APIMission `json:"mission"`
}

type SequenceItem struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Waypoint APIWaypoint `json:"waypoint"`
}

type APIWaypoint struct {
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	Orientation struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
		W float64 `json:"w"`
	} `json:"orientation"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom int     `json:"zoom"`
}

func NewMissionExecutor(node *rclgo.Node) *MissionExecutor {

	waypointsPub, err := std_msgs.NewStringPublisher(node, "waypoints_list", nil)
	if err != nil {
		panic(fmt.Errorf("failed to create waypoints publisher: %w", err))
	}

	// Get API key from MQTT client
	apiKey := "" // Default empty
	if mqtt.DeviceClient != nil {
		apiKey = mqtt.DeviceClient.ApiKey
	}

	// Create waitset and add subscriptions
	ws, err := rclgo.NewWaitSet()
	if err != nil {
		panic(fmt.Errorf("failed to create waitset: %w", err))
	}

	missionExecutor := &MissionExecutor{
		Node:         node,
		WaypointsPub: waypointsPub,
		FeedbackChan: make(chan FeedbackType),
		ApiKey:       apiKey,
		WaitSet:      ws,
	}

	feedbackSub, err := std_msgs.NewStringSubscription(node, "navigation_feedback", nil, func(msg *std_msgs.String, info *rclgo.MessageInfo, err error) {
		slog.Info("Received feedback", "feedback", msg.Data)
		missionExecutor.ProcessWaypointFeedback(msg.Data)
	})

	missionExecutor.FeedbackSub = feedbackSub

	if err != nil {
		panic(fmt.Errorf("failed to create feedback subscriber: %w", err))
	}

	go func() {
		// Add subscription to waitset
		missionExecutor.WaitSet.AddSubscriptions(feedbackSub.Subscription)
		missionExecutor.WaitSet.Run(
			context.Background(),
		)
	}()

	return missionExecutor
}

func (e *MissionExecutor) ExecuteMissionDetails(missionId string) error {
	// Read mission details from JSON
	details, err := loadMissionDetails(missionId)
	if err != nil {
		return fmt.Errorf("failed to load mission details: %w", err)
	}

	// Convert waypoints to JSON string
	waypointsJSON, err := json.Marshal(details.Waypoints)
	if err != nil {
		return fmt.Errorf("failed to marshal waypoints: %w", err)
	}

	// Publish waypoints
	slog.Info("Publishing waypoints", "count", len(details.Waypoints))
	if err := e.WaypointsPub.Publish(&std_msgs.String{Data: string(waypointsJSON)}); err != nil {
		return fmt.Errorf("failed to publish waypoints: %w", err)
	}

	// Set initial pending waypoints count
	e.PendingWaypoints = len(details.Waypoints)

	return nil
}

func transformAPIResponseToMissionDetails(apiResp APIResponse) *MissionDetails {
	details := &MissionDetails{
		Waypoints: make([]Waypoint, 0),
	}

	for _, item := range apiResp.Mission.MetaData.Sequence {
		if item.Type == "waypoint" {
			waypoint := Waypoint{
				X:  item.Waypoint.X,
				Y:  item.Waypoint.Y,
				Z:  0,
				QX: item.Waypoint.Orientation.X,
				QY: item.Waypoint.Orientation.Y,
				QZ: item.Waypoint.Orientation.Z,
				QW: item.Waypoint.Orientation.W,
			}
			slog.Info("Waypoint", "waypoint", waypoint)
			details.Waypoints = append(details.Waypoints, waypoint)
		}
	}

	return details
}

func loadMissionDetails(missionId string) (*MissionDetails, error) {
	// If no API key, fall back to file-based loading
	if mqtt.DeviceClient == nil || mqtt.DeviceClient.ApiKey == "" {
		return loadMissionDetailsFromFile(missionId)
	}

	// Make API request
	slog.Info("Fetching mission details from API", "endpoint", mqtt.DeviceClient.ApiEndpoint, "mission_id", missionId, "final_url", fmt.Sprintf("%s/%s", mqtt.DeviceClient.ApiEndpoint, missionId))
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", mqtt.DeviceClient.ApiEndpoint, missionId), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("APIKEY %s", mqtt.DeviceClient.ApiKey))
	req.Header.Add("Org-Id", mqtt.DeviceClient.OrgId)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch mission details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API returned non-200 status: %d, failed to read response body: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API returned non-200 status: %d, response body: %s", resp.StatusCode, string(body))
	}

	var apiResp APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse mission details: %w", err)
	}

	details := transformAPIResponseToMissionDetails(apiResp)
	slog.Info("Loaded mission details", "details", details)

	return details, nil
}

// Keep the original file-based loading as fallback
func loadMissionDetailsFromFile(missionId string) (*MissionDetails, error) {
	detailsPath := filepath.Join("assets", fmt.Sprintf("waypoints.json"))

	data, err := os.ReadFile(detailsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read mission details: %w", err)
	}

	var details MissionDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return nil, fmt.Errorf("failed to parse mission details: %w", err)
	}

	return &details, nil
}

func (e *MissionExecutor) waitForWaypointsCompletion(statusChan chan FeedbackType) error {
	timeout := time.After(300 * time.Second)

	for {
		select {
		case status := <-statusChan:
			slog.Info("Received waypoint follower status", "status", status)

			switch status {
			case FeedbackType_WAYPOINT_FAILED:
				return fmt.Errorf("waypoint execution failed")
			case FeedbackType_WAYPOINT_REACHED:
				if e.PendingWaypoints == 0 {
					return nil
				}
			}
		case <-timeout:
			return fmt.Errorf("waypoint following timed out")
		}
	}
}

func (e *MissionExecutor) ProcessWaypointFeedback(feedbackJson string) {
	var feedback Feedback
	if err := json.Unmarshal([]byte(feedbackJson), &feedback); err != nil {
		slog.Error("Failed to unmarshal waypoint", "error", err)
		return
	}

	slog.Info("Processed waypoint feedback", "feedback", feedback)
	e.PendingWaypoints--

	if feedback.Type == FeedbackType_WAYPOINT_REACHED {
		if e.PendingWaypoints == 0 {
			shadow.GlobalShadow.MissionSchedule.Status = PM.MissionStatus_MISSION_STATUS_COMPLETED
			shadow.GlobalShadow.UpdateShadow()
		}
	} else if feedback.Type == FeedbackType_WAYPOINT_FAILED {
		shadow.GlobalShadow.MissionSchedule.Status = PM.MissionStatus_MISSION_STATUS_FAILED
		shadow.GlobalShadow.UpdateShadow()
	}
}

// Add cleanup method
func (e *MissionExecutor) Close() {
	if e.WaitSet != nil {
		e.WaitSet.Close()
	}
	if e.WaypointsPub != nil {
		e.WaypointsPub.Close()
	}
	if e.FeedbackSub != nil {
		e.FeedbackSub.Close()
	}
	if e.Node != nil {
		e.Node.Close()
	}
}
