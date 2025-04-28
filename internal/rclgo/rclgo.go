package rclgo

// import rclgo "thornol/internal/rclgo/generated"
import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log/slog"
	geometry_msgs "thornol/internal/rclgo/generated/geometry_msgs/msg"
	std_msgs "thornol/internal/rclgo/generated/std_msgs/msg"
	"time"

	"github.com/google/uuid"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/tiiuae/rclgo/pkg/rclgo"
)

type CompressionType string

const (
	CompressionNone CompressionType = "none"
	CompressionGzip CompressionType = "gzip"
)

type JoystickData struct {
	Type      string  `json:"type"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Direction string  `json:"direction"`
	Distance  float64 `json:"distance"`
	Timestamp int64   `json:"timestamp"`
}

type PointAndOrientationData struct {
	Type        string  `json:"type"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Orientation float64 `json:"orientation"`
}

const ChunkSize = 15 * 1024 // 15 KiB

type ChunkedMessage struct {
	MessageID   uuid.UUID       `json:"messageId"`
	ChunkIndex  int             `json:"chunkIndex"`
	TotalChunks int             `json:"totalChunks"`
	Data        []uint8         `json:"data"`
	Compression CompressionType `json:"compression"`
}

var subscriptions []*rclgo.Subscription

func sendLargeData(sendFunc func([]byte) error, data []byte, compressed bool) error {
	messageID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	totalChunks := (len(data) + ChunkSize - 1) / ChunkSize

	for i := 0; i < totalChunks; i++ {
		start := i * ChunkSize
		end := (i + 1) * ChunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := ChunkedMessage{
			MessageID:   messageID,
			ChunkIndex:  i,
			TotalChunks: totalChunks,
			Data:        data[start:end],
			Compression: CompressionNone,
		}

		if compressed {
			chunk.Compression = CompressionGzip
		}

		slog.Debug("sending chunk", "chunk", chunk)

		jsonChunk, err := json.Marshal(chunk)
		if err != nil {
			return err
		}

		err = sendFunc(jsonChunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func PublishToDataChannel[T any](
	dataChan <-chan T,
	room *lksdk.Room,
	topic string,
	compressed bool,
) {
	for msg := range dataChan {
		b, err := json.Marshal(msg)
		if err != nil {
			slog.Error(err.Error())
		}

		if compressed {
			var buff bytes.Buffer
			// compress data using gzip
			bWriter := gzip.NewWriter(
				&buff,
			)

			_, err = bWriter.Write(b)
			if err != nil {
				slog.Error("failed to write compressed data", "error", err)
			}

			err = bWriter.Close()
			if err != nil {
				slog.Error("failed to close compressed data writer", "error", err)
			}

			b = buff.Bytes()
		}

		if room == nil {
			slog.Error("room is nil")
		}
		if room.ConnectionState() != lksdk.ConnectionStateConnected {
			slog.Error("room is not connected")
		}
		if len(b) > ChunkSize || compressed {
			err = sendLargeData(
				func(data []byte) error {
					return room.LocalParticipant.PublishDataPacket(&lksdk.UserDataPacket{
						Payload: data,
						Topic:   topic,
					}, lksdk.WithDataPublishReliable(true))
				}, b, compressed)
			if err != nil {
				slog.Error("failed to send large data", "error", err)
			}
			// slog.Debug("published data", "topic", topic)

			continue
		}
		// slog.Debug("published data", "topic", topic)
		err = room.LocalParticipant.PublishDataPacket(&lksdk.UserDataPacket{
			Payload: b,
			Topic:   topic,
		})
		if err != nil {
			slog.Error("failed to send data", "error", err)
		}
	}
}

func PublishTwistData(
	pub *geometry_msgs.TwistPublisher,
	dataChan <-chan []byte,
) error {
	if pub == nil {
		return fmt.Errorf("publisher is nil")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in PublishTwistData", "error", r)
			}
		}()

		for data := range dataChan {
			// unmarshal data
			var joystickData JoystickData
			if err := json.Unmarshal(data, &joystickData); err != nil {
				slog.Error("failed to unmarshal joystick data", "error", err)
				continue
			}

			now := time.Now().UnixMilli()
			latency := now - joystickData.Timestamp
			if latency > 500 { // Exit if latency is greater than 500ms
				slog.Warn("latency too high", "latency_ms", latency)
				continue
			}

			// create twist message
			msg := geometry_msgs.NewTwist()
			msg.Linear.X = joystickData.X
			msg.Angular.Z = joystickData.Y

			slog.Debug("publishing twist message", "linear_x", msg.Linear.X, "angular_z", msg.Angular.Z)

			if err := pub.Publish(msg); err != nil {
				slog.Error("failed to publish twist message", "error", err)
				continue
			}
		}
	}()

	return nil
}

func PublishPoseWithCovarianceStampedData(
	pub *geometry_msgs.PoseWithCovarianceStampedPublisher,
	dataChan <-chan []byte,
) error {
	go func() {
		for data := range dataChan {
			_ = data
			fmt.Printf("Recieved: %v", data)

			// unmarshal data
			var pointData geometry_msgs.Pose
			if err := json.Unmarshal(data, &pointData); err != nil {
				panic(err)
			}

			var msg *geometry_msgs.PoseWithCovarianceStamped
			msg = geometry_msgs.NewPoseWithCovarianceStamped()
			msg.Pose.Pose = pointData

			t := time.Now().UnixNano()
			msg.Header.FrameId = "map" // Set the frame ID
			msg.Header.Stamp.Sec = int32(t / 1e9)
			msg.Header.Stamp.Nanosec = uint32(t % 1e9)

			fmt.Printf("Publishing: %v", msg)
			if err := pub.Publish(msg); err != nil {
				panic(err)
			}
		}
	}()

	return nil
}

func PublishPoseStampedData(
	pub *geometry_msgs.PoseStampedPublisher,
	dataChan <-chan []byte,
) error {
	if pub == nil {
		return fmt.Errorf("publisher is nil")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in PublishPoseStampedData", "error", r)
			}
		}()

		for data := range dataChan {
			slog.Debug("received data", "data", string(data))

			// unmarshal data
			var pointData geometry_msgs.Pose
			if err := json.Unmarshal(data, &pointData); err != nil {
				slog.Error("failed to unmarshal pose data", "error", err)
				continue
			}

			msg := geometry_msgs.NewPoseStamped()
			msg.Header.FrameId = "map"
			now := time.Now()
			t := now.UnixNano()
			msg.Header.Stamp.Sec = int32(t / 1e9)
			msg.Header.Stamp.Nanosec = uint32(t % 1e9)
			msg.Pose = pointData

			slog.Debug("publishing pose stamped message", "pose", msg)

			if err := pub.Publish(msg); err != nil {
				slog.Error("failed to publish pose stamped message", "error", err)
				continue
			}
		}
	}()

	return nil
}

func PublishStringData(
	pub *std_msgs.StringPublisher,
	dataChan <-chan []byte,
) error {
	go func() {
		for data := range dataChan {
			pub.Publish(&std_msgs.String{Data: string(data)})
		}
	}()
	return nil
}
