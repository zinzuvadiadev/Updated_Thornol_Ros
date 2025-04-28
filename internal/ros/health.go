package ros

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	shadow "thornol/internal/shadow"
	"time"

	"github.com/tiiuae/rclgo/pkg/rclgo"
)

// SystemState tracks the running state of navigation components
type SystemState struct {
	Nav2Running  bool
	SlamRunning  bool
	CartoRunning bool
	LastChecked  time.Time
}

var currentState SystemState

// Critical nodes that indicate if a system is running
var (
	nav2Nodes = []string{
		"/bt_navigator",
		"/controller_server",
		"/planner_server",
		"/behavior_server",
	}

	slamNodes = []string{
		"/slam_toolbox",
		"/async_slam_toolbox_node",
	}

	cartoNodes = []string{
		"/cartographer_node",
		"/cartographer_occupancy_grid_node",
	}
)

// CheckNavigationSystem checks the state of all navigation components with timeout
func CheckNavigationSystem(ctx *rclgo.Context) error {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create command with context
	cmd := exec.CommandContext(timeoutCtx, "ros2", "node", "list")
	output, err := cmd.Output()
	if err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("timeout while listing ROS2 nodes")
		}
		return fmt.Errorf("failed to list ROS2 nodes: %v", err)
	}

	activeNodes := strings.Split(string(output), "\n")
	nodeMap := make(map[string]bool)
	for _, node := range activeNodes {
		nodeMap[strings.TrimSpace(node)] = true
	}

	// Check Nav2
	nav2Running := true
	for _, required := range nav2Nodes {
		if !nodeMap[required] {
			nav2Running = false
			break
		}
	}

	// Check SLAM
	slamRunning := false
	for _, required := range slamNodes {
		if nodeMap[required] {
			slamRunning = true
			break
		}
	}

	// Check Cartographer
	cartoRunning := false
	for _, required := range cartoNodes {
		if nodeMap[required] {
			cartoRunning = true
			break
		}
	}

	// Update current state
	currentState = SystemState{
		Nav2Running:  nav2Running,
		SlamRunning:  slamRunning,
		CartoRunning: cartoRunning,
		LastChecked:  time.Now(),
	}

	// Update shadow state
	shadow.GlobalShadow.NavstackRunning = nav2Running
	shadow.GlobalShadow.SlamRunning = slamRunning
	shadow.GlobalShadow.CartoRunning = cartoRunning
	shadow.GlobalShadow.UpdateShadow()

	slog.Info("Navigation system state updated",
		"nav2", nav2Running,
		"slam", slamRunning,
		"carto", cartoRunning)

	return nil
}

// StartHealthCheck starts a goroutine that periodically checks system health
func StartHealthCheck(ctx *rclgo.Context, checkInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		slog.Info("Starting navigation system health check", "interval", checkInterval)

		// Do an initial check immediately
		if err := CheckNavigationSystem(ctx); err != nil {
			slog.Error("Initial navigation system check failed", "error", err)
		}

		for {
			select {
			case <-ticker.C:
				if err := CheckNavigationSystem(ctx); err != nil {
					slog.Error("Failed to check navigation system", "error", err)
				}
				// case <-ctx.Done():
				// 	slog.Info("Health check stopped due to context cancellation")
				// 	return
				// }
			}
		}
	}()
}
