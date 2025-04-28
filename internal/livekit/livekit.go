package livekit

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	geometry_msgs "thornol/internal/rclgo/generated/geometry_msgs/msg"
	nav_msgs "thornol/internal/rclgo/generated/nav_msgs/msg"
	sensor_msgs_msg "thornol/internal/rclgo/generated/sensor_msgs/msg"
	std_msgs "thornol/internal/rclgo/generated/std_msgs/msg"

	rclgoHelper "thornol/internal/rclgo"

	Conf "thornol/config"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v3"
	"github.com/tiiuae/rclgo/pkg/rclgo"

	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/driver/microphone"
)

type RTCJoinRoomRequest struct {
	ParticipantId string              `json:"participant_id" validate:"omitempty"`
	Token         string              `json:"token" validate:"omitempty"`
	URL           string              `json:"url" validate:"omitempty"`
	RoomId        string              `json:"room_id" validate:"omitempty"`
	Metadata      RTCConnectionParams `json:"metadata" validate:"omitempty"`
}

type RTCConnectionParams struct {
	// AudioBitRate     int                  `json:"audio_bit_rate" validate:"omitempty"`
	// VideoBitRate     int                  `json:"video_bit_rate" validate:"omitempty"`
	// VideoWidth       int                  `json:"video_width" validate:"omitempty"`
	// VideoHeight      int                  `json:"video_height" validate:"omitempty"`
	// FrameRate        int                  `json:"frame_rate" validate:"omitempty"`
	// AudioNumChannels int                  `json:"audio_num_channels" validate:"omitempty"`
	// TrackIds         []mediadevices.Track `json:"track_ids"`
	EnableAudio bool `json:"enable_audio" validate:"required"`
	EnableVideo bool `json:"enable_video" validate:"required"`

	AudioTrack string `json:"audio_track"`
	VideoTrack string `json:"video_track"`
}
type RTCDeviceCapabilitiesRequest struct {
	DeviceId  string `json:"device_id"`
	ProjectId string `json:"project_id"`
	FleetId   string `json:"fleet_id"`
}
type Capabilities struct {
	Audio        bool                           `json:"audio"`
	Video        bool                           `json:"video"`
	AudioDevices []mediadevices.MediaDeviceInfo `json:"audio_devices"`
	VideoDevices []mediadevices.MediaDeviceInfo `json:"video_devices"`
}

type RTCDeviceCapabilitiesResponse struct {
	DeviceId           string       `json:"device_id"`
	DeviceCapabilities Capabilities `json:"device_capabilities"`
}

type RTCDataChannelTopic string

const (
	RTCJoystickTopic    RTCDataChannelTopic = "joystick"
	RTCInitialPoseTopic RTCDataChannelTopic = "initialpose"
	RTCQuickGoalTopic   RTCDataChannelTopic = "quick_goal"
	RTCGoalPoseTopic    RTCDataChannelTopic = "goal_pose"
)

func RTCDeviceCapabilitiesCheck(body *RTCDeviceCapabilitiesRequest) (*RTCDeviceCapabilitiesResponse, error) {
	capabilities := RTCDeviceCapabilitiesResponse{
		DeviceId: body.DeviceId,
	}

	deviceinfo := mediadevices.EnumerateDevices()
	audiocap := []mediadevices.MediaDeviceInfo{}
	videocap := []mediadevices.MediaDeviceInfo{}
	for _, device := range deviceinfo {
		if device.Kind == mediadevices.AudioInput {
			audiocap = append(audiocap, device)
		}
		if device.Kind == mediadevices.VideoInput {
			videocap = append(videocap, device)
		}

	}

	capabilities.DeviceCapabilities.AudioDevices = audiocap
	capabilities.DeviceCapabilities.Audio = len(audiocap) > 0
	capabilities.DeviceCapabilities.VideoDevices = videocap
	capabilities.DeviceCapabilities.Video = len(videocap) > 0

	// mediaStream, err := mediadevices.GetUserMedia(
	// 	mediadevices.MediaStreamConstraints{
	// 		Video: func(c *mediadevices.MediaTrackConstraints) {
	// 		},
	// 		Audio: func(c *mediadevices.MediaTrackConstraints) {
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	log.Println("NewMediaStream error", err)
	// 	panic(err)
	// }
	// audioTracks := mediaStream.GetAudioTracks()
	// videoTracks := mediaStream.GetVideoTracks()

	// for _, track := range audioTracks {
	// 	err := track.Close()
	// 	if err != nil {
	// 		log.Println("Error closing track", err)
	// 	}
	// }
	// for _, track := range videoTracks {
	// 	err := track.Close()
	// 	if err != nil {
	// 		log.Println("Error closing track", err)
	// 	}
	// }
	return &capabilities, nil
}

var twistDataChan = make(chan []byte)
var initPoseDataChan = make(chan []byte)
var goalPoseDataChan = make(chan []byte)
var quickGoalDataChan = make(chan []byte)

var topicChannelMap = map[RTCDataChannelTopic]chan []byte{
	RTCJoystickTopic:    twistDataChan,
	RTCInitialPoseTopic: initPoseDataChan,
	RTCQuickGoalTopic:   quickGoalDataChan,
	RTCGoalPoseTopic:    goalPoseDataChan,
}

type ROSSubscription struct {
	Topic                string
	SubscriptionFunction func(node *rclgo.Node, topic string, opts *rclgo.SubscriptionOptions, callback func(msg []byte))
	Throttle             int
	Channel              chan []byte
}

func onDataPacket(data lksdk.DataPacket, params lksdk.DataReceiveParams) {
	fmt.Println("Data packet received")
	payload := data.ToProto().GetUser().GetPayload()

	if payload == nil {
		return
	}

	topic := RTCDataChannelTopic(data.ToProto().GetUser().GetTopic())

	switch topic {
	case RTCJoystickTopic:
		twistDataChan <- payload
	case RTCInitialPoseTopic:
		initPoseDataChan <- payload
	case RTCGoalPoseTopic:
		goalPoseDataChan <- payload
	case RTCQuickGoalTopic:
		quickGoalDataChan <- payload
	default:
		slog.Error("Received unknown topic on data channel", "topic", topic)
	}
}

func JoinRoom(body *RTCJoinRoomRequest, started chan<- bool) error {
	roomCtx, roomCtxCancel := context.WithCancel(context.Background())
	defer roomCtxCancel()

	mediaStream, err := mediadevices.NewMediaStream()
	if err != nil {
		log.Println("NewMediaStream error", err)
		panic(err)
	}
	// Check if audio and video are enabled

	codecSelector := mediadevices.NewCodecSelector(
		audioCodecSelector(48000),
		videoCodecSelector(1000000),
	)

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)
	constraint := mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.DeviceID = prop.String(body.Metadata.AudioTrack)
		},
		Codec: codecSelector,
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.DeviceID = prop.String(body.Metadata.VideoTrack)
		},
	}
	mediaStream, err = mediadevices.GetUserMedia(constraint)
	if err != nil {
		log.Println("GetUserMedia error", err)
		panic(err)
	}

	var pubNode *rclgo.Node
	var pub *geometry_msgs.TwistPublisher

	if err := rclgo.Init(nil); err != nil {
		panic(fmt.Errorf("failed to initialize rclgo: %v", err))
	}

	pubNode, err = rclgo.NewNode("thornol_publisher", "")
	if err != nil {
		panic(fmt.Errorf("failed to create pubNode: %v", err))
	}
	defer pubNode.Close()

	pub, err = geometry_msgs.NewTwistPublisher(
		pubNode,
		Conf.Conf.TopicMaps.Publications[Conf.CmdVel].Topic,
		nil,
	)
	if err != nil {
		panic(fmt.Errorf("failed to create publisher: %v", err))
	}
	defer pub.Close()

	rclgoHelper.PublishTwistData(pub, twistDataChan)

	pubInitialPose, err := geometry_msgs.NewPoseWithCovarianceStampedPublisher(pubNode, Conf.Conf.TopicMaps.Publications[Conf.InitialPose].Topic, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create publisher: %v", err))
	}
	defer pubInitialPose.Close()

	rclgoHelper.PublishPoseWithCovarianceStampedData(pubInitialPose, initPoseDataChan)

	pubGoalPose, err := geometry_msgs.NewPoseStampedPublisher(pubNode, Conf.Conf.TopicMaps.Publications[Conf.GoalPose].Topic, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create publisher: %v", err))
	}
	defer pubGoalPose.Close()

	rclgoHelper.PublishPoseStampedData(pubGoalPose, goalPoseDataChan)

	pubQuickGoal, err := std_msgs.NewStringPublisher(pubNode, Conf.Conf.TopicMaps.Publications[Conf.WaypointsList].Topic, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create publisher: %v", err))
	}
	defer pubQuickGoal.Close()

	rclgoHelper.PublishStringData(pubQuickGoal, quickGoalDataChan)

	subNode, err := rclgo.NewNode("thornol_subscriber", "")
	if err != nil {
		return fmt.Errorf("failed to create node: %v", err)
	}
	defer subNode.Close()

	var odomLastSent time.Time
	odomDataChan := make(chan *nav_msgs.Odometry)
	log.Default().Println("Connected to room")

	odomConfig := Conf.Conf.TopicMaps.Subscriptions[Conf.Odometry]
	subcriptionOdom, err := nav_msgs.NewOdometrySubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.Odometry].Topic,
		nil,
		func(msg *nav_msgs.Odometry, info *rclgo.MessageInfo, err error) {
			if odomConfig.Throttle > 0 && time.Since(odomLastSent) < time.Duration(odomConfig.Throttle)*time.Millisecond {
				return
			}
			odomLastSent = time.Now()
			if err != nil {
				slog.Error("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX Odom", "data", info)
			odomDataChan <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subcriptionOdom.Close()

	amclPoseChan := make(chan *geometry_msgs.PoseWithCovarianceStamped)
	subcriptionAmclPose, err := geometry_msgs.NewPoseWithCovarianceStampedSubscription(subNode, Conf.Conf.TopicMaps.Subscriptions[Conf.AMCLPose].Topic, nil, func(msg *geometry_msgs.PoseWithCovarianceStamped, info *rclgo.MessageInfo, err error) {
		if err != nil {
			slog.Error("failed to receive message: %v", err)
			return
		}
		slog.Debug("RX AmclPose", "data", info)
		amclPoseChan <- msg
	})
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subcriptionAmclPose.Close()

	occupancyChan := make(chan *nav_msgs.OccupancyGrid)
	subcriptionOccupancy, err := nav_msgs.NewOccupancyGridSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.Map].Topic,
		nil,
		func(msg *nav_msgs.OccupancyGrid, info *rclgo.MessageInfo, err error) {
			if err != nil {
				slog.Error("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX Map", "data", info)
			occupancyChan <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subcriptionOccupancy.Close()

	scanChan := make(chan *sensor_msgs_msg.LaserScan)
	subscriptionScan, err := sensor_msgs_msg.NewLaserScanSubscription(subNode, Conf.Conf.TopicMaps.Subscriptions[Conf.LaserScan].Topic, nil, func(msg *sensor_msgs_msg.LaserScan, info *rclgo.MessageInfo, err error) {
		if err != nil {
			slog.Error("failed to receive message: %v", err)
			return
		}
		slog.Info("RX Scan", "info", info, "data", msg)

		scanChan <- msg
	})
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionScan.Close()

	pathChannel := make(chan *nav_msgs.Path)
	subscriptionPath, err := nav_msgs.NewPathSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.Path].Topic,
		nil,
		func(msg *nav_msgs.Path, info *rclgo.MessageInfo, err error) {
			if err != nil {
				fmt.Println("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX Path", "data", info)
			pathChannel <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionPath.Close()

	localPathChannel := make(chan *nav_msgs.Path)
	subscriptionLocalPath, err := nav_msgs.NewPathSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.LocalPath].Topic,
		nil,
		func(msg *nav_msgs.Path, info *rclgo.MessageInfo, err error) {
			if err != nil {
				fmt.Println("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX Path", "data", info)
			localPathChannel <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionLocalPath.Close()

	mapMetaChannel := make(chan *nav_msgs.MapMetaData)
	subMapMeta, err := nav_msgs.NewMapMetaDataSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.MapMetadata].Topic,
		nil,
		func(msg *nav_msgs.MapMetaData, info *rclgo.MessageInfo, err error) {
			if err != nil {
				fmt.Println("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX MapMeta", "data", info)
			mapMetaChannel <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subMapMeta.Close()

	feedbackChan := make(chan string)
	feedbackSub, err := std_msgs.NewStringSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.NavigationFeedback].Topic,
		nil,
		func(msg *std_msgs.String, info *rclgo.MessageInfo, err error) {
			if err != nil {
				slog.Error("Failed to receive feedback message", "error", err)
				return
			}
			feedbackChan <- msg.Data
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer feedbackSub.Close()

	globalCostmapChan := make(chan *nav_msgs.OccupancyGrid)
	globalCostmapSub, err := nav_msgs.NewOccupancyGridSubscription(
		subNode,
		Conf.Conf.TopicMaps.Subscriptions[Conf.GlobalCostmap].Topic,
		nil,
		func(msg *nav_msgs.OccupancyGrid, info *rclgo.MessageInfo, err error) {
			if err != nil {
				slog.Error("failed to receive message: %v", err)
				return
			}
			slog.Debug("RX Map", "data", info)
			globalCostmapChan <- msg
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer globalCostmapSub.Close()

	socChan := make(chan *std_msgs.Float32)
	subscriptionSoc, err := std_msgs.NewFloat32Subscription(subNode, Conf.Conf.TopicMaps.Subscriptions[Conf.Soc].Topic, nil, func(msg *std_msgs.Float32, info *rclgo.MessageInfo, err error) {
		if err != nil {
			slog.Error("failed to receive message: %v", err)
			return
		}
		slog.Debug("RX Soc", "data", info)
		socChan <- msg
	})
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionSoc.Close()

	sohChan := make(chan *std_msgs.Float32)
	subscriptionSoh, err := std_msgs.NewFloat32Subscription(subNode, Conf.Conf.TopicMaps.Subscriptions[Conf.Soh].Topic, nil, func(msg *std_msgs.Float32, info *rclgo.MessageInfo, err error) {
		if err != nil {
			slog.Error("failed to receive message: %v", err)
			return
		}
		slog.Debug("RX Soc", "data", info)
		sohChan <- msg
	})
	if err != nil {
		return fmt.Errorf("failed to create subscriber: %v", err)
	}
	defer subscriptionSoh.Close()

	room, err := lksdk.ConnectToRoomWithToken(body.URL, body.Token, &lksdk.RoomCallback{
		OnParticipantConnected: func(participant *lksdk.RemoteParticipant) {
			fmt.Println("Participant connected")
		},

		OnDisconnected: func() {
			tracks := mediaStream.GetTracks()
			for _, track := range tracks {
				if (body.Metadata.EnableVideo && track.Kind() == webrtc.RTPCodecTypeVideo) || (body.Metadata.EnableAudio && track.Kind() == webrtc.RTPCodecTypeAudio) {
					err := track.Close()
					if err != nil {
						log.Println("Error closing track", err)
					}
				}
			}
		},
		OnDisconnectedWithReason: func(reason lksdk.DisconnectionReason) {
			tracks := mediaStream.GetTracks()
			for _, track := range tracks {
				err := track.Close()
				if err != nil {
					log.Println("Error closing track", err)
				}
			}
			roomCtxCancel()
		},
		ParticipantCallback: lksdk.ParticipantCallback{
			OnDataPacket: onDataPacket,
		},
	}, lksdk.WithAutoSubscribe(false))
	if err != nil {
		return err
	}

	slog.Info("Connected to room")

	started <- true

	go rclgoHelper.PublishToDataChannel(mapMetaChannel, room, "map_metadata", false)
	go rclgoHelper.PublishToDataChannel(pathChannel, room, "path", false)
	go rclgoHelper.PublishToDataChannel(odomDataChan, room, "odom", false)
	go rclgoHelper.PublishToDataChannel(occupancyChan, room, "map", true)
	go rclgoHelper.PublishToDataChannel(scanChan, room, "laser_scan", true)
	go rclgoHelper.PublishToDataChannel(feedbackChan, room, "feedback", true)
	go rclgoHelper.PublishToDataChannel(globalCostmapChan, room, "global_costmap", true)
	go rclgoHelper.PublishToDataChannel(localPathChannel, room, "local_plan", true)
	go rclgoHelper.PublishToDataChannel(socChan, room, "soc", true)
	go rclgoHelper.PublishToDataChannel(sohChan, room, "soh", true)
	go rclgoHelper.PublishToDataChannel(amclPoseChan, room, "amcl_pose", true)
	log.Default().Println("connected and subscribed to topics")

	ws, err := rclgo.NewWaitSet()
	if err != nil {
		return fmt.Errorf("failed to create waitset: %v", err)
	}
	defer ws.Close()
	ws.AddSubscriptions(
		subMapMeta.Subscription,
		subcriptionOccupancy.Subscription,
		subcriptionOdom.Subscription,
		subscriptionPath.Subscription,
		subscriptionScan.Subscription,
		feedbackSub.Subscription,
		globalCostmapSub.Subscription,
		subscriptionLocalPath.Subscription,
		subscriptionSoc.Subscription,
		subscriptionSoh.Subscription,
		subcriptionAmclPose.Subscription,
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	for _, track := range mediaStream.GetTracks() {
		track.OnEnded(func(err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered in f", r)
				}
			}()
			log.Println("Track ended with error:", track.ID(), err)
			_ = track.Close()
			// ensure the process exits as well when the track ends
			sigChan <- syscall.SIGINT
		})
		if (body.Metadata.EnableVideo && track.Kind() == webrtc.RTPCodecTypeVideo) || (body.Metadata.EnableAudio && track.Kind() == webrtc.RTPCodecTypeAudio) {
			_, err = room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{})
			if err != nil {
				log.Println("Error publishing track", err)
				err := track.Close()
				if err != nil {
					log.Println("Error closing track", err)
				}
			}
		}
	}
	if err := ws.Run(roomCtx); err != nil {
		slog.Error(err.Error())
	}
	<-sigChan

	fmt.Printf("Disconnecting from room\n")
	if room.ConnectionState() == lksdk.ConnectionStateConnected ||
		room.ConnectionState() == lksdk.ConnectionStateReconnecting {
		room.Disconnect()
	}
	return nil
}
func audioCodecSelector(bitrate int) mediadevices.CodecSelectorOption {
	opusParams, err := opus.NewParams()
	opusParams.BitRate = bitrate
	if err != nil {
		slog.Error("failed to create opus params", "error", err)
		panic(err)
	}

	selectorOptions := mediadevices.WithAudioEncoders(&opusParams)

	return selectorOptions

}
func videoCodecSelector(bitRate int) mediadevices.CodecSelectorOption {
	vpxParams, err := vpx.NewVP8Params()

	if err != nil {
		slog.Error("failed to create vpx params", "error", err)
		panic(err)
	}

	vpxParams.BitRate = bitRate

	codecSelector := mediadevices.WithVideoEncoders(&vpxParams)

	return codecSelector

}
