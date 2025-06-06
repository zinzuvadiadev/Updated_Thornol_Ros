// Code generated by rclgo-gen. DO NOT EDIT.

package geometry_msgs_msg
import (
	"unsafe"

	"github.com/tiiuae/rclgo/pkg/rclgo"
	"github.com/tiiuae/rclgo/pkg/rclgo/types"
	"github.com/tiiuae/rclgo/pkg/rclgo/typemap"
	std_msgs_msg "thornol/internal/rclgo/generated/std_msgs/msg"
	
)
/*
#include <rosidl_runtime_c/message_type_support_struct.h>

#include <geometry_msgs/msg/pose_stamped.h>

*/
import "C"

func init() {
	typemap.RegisterMessage("geometry_msgs/PoseStamped", PoseStampedTypeSupport)
	typemap.RegisterMessage("geometry_msgs/msg/PoseStamped", PoseStampedTypeSupport)
}

type PoseStamped struct {
	Header std_msgs_msg.Header `yaml:"header"`
	Pose Pose `yaml:"pose"`
}

// NewPoseStamped creates a new PoseStamped with default values.
func NewPoseStamped() *PoseStamped {
	self := PoseStamped{}
	self.SetDefaults()
	return &self
}

func (t *PoseStamped) Clone() *PoseStamped {
	c := &PoseStamped{}
	c.Header = *t.Header.Clone()
	c.Pose = *t.Pose.Clone()
	return c
}

func (t *PoseStamped) CloneMsg() types.Message {
	return t.Clone()
}

func (t *PoseStamped) SetDefaults() {
	t.Header.SetDefaults()
	t.Pose.SetDefaults()
}

func (t *PoseStamped) GetTypeSupport() types.MessageTypeSupport {
	return PoseStampedTypeSupport
}

// PoseStampedPublisher wraps rclgo.Publisher to provide type safe helper
// functions
type PoseStampedPublisher struct {
	*rclgo.Publisher
}

// NewPoseStampedPublisher creates and returns a new publisher for the
// PoseStamped
func NewPoseStampedPublisher(node *rclgo.Node, topic_name string, options *rclgo.PublisherOptions) (*PoseStampedPublisher, error) {
	pub, err := node.NewPublisher(topic_name, PoseStampedTypeSupport, options)
	if err != nil {
		return nil, err
	}
	return &PoseStampedPublisher{pub}, nil
}

func (p *PoseStampedPublisher) Publish(msg *PoseStamped) error {
	return p.Publisher.Publish(msg)
}

// PoseStampedSubscription wraps rclgo.Subscription to provide type safe helper
// functions
type PoseStampedSubscription struct {
	*rclgo.Subscription
}

// PoseStampedSubscriptionCallback type is used to provide a subscription
// handler function for a PoseStampedSubscription.
type PoseStampedSubscriptionCallback func(msg *PoseStamped, info *rclgo.MessageInfo, err error)

// NewPoseStampedSubscription creates and returns a new subscription for the
// PoseStamped
func NewPoseStampedSubscription(node *rclgo.Node, topic_name string, opts *rclgo.SubscriptionOptions, subscriptionCallback PoseStampedSubscriptionCallback) (*PoseStampedSubscription, error) {
	callback := func(s *rclgo.Subscription) {
		var msg PoseStamped
		info, err := s.TakeMessage(&msg)
		subscriptionCallback(&msg, info, err)
	}
	sub, err := node.NewSubscription(topic_name, PoseStampedTypeSupport, opts, callback)
	if err != nil {
		return nil, err
	}
	return &PoseStampedSubscription{sub}, nil
}

func (s *PoseStampedSubscription) TakeMessage(out *PoseStamped) (*rclgo.MessageInfo, error) {
	return s.Subscription.TakeMessage(out)
}

// ClonePoseStampedSlice clones src to dst by calling Clone for each element in
// src. Panics if len(dst) < len(src).
func ClonePoseStampedSlice(dst, src []PoseStamped) {
	for i := range src {
		dst[i] = *src[i].Clone()
	}
}

// Modifying this variable is undefined behavior.
var PoseStampedTypeSupport types.MessageTypeSupport = _PoseStampedTypeSupport{}

type _PoseStampedTypeSupport struct{}

func (t _PoseStampedTypeSupport) New() types.Message {
	return NewPoseStamped()
}

func (t _PoseStampedTypeSupport) PrepareMemory() unsafe.Pointer { //returns *C.geometry_msgs__msg__PoseStamped
	return (unsafe.Pointer)(C.geometry_msgs__msg__PoseStamped__create())
}

func (t _PoseStampedTypeSupport) ReleaseMemory(pointer_to_free unsafe.Pointer) {
	C.geometry_msgs__msg__PoseStamped__destroy((*C.geometry_msgs__msg__PoseStamped)(pointer_to_free))
}

func (t _PoseStampedTypeSupport) AsCStruct(dst unsafe.Pointer, msg types.Message) {
	m := msg.(*PoseStamped)
	mem := (*C.geometry_msgs__msg__PoseStamped)(dst)
	std_msgs_msg.HeaderTypeSupport.AsCStruct(unsafe.Pointer(&mem.header), &m.Header)
	PoseTypeSupport.AsCStruct(unsafe.Pointer(&mem.pose), &m.Pose)
}

func (t _PoseStampedTypeSupport) AsGoStruct(msg types.Message, ros2_message_buffer unsafe.Pointer) {
	m := msg.(*PoseStamped)
	mem := (*C.geometry_msgs__msg__PoseStamped)(ros2_message_buffer)
	std_msgs_msg.HeaderTypeSupport.AsGoStruct(&m.Header, unsafe.Pointer(&mem.header))
	PoseTypeSupport.AsGoStruct(&m.Pose, unsafe.Pointer(&mem.pose))
}

func (t _PoseStampedTypeSupport) TypeSupport() unsafe.Pointer {
	return unsafe.Pointer(C.rosidl_typesupport_c__get_message_type_support_handle__geometry_msgs__msg__PoseStamped())
}

type CPoseStamped = C.geometry_msgs__msg__PoseStamped
type CPoseStamped__Sequence = C.geometry_msgs__msg__PoseStamped__Sequence

func PoseStamped__Sequence_to_Go(goSlice *[]PoseStamped, cSlice CPoseStamped__Sequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]PoseStamped, cSlice.size)
	src := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range src {
		PoseStampedTypeSupport.AsGoStruct(&(*goSlice)[i], unsafe.Pointer(&src[i]))
	}
}
func PoseStamped__Sequence_to_C(cSlice *CPoseStamped__Sequence, goSlice []PoseStamped) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.geometry_msgs__msg__PoseStamped)(C.malloc(C.sizeof_struct_geometry_msgs__msg__PoseStamped * C.size_t(len(goSlice))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range goSlice {
		PoseStampedTypeSupport.AsCStruct(unsafe.Pointer(&dst[i]), &goSlice[i])
	}
}
func PoseStamped__Array_to_Go(goSlice []PoseStamped, cSlice []CPoseStamped) {
	for i := 0; i < len(cSlice); i++ {
		PoseStampedTypeSupport.AsGoStruct(&goSlice[i], unsafe.Pointer(&cSlice[i]))
	}
}
func PoseStamped__Array_to_C(cSlice []CPoseStamped, goSlice []PoseStamped) {
	for i := 0; i < len(goSlice); i++ {
		PoseStampedTypeSupport.AsCStruct(unsafe.Pointer(&cSlice[i]), &goSlice[i])
	}
}
