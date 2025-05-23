// Code generated by rclgo-gen. DO NOT EDIT.

package nav_msgs_msg
import (
	"unsafe"

	"github.com/tiiuae/rclgo/pkg/rclgo"
	"github.com/tiiuae/rclgo/pkg/rclgo/types"
	"github.com/tiiuae/rclgo/pkg/rclgo/typemap"
	geometry_msgs_msg "thornol/internal/rclgo/generated/geometry_msgs/msg"
	std_msgs_msg "thornol/internal/rclgo/generated/std_msgs/msg"
)
/*
#include <rosidl_runtime_c/message_type_support_struct.h>

#include <nav_msgs/msg/grid_cells.h>

*/
import "C"

func init() {
	typemap.RegisterMessage("nav_msgs/GridCells", GridCellsTypeSupport)
	typemap.RegisterMessage("nav_msgs/msg/GridCells", GridCellsTypeSupport)
}

type GridCells struct {
	Header std_msgs_msg.Header `yaml:"header"`
	CellWidth float32 `yaml:"cell_width"`// Width of each cell
	CellHeight float32 `yaml:"cell_height"`// Height of each cell
	Cells []geometry_msgs_msg.Point `yaml:"cells"`// Each cell is represented by the Point at the center of the cell
}

// NewGridCells creates a new GridCells with default values.
func NewGridCells() *GridCells {
	self := GridCells{}
	self.SetDefaults()
	return &self
}

func (t *GridCells) Clone() *GridCells {
	c := &GridCells{}
	c.Header = *t.Header.Clone()
	c.CellWidth = t.CellWidth
	c.CellHeight = t.CellHeight
	if t.Cells != nil {
		c.Cells = make([]geometry_msgs_msg.Point, len(t.Cells))
		geometry_msgs_msg.ClonePointSlice(c.Cells, t.Cells)
	}
	return c
}

func (t *GridCells) CloneMsg() types.Message {
	return t.Clone()
}

func (t *GridCells) SetDefaults() {
	t.Header.SetDefaults()
	t.CellWidth = 0
	t.CellHeight = 0
	t.Cells = nil
}

func (t *GridCells) GetTypeSupport() types.MessageTypeSupport {
	return GridCellsTypeSupport
}

// GridCellsPublisher wraps rclgo.Publisher to provide type safe helper
// functions
type GridCellsPublisher struct {
	*rclgo.Publisher
}

// NewGridCellsPublisher creates and returns a new publisher for the
// GridCells
func NewGridCellsPublisher(node *rclgo.Node, topic_name string, options *rclgo.PublisherOptions) (*GridCellsPublisher, error) {
	pub, err := node.NewPublisher(topic_name, GridCellsTypeSupport, options)
	if err != nil {
		return nil, err
	}
	return &GridCellsPublisher{pub}, nil
}

func (p *GridCellsPublisher) Publish(msg *GridCells) error {
	return p.Publisher.Publish(msg)
}

// GridCellsSubscription wraps rclgo.Subscription to provide type safe helper
// functions
type GridCellsSubscription struct {
	*rclgo.Subscription
}

// GridCellsSubscriptionCallback type is used to provide a subscription
// handler function for a GridCellsSubscription.
type GridCellsSubscriptionCallback func(msg *GridCells, info *rclgo.MessageInfo, err error)

// NewGridCellsSubscription creates and returns a new subscription for the
// GridCells
func NewGridCellsSubscription(node *rclgo.Node, topic_name string, opts *rclgo.SubscriptionOptions, subscriptionCallback GridCellsSubscriptionCallback) (*GridCellsSubscription, error) {
	callback := func(s *rclgo.Subscription) {
		var msg GridCells
		info, err := s.TakeMessage(&msg)
		subscriptionCallback(&msg, info, err)
	}
	sub, err := node.NewSubscription(topic_name, GridCellsTypeSupport, opts, callback)
	if err != nil {
		return nil, err
	}
	return &GridCellsSubscription{sub}, nil
}

func (s *GridCellsSubscription) TakeMessage(out *GridCells) (*rclgo.MessageInfo, error) {
	return s.Subscription.TakeMessage(out)
}

// CloneGridCellsSlice clones src to dst by calling Clone for each element in
// src. Panics if len(dst) < len(src).
func CloneGridCellsSlice(dst, src []GridCells) {
	for i := range src {
		dst[i] = *src[i].Clone()
	}
}

// Modifying this variable is undefined behavior.
var GridCellsTypeSupport types.MessageTypeSupport = _GridCellsTypeSupport{}

type _GridCellsTypeSupport struct{}

func (t _GridCellsTypeSupport) New() types.Message {
	return NewGridCells()
}

func (t _GridCellsTypeSupport) PrepareMemory() unsafe.Pointer { //returns *C.nav_msgs__msg__GridCells
	return (unsafe.Pointer)(C.nav_msgs__msg__GridCells__create())
}

func (t _GridCellsTypeSupport) ReleaseMemory(pointer_to_free unsafe.Pointer) {
	C.nav_msgs__msg__GridCells__destroy((*C.nav_msgs__msg__GridCells)(pointer_to_free))
}

func (t _GridCellsTypeSupport) AsCStruct(dst unsafe.Pointer, msg types.Message) {
	m := msg.(*GridCells)
	mem := (*C.nav_msgs__msg__GridCells)(dst)
	std_msgs_msg.HeaderTypeSupport.AsCStruct(unsafe.Pointer(&mem.header), &m.Header)
	mem.cell_width = C.float(m.CellWidth)
	mem.cell_height = C.float(m.CellHeight)
	geometry_msgs_msg.Point__Sequence_to_C((*geometry_msgs_msg.CPoint__Sequence)(unsafe.Pointer(&mem.cells)), m.Cells)
}

func (t _GridCellsTypeSupport) AsGoStruct(msg types.Message, ros2_message_buffer unsafe.Pointer) {
	m := msg.(*GridCells)
	mem := (*C.nav_msgs__msg__GridCells)(ros2_message_buffer)
	std_msgs_msg.HeaderTypeSupport.AsGoStruct(&m.Header, unsafe.Pointer(&mem.header))
	m.CellWidth = float32(mem.cell_width)
	m.CellHeight = float32(mem.cell_height)
	geometry_msgs_msg.Point__Sequence_to_Go(&m.Cells, *(*geometry_msgs_msg.CPoint__Sequence)(unsafe.Pointer(&mem.cells)))
}

func (t _GridCellsTypeSupport) TypeSupport() unsafe.Pointer {
	return unsafe.Pointer(C.rosidl_typesupport_c__get_message_type_support_handle__nav_msgs__msg__GridCells())
}

type CGridCells = C.nav_msgs__msg__GridCells
type CGridCells__Sequence = C.nav_msgs__msg__GridCells__Sequence

func GridCells__Sequence_to_Go(goSlice *[]GridCells, cSlice CGridCells__Sequence) {
	if cSlice.size == 0 {
		return
	}
	*goSlice = make([]GridCells, cSlice.size)
	src := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range src {
		GridCellsTypeSupport.AsGoStruct(&(*goSlice)[i], unsafe.Pointer(&src[i]))
	}
}
func GridCells__Sequence_to_C(cSlice *CGridCells__Sequence, goSlice []GridCells) {
	if len(goSlice) == 0 {
		cSlice.data = nil
		cSlice.capacity = 0
		cSlice.size = 0
		return
	}
	cSlice.data = (*C.nav_msgs__msg__GridCells)(C.malloc(C.sizeof_struct_nav_msgs__msg__GridCells * C.size_t(len(goSlice))))
	cSlice.capacity = C.size_t(len(goSlice))
	cSlice.size = cSlice.capacity
	dst := unsafe.Slice(cSlice.data, cSlice.size)
	for i := range goSlice {
		GridCellsTypeSupport.AsCStruct(unsafe.Pointer(&dst[i]), &goSlice[i])
	}
}
func GridCells__Array_to_Go(goSlice []GridCells, cSlice []CGridCells) {
	for i := 0; i < len(cSlice); i++ {
		GridCellsTypeSupport.AsGoStruct(&goSlice[i], unsafe.Pointer(&cSlice[i]))
	}
}
func GridCells__Array_to_C(cSlice []CGridCells, goSlice []GridCells) {
	for i := 0; i < len(goSlice); i++ {
		GridCellsTypeSupport.AsCStruct(unsafe.Pointer(&cSlice[i]), &goSlice[i])
	}
}
