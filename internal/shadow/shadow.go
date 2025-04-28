package pkg

import (
	"log/slog"

	mqtt "thornol/internal/mqtt"
	PM "thornol/internal/protos"

	C "thornol/config"

	"google.golang.org/protobuf/proto"
)

type Shadow struct {
	*PM.Shadow
	// update shadow callback will recieve (previous shadow, current shadow)
	UpdateShadowCallback func(*PM.Shadow, *PM.Shadow)
	// first shadow update callback
	firstShadowUpdateCallback func()
	firstShadowUpdate         bool
}

var GlobalShadow Shadow
var previousShadow PM.Shadow
var tempShadowForMerge *PM.Shadow

func init() {
	GlobalShadow = Shadow{}
	GlobalShadow.Shadow = &PM.Shadow{}
	tempShadowForMerge = &PM.Shadow{}
}

// updates the shadow to the platform
func (s *Shadow) UpdateShadow() {
	data, err := proto.Marshal(s.Shadow)
	if err != nil {
		slog.Error("Error marshalling shadow", "error", err)
	}
	if err := mqtt.DeviceClient.Publish(C.MQTTConf.ShadowUpdateTopic, data); err != nil {
		slog.Error("Error publishing shadow update", "error", err)
	}
}

func (s *Shadow) SetUpdateShadowCallback(callback func(*PM.Shadow, *PM.Shadow)) {
	s.UpdateShadowCallback = callback
}

func (s *Shadow) SetFirstShadowUpdateCallback(f func()) {
	s.firstShadowUpdateCallback = f
}

func (s *Shadow) ShadowCallback(buffer []byte) error {
	// marshal the current shadow
	newShadow := PM.Shadow{}

	// unmarshal the incoming buffer into global shadow
	if err := proto.Unmarshal(buffer, &newShadow); err != nil {
		slog.Error("Error unmarshalling in global shadow", "error", err)
		return err
	} else {
		slog.Info("Shadow unmarshalled", "shadow", newShadow)
	}

	s.Shadow = &newShadow

	if s.UpdateShadowCallback != nil {
		s.UpdateShadowCallback(s.Shadow, &previousShadow)
	}

	if s.firstShadowUpdate {
		return nil
	}

	s.firstShadowUpdate = true
	if s.firstShadowUpdateCallback != nil {
		s.firstShadowUpdateCallback()
	}
	return nil
}
