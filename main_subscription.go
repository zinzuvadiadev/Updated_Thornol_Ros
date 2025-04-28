package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/tiiuae/rclgo/pkg/rclgo"

	rclgoHelper "thornol/internal/rclgo"
	geometry_msgs "thornol/internal/rclgo/generated/geometry_msgs/msg"
	sensor_msgs_msg "thornol/internal/rclgo/generated/sensor_msgs/msg"
	Config "thornol/config"
)

func run(conf *Config.Config) error {
	roomCtx := context.Background()

	var pubNode *rclgo.Node
	var pub *geometry_msgs.TwistPublisher
	var err error

	twistDataChan := make(chan []byte)
	filteredCloudChan := make(chan *sensor_msgs_msg.PointCloud2)

	if err := rclgo.Init(nil); err != nil {
		panic(fmt.Errorf("failed to initialize rclgo: %v", err))
	}

	pubNode, err = rclgo.NewNode("publisher", "")
	if err != nil {
		panic(fmt.Errorf("failed to create pubNode: %v", err))
	}
	defer pubNode.Close()

	// Get topics from configuration
	cmdVelTopic := conf.TopicMaps.Publications[Config.CmdVel].Topic
	filteredCloudTopic := conf.TopicMaps.Subscriptions["filtered_cloud"].Topic
	scanTopic := conf.TopicMaps.Subscriptions[Config.LaserScan].Topic

	fmt.Println("\nAvailable topics from configuration:")
	fmt.Printf("1. %s (publisher)\n", cmdVelTopic)
	fmt.Printf("2. %s (subscription)\n", scanTopic)
	fmt.Printf("3. %s (subscription)\n", filteredCloudTopic)
	fmt.Println("\nAttempting to subscribe to topics...")

	pub, err = geometry_msgs.NewTwistPublisher(pubNode, cmdVelTopic, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create publisher: %v", err))
	}
	defer pub.Close()
	fmt.Printf("Successfully created publisher for topic: %s\n", cmdVelTopic)

	rclgoHelper.PublishTwistData(pub, twistDataChan)

	subNode, err := rclgo.NewNode("thornol_subscriber", "")
	if err != nil {
		return fmt.Errorf("failed to create node: %v", err)
	}
	defer subNode.Close()

	scanChan := make(chan *sensor_msgs_msg.LaserScan)
	subscriptionScan, err := sensor_msgs_msg.NewLaserScanSubscription(subNode, scanTopic, nil, func(msg *sensor_msgs_msg.LaserScan, info *rclgo.MessageInfo, err error) {
		if err != nil {
			slog.Error("failed to receive message: %v", err)
			return
		}
		scanChan <- msg
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to %s: %v\n", scanTopic, err)
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionScan.Close()
	fmt.Printf("Successfully subscribed to %s\n", scanTopic)

	// Add filtered cloud subscription
	subscriptionFilteredCloud, err := sensor_msgs_msg.NewPointCloud2Subscription(
		subNode,
		filteredCloudTopic,
		nil,
		func(msg *sensor_msgs_msg.PointCloud2, info *rclgo.MessageInfo, err error) {
			if err != nil {
				slog.Error("failed to receive filtered cloud message: %v", err)
				return
			}
			slog.Debug("Received filtered cloud", "data", msg)
			filteredCloudChan <- msg
		},
	)
	if err != nil {
		fmt.Printf("Failed to subscribe to %s: %v\n", filteredCloudTopic, err)
		return fmt.Errorf("failed to create filtered cloud subscriber: %v", err)
	}
	defer subscriptionFilteredCloud.Close()
	fmt.Printf("Successfully subscribed to %s\n", filteredCloudTopic)

	slog.Info("All subscriptions established")

	go func() {
		var count atomic.Int64
		for {
			select {
			case scan := <-scanChan:
				_ = scan
				count.Add(1)
				slog.Debug("Received scan", "count", count.Load())
			case cloud := <-filteredCloudChan:
				slog.Debug("Processing filtered cloud", "data", cloud)
			}
		}
	}()

	ws, err := rclgo.NewWaitSet()
	if err != nil {
		return fmt.Errorf("failed to create waitset: %v", err)
	}
	defer ws.Close()
	ws.AddSubscriptions(
		subscriptionScan.Subscription,
		subscriptionFilteredCloud.Subscription,
	)
	
	if err := ws.Run(roomCtx); err != nil {
		slog.Error(err.Error())
	}
	return nil
}