package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"thornol/internal/mqtt"
	pb "thornol/internal/protos"
	shadow "thornol/internal/shadow"

	Config "thornol/config"

	bridge "github.com/golain-io/mqtt-bridge"

	// "thornol/internal/bridge"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"thornol/internal/utils"
)

type ThornolDefaultServiceServer struct {
	pb.UnimplementedThornolDefaultServiceServer
	navStackProcess     *exec.Cmd
	slamToolboxProcess  *exec.Cmd
	cartographerProcess *exec.Cmd
}

const (
	NAVSTACK_RUN_CMD_ENV      = "ROS2_NAVSTACK_RUN_CMD"
	NAVSTACK_STOP_CMD_ENV     = "ROS2_NAVSTACK_STOP_CMD"
	SLAM_RUN_CMD_ENV          = "ROS2_SLAM_RUN_CMD"
	SLAM_STOP_CMD_ENV         = "ROS2_SLAM_STOP_CMD"
	CARTOGRAPHER_RUN_CMD_ENV  = "ROS2_CARTOGRAPHER_RUN_CMD"
	CARTOGRAPHER_STOP_CMD_ENV = "ROS2_CARTOGRAPHER_STOP_CMD"
)

func (s *ThornolDefaultServiceServer) GetProcessStatus(ctx context.Context, req *pb.GetProcessStatusRequest) (*pb.GetProcessStatusResponse, error) {
	return &pb.GetProcessStatusResponse{
		NavStackRunning:     s.navStackProcess != nil,
		SlamToolboxRunning:  s.slamToolboxProcess != nil,
		CartographerRunning: s.cartographerProcess != nil,
	}, nil
}

func (s *ThornolDefaultServiceServer) RunNavStack(ctx context.Context, req *pb.RunNavStackRequest) (*pb.RunNavStackResponse, error) {
	// Dynamic project root path: Use ROS_ROOT from env, fallback to os.Getwd() if not set.
	projectRoot := os.Getenv("ROS_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			slog.Error("ROS_ROOT not set and failed to get working directory", "error", err)
			return &pb.RunNavStackResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
	}

	setupBashPath := filepath.Join(projectRoot, "install", "setup.bash")

	// Use ROS2 command from env variable or fallback
	navStackCmd := os.Getenv(NAVSTACK_RUN_CMD_ENV)
	if navStackCmd == "" {
		navStackCmd = "ros2 launch testbed_navigation navigation.launch.py"
	}

	finalCmd := fmt.Sprintf("source %s && %s", setupBashPath, navStackCmd)
	// Create command with captured output using the command from env
	cmd := exec.Command("bash", "-c", finalCmd)

	// Setup logging
	cmdLogger, err := utils.NewCmdLogger("navstack")
	if err != nil {
		slog.Error("Failed to create command logger", "error", err)
		return &pb.RunNavStackResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to setup logging: %v", err),
		}, nil
	}

	cmd.Stdout, cmd.Stderr = cmdLogger.GetWriters()
	s.navStackProcess = cmd

	// Log the command being executed
	slog.Info("Executing command",
		"command", cmd.String())
	// Run the command in a goroutine
	go func() {
		defer cmdLogger.Close()
		err := cmd.Run()
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
			s.navStackProcess = nil
			shadow.GlobalShadow.NavstackRunning = false
			shadow.GlobalShadow.UpdateShadow()
		}
	}()

	shadow.GlobalShadow.NavstackRunning = true
	shadow.GlobalShadow.UpdateShadow()

	return &pb.RunNavStackResponse{
		Success: true,
	}, nil
}

func (s *ThornolDefaultServiceServer) StopNavStack(ctx context.Context, req *pb.StopNavStackRequest) (*pb.StopNavStackResponse, error) {
	// Dynamic project root path: Use ROS_ROOT from env, fallback to os.Getwd() if not set.
	projectRoot := os.Getenv("ROS_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			slog.Error("ROS_ROOT not set and failed to get working directory", "error", err)
			return &pb.StopNavStackResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
	}

	setupBashPath := filepath.Join(projectRoot, "install", "setup.bash")

	// Use ROS2 command from env variable or fallback
	navStackCmd := os.Getenv(NAVSTACK_STOP_CMD_ENV)
	if navStackCmd == "" {
		navStackCmd = "ros2 launch testbed_navigation navigation.launch.py"
	}

	finalCmd := fmt.Sprintf("source %s && %s", setupBashPath, navStackCmd)
	// Create command with captured output using the command from env
	cmd := exec.Command("bash", "-c", finalCmd)

	// Setup logging
	cmdLogger, err := utils.NewCmdLogger("navstack")
	if err != nil {
		slog.Error("Failed to create command logger", "error", err)
		return &pb.StopNavStackResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to setup logging: %v", err),
		}, nil
	}

	cmd.Stdout, cmd.Stderr = cmdLogger.GetWriters()
	s.navStackProcess = cmd

	// Log the command being executed
	slog.Info("Executing command",
		"command", cmd.String())
	// Run the command in a goroutine
	go func() {
		defer cmdLogger.Close()
		err := cmd.Run()
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
			s.navStackProcess = nil
			shadow.GlobalShadow.NavstackRunning = false
			shadow.GlobalShadow.UpdateShadow()
		}
	}()

	shadow.GlobalShadow.NavstackRunning = false
	shadow.GlobalShadow.UpdateShadow()

	return &pb.StopNavStackResponse{
		Success: true,
	}, nil
}

func (s *ThornolDefaultServiceServer) RunSlamToolbox(ctx context.Context, req *pb.RunSlamToolboxRequest) (*pb.RunSlamToolboxResponse, error) {
	// Dynamic project root path: Use ROS_ROOT from env, fallback to os.Getwd() if not set.
	projectRoot := os.Getenv("ROS_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			slog.Error("ROS_ROOT not set and failed to get working directory", "error", err)
			return &pb.RunSlamToolboxResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
	}

	setupBashPath := filepath.Join(projectRoot, "install", "setup.bash")

	// Use ROS2 command from env variable or fallback
	slamCmd := os.Getenv(SLAM_RUN_CMD_ENV)
	if slamCmd == "" {
		slamCmd = "ros2 launch slam_toolbox online_async_launch.py"
	}
	// Create command with captured output using the command from env
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && %s", setupBashPath, slamCmd))

	// Setup logging
	cmdLogger, err := utils.NewCmdLogger("slamtoolbox")
	if err != nil {
		slog.Error("Failed to create command logger", "error", err)
		return &pb.RunSlamToolboxResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to setup logging: %v", err),
		}, nil
	}

	// Set command output writers
	cmd.Stdout, cmd.Stderr = cmdLogger.GetWriters()

	s.slamToolboxProcess = cmd

	// Log the command being executed
	slog.Info("Executing command",
		"command", cmd.String())

	// Run the command in a goroutine
	go func() {
		defer cmdLogger.Close()

		err := cmd.Run()
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

			s.slamToolboxProcess = nil
			shadow.GlobalShadow.SlamRunning = false
			shadow.GlobalShadow.UpdateShadow()
		}
	}()

	shadow.GlobalShadow.SlamRunning = true
	shadow.GlobalShadow.UpdateShadow()

	return &pb.RunSlamToolboxResponse{
		Success: true,
	}, nil
}

func (s *ThornolDefaultServiceServer) StopSlamToolbox(ctx context.Context, req *pb.StopSlamToolboxRequest) (*pb.StopSlamToolboxResponse, error) {
	if s.slamToolboxProcess != nil {
		slog.Debug("Stopping SlamToolbox")
		if err := s.slamToolboxProcess.Process.Signal(os.Interrupt); err != nil {
			slog.Error("Failed to send interrupt signal to SlamToolbox", "error", err)
		}

		// Wait for process to end in a goroutine
		go func() {
			if err := s.slamToolboxProcess.Wait(); err != nil {
				slog.Error("Error waiting for SlamToolbox to stop", "error", err)
			}
			s.slamToolboxProcess = nil
			shadow.GlobalShadow.SlamRunning = false
			shadow.GlobalShadow.UpdateShadow()
		}()
	}

	return &pb.StopSlamToolboxResponse{
		Success: true,
	}, nil
}

func (s *ThornolDefaultServiceServer) RunCartographer(ctx context.Context, req *pb.RunCartographerRequest) (*pb.RunCartographerResponse, error) {
	if s.cartographerProcess != nil {
		slog.Debug("Cartographer already running")
		return &pb.RunCartographerResponse{
			Success: false,
			Error:   "Cartographer already running",
		}, nil
	}

	// Dynamic project root path: Use ROS_ROOT from env, fallback to os.Getwd() if not set.
	projectRoot := os.Getenv("ROS_ROOT")
	if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			slog.Error("ROS_ROOT not set and failed to get working directory", "error", err)
			return &pb.RunCartographerResponse{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
	}

	setupBashPath := filepath.Join(projectRoot, "install", "setup.bash")

	// Use ROS2 command from env variable or fallback
	cartoCmd := os.Getenv(CARTOGRAPHER_RUN_CMD_ENV)
	if cartoCmd == "" {
		cartoCmd = "ros2 launch cartographer_ros online_launch_cartographer.launch.py"
	}
	// Create command with captured output using the command from env
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && %s", setupBashPath, cartoCmd))

	// Setup logging
	cmdLogger, err := utils.NewCmdLogger("cartographer")
	if err != nil {
		slog.Error("Failed to create command logger", "error", err)
		return &pb.RunCartographerResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to setup logging: %v", err),
		}, nil
	}

	cmd.Stdout, cmd.Stderr = cmdLogger.GetWriters()
	s.cartographerProcess = cmd

	// Log the command being executed
	slog.Info("Executing command",
		"command", cmd.String(),
		"setupBashPath", setupBashPath)

	// Run the command
	go func() {
		defer cmdLogger.Close()
		err := cmd.Run()
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

			s.cartographerProcess = nil
			shadow.GlobalShadow.CartoRunning = false
			shadow.GlobalShadow.UpdateShadow()

			return
		}

	}()

	shadow.GlobalShadow.CartoRunning = true
	shadow.GlobalShadow.UpdateShadow()

	return &pb.RunCartographerResponse{
		Success: true,
	}, nil
}

func (s *ThornolDefaultServiceServer) StopCartographer(ctx context.Context, req *pb.StopCartographerRequest) (*pb.StopCartographerResponse, error) {
	if s.cartographerProcess != nil {
		slog.Debug("Stopping Cartographer")
		// Send SIGINT instead of Kill
		if err := s.cartographerProcess.Process.Signal(os.Interrupt); err != nil {
			slog.Error("Failed to send interrupt signal to Cartographer", "error", err)
		}

		// Wait for process to end in a goroutine
		go func() {
			if err := s.cartographerProcess.Wait(); err != nil {
				slog.Error("Error waiting for Cartographer to stop", "error", err)
			}
			s.cartographerProcess = nil
			shadow.GlobalShadow.CartoRunning = false
			shadow.GlobalShadow.UpdateShadow()
		}()
	}

	return &pb.StopCartographerResponse{
		Success: true,
	}, nil
}

func StartRPCService(client mqtt.GolainClient) *bridge.MQTTNetBridge {

	// listener, err := net.Listen("tcp", ":1884")
	// if err != nil {
	// 	log.Fatalf("failed to listen: %v", err)
	// }

	logger, _ := zap.NewProduction()
	rootTopic := Config.MQTTConf.BaseTopic

	bridgeClient := bridge.NewMQTTNetBridge(client.Client, "thornol", bridge.WithRootTopic(rootTopic), bridge.WithLogger(logger), bridge.WithQoS(2))

	// // Setup bridge connection
	// err := bridgeClient.AddHook(NewTraceHook(nil, []byte(":TID:")), "trace_hook")
	// if err != nil {
	// 	slog.Error("Error adding hook", "error", err)
	// }

	grpcServer := grpc.NewServer()
	testService := &ThornolDefaultServiceServer{}
	pb.RegisterThornolDefaultServiceServer(grpcServer, testService)

	reflection.Register(grpcServer)

	go func() {
		defer grpcServer.GracefulStop()
		defer bridgeClient.Close()

		if err := grpcServer.Serve(bridgeClient); err != nil {
			slog.Error("failed to serve: %v", err)
		}

	}()

	return bridgeClient

}
