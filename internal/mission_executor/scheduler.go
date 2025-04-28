package mission_executor

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"thornol/internal/mqtt"
	PM "thornol/internal/protos"
	"thornol/internal/ros"
	shadow "thornol/internal/shadow"
	utils "thornol/internal/utils"

	"github.com/go-co-op/gocron"
	"github.com/tiiuae/rclgo/pkg/rclgo"
)

type Scheduler struct {
	scheduler       *gocron.Scheduler
	activeMission   sync.Mutex
	isRunning       bool
	MissionExecutor *MissionExecutor
}

var GlobalScheduler *Scheduler

func Init(golainClient *mqtt.GolainClient) {
	GlobalScheduler = &Scheduler{
		scheduler: gocron.NewScheduler(time.Local),
		isRunning: false,
	}

	// Initialize ROS2
	if err := rclgo.Init(nil); err != nil {
		slog.Error("Failed to initialize ROS2", "error", err)
		return
	}

	ctx, err := rclgo.NewContext(rclgo.ClockTypeSystemTime, nil)
	if err != nil {
		slog.Error("Failed to create ROS2 context", "error", err)
		return
	}

	// Start health check system (check every 30 seconds)
	// This is non-blocking as it runs in a goroutine
	ros.StartHealthCheck(ctx, 10*time.Second)

	// Load map metadata if it exists
	if metadata, err := utils.LoadMapMetadata(golainClient.NavStackMapFolderPath); err != nil {
		slog.Error("Failed to load map metadata", "error", err)
	} else if metadata != nil && metadata.MapID != "" {
		shadow.GlobalShadow.MapId = metadata.MapID
		shadow.GlobalShadow.MapFound = true
		shadow.GlobalShadow.UpdateShadow()
	}
}

func (s *Scheduler) Start(options ...func(*Scheduler)) {
	slog.Debug("Starting mission scheduler")

	for _, option := range options {
		option(s)
	}

	// TODO: Check if there's a map of the current environment of the bot
	// TODO: if not, throw an error, update the shadow with the error
	// TODO: if there is, send the map to ericr container and run navstack on it.

	// mapFound := true

	// check if map.pgm exists in ./map folder
	projectRoot, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get working directory", "error", err)
		return
	}
	mapPath := filepath.Join(projectRoot, "internal", "assets", "map", "map.pgm")
	if _, err := os.Stat(mapPath); os.IsNotExist(err) {
		// mapFound = false
	}

	// if !mapFound {
	// 	slog.Error("Map file not found",
	// 		"path", mapPath,
	// 		"error", err)
	// 	shadow.GlobalShadow.MapFound = false
	// 	shadow.GlobalShadow.UpdateShadow()
	// 	return
	// }

	shadow.GlobalShadow.MapFound = true

	fmt.Println("Map found, updating shadow")
	shadow.GlobalShadow.UpdateShadow()

	if err := rclgo.Init(nil); err != nil {
		panic(fmt.Errorf("failed to initialize rclgo: %v", err))
	}

	node, err := rclgo.NewNode("mission_executor", "")
	if err != nil {
		panic(fmt.Errorf("failed to create node: %w", err))
	}

	s.MissionExecutor = NewMissionExecutor(node)

	// err = s.RunNavStack()
	// if err != nil {
	// 	slog.Error("Failed to run navstack", "error", err)
	// 	return
	// }

	// Run NavStack on this map.

	// Run every minute to check for missions
	s.scheduler.Every(15).Seconds().Do(s.checkAndExecuteMissions)
	s.scheduler.StartAsync()
}

func (s *Scheduler) Stop() {
	s.scheduler.Stop()
	s.MissionExecutor.Close()
}

func (s *Scheduler) RunNavStack() error {
	// TODO: Run NavStack on this map.
	// firstly, copy the map.pgm to the navstack container
	// secondly, run the navstack container
	// thirdly, wait for the navstack container to finish
	// For now, we just copy the ./map/map.pgm and ./map/map.yaml to the correct folder, since we do not have dockers running.

	// copy the map.pgm to the navstack container
	// cpCmd := exec.Command("cp", "./map/map.pgm", "~/opensource/ericr-bot-code/src/Testbed_ROS2/testbed_navigation/maps/map.pgm")
	// cpErr := cpCmd.Run()
	// if cpErr != nil {
	// 	slog.Error("Failed to copy map.pgm to navstack container", "error", cpErr)
	// 	return
	// }

	// // copy the map.yaml to the navstack container
	// cpCmd = exec.Command("cp", "./map/map.yaml", "/opensource/ericr-bot-code/src/Testbed_ROS2/testbed_navigation/maps/map.yaml")
	// cpErr = cpCmd.Run()
	// if cpErr != nil {
	// 	slog.Error("Failed to copy map.yaml to navstack container", "error", cpErr)
	// 	return
	// }

	projectRoot, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to get working directory", "error", err)
		return err
	}

	setupBashPath := filepath.Join(projectRoot, "bot", "install", "setup.bash")

	// Create command with captured output
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && ros2 launch testbed_navigation navigation.launch.py", setupBashPath))

	// Capture both stdout and stderr
	cmd.Stdout = os.Stdout // This will show in your application logs
	cmd.Stderr = os.Stderr // This will show in your application logs

	// Log the command being executed
	slog.Debug("Executing command",
		"command", cmd.String(),
		"setupBashPath", setupBashPath)

	// Run the command
	err = cmd.Run()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			slog.Error("Command failed",
				"error", err,
				"stderr", string(exitErr.Stderr),
				"status", exitErr.ExitCode())
		} else {
			slog.Error("Command failed with non-exit error", "error", err)
		}

		return err
	}

	return nil
}

func (s *Scheduler) checkAndExecuteMissions() {
	slog.Info("Checking and executing missions")

	if s.isRunning {
		return
	}

	slog.Info("No active mission, checking for missions to execute")

	// Get current mission schedule from shadow
	currentMission := shadow.GlobalShadow.MissionSchedule
	if currentMission == nil {
		slog.Info("No mission schedule found, skipping")
		return
	}

	// Check if mission should run now based on schedule
	shouldRun := s.shouldRunMission(currentMission)
	if !shouldRun {
		slog.Info("Mission should not run, skipping")
		return
	}

	slog.Info("Mission should run, executing mission")
	go func() {
		s.executeMission(currentMission)
	}()
}

func (s *Scheduler) shouldRunMission(mission *PM.MissionSchedule) bool {
	if mission.Status != PM.MissionStatus_MISSION_STATUS_PENDING {
		return false
	}

	switch mission.RunType {
	case PM.MissionRunType_MISSION_RUN_TYPE_IMMEDIATE:
		// Parse the creation time
		nextRunTime, err := time.Parse(time.RFC3339, mission.NextRunTime)
		if err != nil {
			slog.Error("Failed to parse creation time", "error", err)
			return false
		}

		// Check if the mission was created in the current minute
		now := time.Now()
		if now.Sub(nextRunTime) > time.Minute {
			// Mission was created more than a minute ago, mark as skipped
			slog.Debug("Mission was created more than a minute ago, marking as skipped")
			mission.Status = PM.MissionStatus_MISSION_STATUS_SKIPPED_CLIENT_OFFLINE
			shadow.GlobalShadow.UpdateShadow()
			return false
		}
		return true

	case PM.MissionRunType_MISSION_RUN_TYPE_RECURRING:
		if mission.Recurring == nil || mission.Recurring.GetWeekly() == nil {
			return false
		}

		schedule := mission.Recurring.GetWeekly()
		now := time.Now()
		currentDay := int(now.Weekday())

		// Check if current day is in schedule
		dayScheduled := false
		for _, day := range schedule.Days {
			if int(day) == currentDay {
				dayScheduled = true
				break
			}
		}

		if !dayScheduled {
			return false
		}

		// Check if current time matches scheduled time
		scheduleTime := schedule.Time
		if scheduleTime == nil {
			return false
		}

		currentHour := now.Hour()
		currentMinute := now.Minute()

		return currentHour == int(scheduleTime.Hours) && currentMinute == int(scheduleTime.Minutes)

	default:
		return false
	}
}

func (s *Scheduler) executeMission(mission *PM.MissionSchedule) {
	s.isRunning = true

	// Update mission status to in progress
	mission.Status = PM.MissionStatus_MISSION_STATUS_IN_PROGRESS
	mission.LastExecutedAt = time.Now().Format(time.RFC3339)
	shadow.GlobalShadow.UpdateShadow()

	// Execute the mission details
	err := s.MissionExecutor.ExecuteMissionDetails(mission.MissionId)
	if err != nil {
		slog.Error("Failed to execute mission", "error", err, "mission_id", mission.MissionId)
		mission.Status = PM.MissionStatus_MISSION_STATUS_FAILED
		shadow.GlobalShadow.UpdateShadow()
		s.isRunning = false
		return
	}


	// Calculate next run time for recurring missions
	if mission.RunType == PM.MissionRunType_MISSION_RUN_TYPE_RECURRING && mission.Recurring != nil {
		if weekly := mission.Recurring.GetWeekly(); weekly != nil {
			nextRun := s.calculateNextRunTime(weekly)
			mission.NextRunTime = nextRun.Format(time.RFC3339)
		}
	}

	shadow.GlobalShadow.UpdateShadow()
	s.isRunning = false
}

func (s *Scheduler) calculateNextRunTime(schedule *PM.WeeklySchedule) time.Time {
	now := time.Now()
	currentWeekday := int(now.Weekday())

	// Find the next scheduled day
	var nextDay int
	for _, day := range schedule.Days {
		if int(day) > currentWeekday {
			nextDay = int(day)
			break
		}
	}

	// If no next day found in current week, get first day from next week
	if nextDay == 0 && len(schedule.Days) > 0 {
		nextDay = int(schedule.Days[0])
		// Add days until next week's scheduled day
		daysUntilNext := (7 - currentWeekday) + nextDay
		now = now.AddDate(0, 0, daysUntilNext)
	} else if nextDay > 0 {
		// Add days until next scheduled day this week
		now = now.AddDate(0, 0, nextDay-currentWeekday)
	}

	// Set the time component
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		int(schedule.Time.Hours), int(schedule.Time.Minutes), 0, 0,
		time.Local,
	)
}
