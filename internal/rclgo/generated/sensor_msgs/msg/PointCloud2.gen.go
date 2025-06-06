/*
This file is part of rclgo

Copyright © 2021 Technology Innovation Institute, United Arab Emirates

Licensed under the Apache License, Version 2.0 (the "License");
	http://www.apache.org/licenses/LICENSE-2.0
*/

// Code generated by rclgo-gen. DO NOT EDIT.

package sensor_msgs_msg
import (
	"unsafe"

	"github.com/tiiuae/rclgo/pkg/rclgo"
	"github.com/tiiuae/rclgo/pkg/rclgo/types"
	"github.com/tiiuae/rclgo/pkg/rclgo/typemap"
	std_msgs_msg "thornol/internal/rclgo/generated/std_msgs/msg"

	primitives "github.com/tiiuae/rclgo/pkg/rclgo/primitives"
	
)
/*
#include <rosidl_runtime_c/message_type_support_struct.h>

#include <sensor_msgs/msg/point_cloud2.h>

*/
import "C"

func init() {
	typemap.RegisterMessage("sensor_msgs/PointCloud2", PointCloud2TypeSupport)
	typemap.RegisterMessage("sensor_msgs/msg/PointCloud2", PointCloud2TypeSupport)
}

type PointCloud2 struct {
	Header std_msgs_msg.Header `yaml:"header"`// Time of sensor data acquisition, and the coordinate frame ID (for 3d points).
	Height uint32 `yaml:"height"`// 2D structure of the point cloud. If the cloud is unordered, height is1 and width is the length of the point cloud.
	Width uint32 `yaml:"width"`
	Fields []PointField `yaml:"fields"`// Describes the channels and their layout in the binary data blob.
	IsBigendian bool `yaml:"is_bigendian"`// Is this data bigendian?
	PointStep uint32 `yaml:"point_step"`// Length of a point in bytes
	RowStep uint32 `yaml:"row_step"`// Length of a row in bytes
	Data []uint8 `yaml:"data"`// Actual point data, size is (row_step*height)
	IsDense bool `yaml:"is_dense"`// True if there are no invalid points
}

// NewPointCloud2 creates a new PointCloud2 with default values.
func NewPointCloud2() *PointCloud2 {
	self := PointCloud2{}
	self.SetDefaults()
	return &self
}

func (t *PointCloud2) Clone() *PointCloud2 {
	c := &PointCloud2{}
	c.Header = *t.Header.Clone()
	c.Height = t.Height
	c.Width = t.Width
	if t.Fields != nil {
		c.Fields = make([]PointField, len(t.Fields))
		ClonePointFieldSlice(c.Fields, t.Fields)
	}
	c.IsBigendian = t.IsBigendian
	c.PointStep = t.PointStep
	c.RowStep = t.RowStep
	if t.Data != nil {
		c.Data = make([]uint8, len(t.Data))
		copy(c.Data, t.Data)
	}
	c.IsDense = t.IsDense
	return c
}

func (t *PointCloud2) CloneMsg() types.Message {
	return t.Clone()
}

func (t *PointCloud2) SetDefaults() {
	t.Header.SetDefaults()
	t.Height = 0
	t.Width = 0
	t.Fields = nil
	t.IsBigendian = false
	t.PointStep = 0
	t.RowStep = 0
	t.Data = nil
	t.IsDense = false
}

func (t *PointCloud2) GetTypeSupport() types.MessageTypeSupport {
	return PointCloud2TypeSupport
}

// PointCloud2Publisher wraps rclgo.Publisher to provide type safe helper
// functions
type PointCloud2Publisher struct {
	*rclgo.Publisher
}

// NewPointCloud2Publisher creates and returns a new publisher for the
// PointCloud2
func NewPointCloud2Publisher(node *rclgo.Node, topic_name string, options *rclgo.PublisherOptions) (*PointCloud2Publisher, error) {
	pub, err := node.NewPublisher(topic_name, PointCloud2TypeSupport, options)
	if err != nil {
		return nil, err
	}
	return &PointCloud2Publisher{pub}, nil
}

func (p *PointCloud2Publisher) Publish(msg *PointCloud2) error {
	return p.Publisher.Publish(msg)
}

// PointCloud2Subscription wraps rclgo.Subscription to provide type safe helper
// functions
type PointCloud2Subscription struct {
	*rclgo.Subscription
}

// PointCloud2SubscriptionCallback type is used to provide a subscription
// handler function for a PointCloud2Subscription.
type PointCloud2SubscriptionCallback func(msg *PointCloud2, info *rclgo.MessageInfo, err error)

// NewPointCloud2Subscription creates and returns a new subscription for the
// PointCloud2
func NewPointCloud2Subscription(node *rclgo.Node, topic_name string, opts *rclgo.SubscriptionOptions, subscriptionCallback PointCloud2SubscriptionCallback) (*PointCloud2Subscription, error) {
	callback := func(s *rclgo.Subscription) {
		var msg PointCloud2
		info, err := s.TakeMessage(&msg)
		subscriptionCallback(&msg, info, err)
	}
	sub, err := node.NewSubscription(topic_name, PointCloud2TypeSupport, opts, callback)
	if err != nil {
		return nil, err
	}
	return &PointCloud2Subscription{sub}, nil
}

func (s *PointCloud2Subscription) TakeMessage(out *PointCloud2) (*rclgo.MessageInfo, error) {
	return s.Subscription.TakeMessage(out)
}

// ClonePointCloud2Slice clones src to dst by calling Clone for each element in
// src. Panics if len(dst) < len(src).
func ClonePointCloud2Slice(dst, src []PointCloud2) {
	for i := range src {
		dst[i] = *src[i].Clone()
	}
}

// Modifying this variable is undefined behavior.
var PointCloud2TypeSupport types.MessageTypeSupport = _PointCloud2TypeSupport{}

type _PointCloud2TypeSupport struct{}

func (t _PointCloud2TypeSupport) New() types.Message {
	return NewPointCloud2()
}

func (t _PointCloud2TypeSupport) PrepareMemory() unsafe.Pointer { //returns *C.sensor_msgs__msg__PointCloud2
	return (unsafe.Pointer)(C.sensor_msgs__msg__PointCloud2__create())
}

func (t _PointCloud2TypeSupport) ReleaseMemory(pointer_to_free unsafe.Pointer) {
	C.sensor_msgs__msg__PointCloud2__destroy((*C.sensor_msgs__msg__PointCloud2)(pointer_to_free))
}

func (t _PointCloud2TypeSupport) AsCStruct(dst unsafe.Pointer, msg types.Message) {
	m := msg.(*PointCloud2)
	mem := (*C.sensor_msgs__msg__PointCloud2)(dst)
	std_msgs_msg.HeaderTypeSupport.AsCStruct(unsafe.Pointer(&mem.header), &m.Header)
	mem.height = C.uint32_t(m.Height)
	mem.width = C.uint32_t(m.Width)
	PointField__Sequence_to_C(&mem.fields, m.Fields)
	mem.is_bigendian = C.bool(m.IsBigendian)
	mem.point_step = C.uint32_t(m.PointStep)
	mem.row_step = C.uint32_t(m.RowStep)
	primitives.Uint8__Sequence_to_C((*primitives.CUint8__Sequence)(unsafe.Pointer(&mem.data)), m.Data)
	mem.is_dense = C.bool(m.IsDense)
}

func (t _PointCloud2TypeSupport) AsGoStruct(msg types.Message, ros2_message_buffer unsafe.Pointer) {
	m := msg.(*PointCloud2)
	mem := (*C.sensor_msgs__msg__PointCloud2)(ros2_message_buffer)
	std_msgs_msg.HeaderTypeSupport.AsGoStruct(&m.Header, unsafe.Pointer(&mem.header))
	m.Height = uint32(mem.height)
	m.Width = uint32(mem.width)
	PointField__Sequence_to_Go(&m.Fields, mem.fields)
	m.IsBigendian = bool(mem.is_bigendian)
	m.PointStep = uint32(mem.point_step)
	m.RowStep = uint32(mem.row_step)
	primitives.Uint8__Sequence_to_Go(&m.Data, *(*primitives.CUint8__Sequence)(unsafe.Pointer(&mem.data)))
	m.IsDense = bool(mem.is_dense)
}

func (t _PointCloud2TypeSupport) TypeSupport() unsafe.Pointer {
	return unsafe.Pointer(C.rosidl_typesupport_c__get_message_type_support_handle__sensor_msgs__msg__PointCloud2())
}

type CPointCloud2 = C.sensor_msgs__msg__PointCloud2
type CPointCloud2__Sequence = C.sensor_msgs__msg__PointCloud2__Sequence

func PointCloud2__Sequence_to_Go(goSlice *[]PointCloud2, cSlice CPointCloud2__Sequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]PointCloud2, cSlice.size)
	src := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range src {
		PointCloud2TypeSupport.AsGoStruct(&(*goSlice)[i], unsafe.Pointer(&src[i]))
	}
}
func PointCloud2__Sequence_to_C(cSlice *CPointCloud2__Sequence, goSlice []PointCloud2) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.sensor_msgs__msg__PointCloud2)(C.malloc(C.sizeof_struct_sensor_msgs__msg__PointCloud2 * C.size_t(len(goSlice))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range goSlice {
		PointCloud2TypeSupport.AsCStruct(unsafe.Pointer(&dst[i]), &goSlice[i])
	}
}
func PointCloud2__Array_to_Go(goSlice []PointCloud2, cSlice []CPointCloud2) {
	for i := 0; i < len(cSlice); i++ {
		PointCloud2TypeSupport.AsGoStruct(&goSlice[i], unsafe.Pointer(&cSlice[i]))
	}
}
func PointCloud2__Array_to_C(cSlice []CPointCloud2, goSlice []PointCloud2) {
	for i := 0; i < len(goSlice); i++ {
		PointCloud2TypeSupport.AsCStruct(unsafe.Pointer(&cSlice[i]), &goSlice[i])
	}
}
