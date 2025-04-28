package shadow_rclgo

import (
	"log/slog"
	std_msgs "thornol/internal/rclgo/generated/std_msgs/msg"
	shadow "thornol/internal/shadow"
)

func PublishSOHFeedback(dataChan <-chan *std_msgs.Float32) error {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("recovered from panic in PublishTwistData", "error", r)
			}
		}()

		for data := range dataChan {
			// create twist message
			msg := std_msgs.NewFloat32()
			msg.Data = data.Data

			shadow.GlobalShadow.Battery = data.Data
			shadow.GlobalShadow.UpdateShadow()
		}
	}()

	return nil
}
