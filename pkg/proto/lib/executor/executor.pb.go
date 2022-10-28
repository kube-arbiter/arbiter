// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: pkg/proto/executor.proto

package executor

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Kind int32

const (
	Kind_pod  Kind = 0
	Kind_node Kind = 1
)

// Enum value maps for Kind.
var (
	Kind_name = map[int32]string{
		0: "pod",
		1: "node",
	}
	Kind_value = map[string]int32{
		"pod":  0,
		"node": 1,
	}
)

func (x Kind) Enum() *Kind {
	p := new(Kind)
	*p = x
	return p
}

func (x Kind) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Kind) Descriptor() protoreflect.EnumDescriptor {
	return file_pkg_proto_executor_proto_enumTypes[0].Descriptor()
}

func (Kind) Type() protoreflect.EnumType {
	return &file_pkg_proto_executor_proto_enumTypes[0]
}

func (x Kind) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Kind.Descriptor instead.
func (Kind) EnumDescriptor() ([]byte, []int) {
	return file_pkg_proto_executor_proto_rawDescGZIP(), []int{0}
}

type ExecuteActionMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ExecuteActionMessage) Reset() {
	*x = ExecuteActionMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_executor_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteActionMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteActionMessage) ProtoMessage() {}

func (x *ExecuteActionMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_executor_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteActionMessage.ProtoReflect.Descriptor instead.
func (*ExecuteActionMessage) Descriptor() ([]byte, []int) {
	return file_pkg_proto_executor_proto_rawDescGZIP(), []int{0}
}

type ExecuteActionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Action []string `protobuf:"bytes,1,rep,name=action,proto3" json:"action,omitempty"`
}

func (x *ExecuteActionResponse) Reset() {
	*x = ExecuteActionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_executor_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteActionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteActionResponse) ProtoMessage() {}

func (x *ExecuteActionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_executor_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteActionResponse.ProtoReflect.Descriptor instead.
func (*ExecuteActionResponse) Descriptor() ([]byte, []int) {
	return file_pkg_proto_executor_proto_rawDescGZIP(), []int{1}
}

func (x *ExecuteActionResponse) GetAction() []string {
	if x != nil {
		return x.Action
	}
	return nil
}

type ExecuteMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ResourceName string                `protobuf:"bytes,1,opt,name=resourceName,proto3" json:"resourceName,omitempty"`
	Kind         Kind                  `protobuf:"varint,2,opt,name=kind,proto3,enum=execute.Kind" json:"kind,omitempty"`
	Namespace    string                `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	ExprVal      float64               `protobuf:"fixed64,4,opt,name=exprVal,proto3" json:"exprVal,omitempty"`
	CondVal      bool                  `protobuf:"varint,5,opt,name=condVal,proto3" json:"condVal,omitempty"`
	ActionData   *runtime.RawExtension `protobuf:"bytes,6,opt,name=actionData,proto3,oneof" json:"actionData,omitempty"`
	Group        string                `protobuf:"bytes,7,opt,name=group,proto3" json:"group,omitempty"`
	Version      string                `protobuf:"bytes,8,opt,name=version,proto3" json:"version,omitempty"`
	Resources    string                `protobuf:"bytes,9,opt,name=resources,proto3" json:"resources,omitempty"`
	Action       string                `protobuf:"bytes,10,opt,name=action,proto3" json:"action,omitempty"`
}

func (x *ExecuteMessage) Reset() {
	*x = ExecuteMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_executor_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteMessage) ProtoMessage() {}

func (x *ExecuteMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_executor_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteMessage.ProtoReflect.Descriptor instead.
func (*ExecuteMessage) Descriptor() ([]byte, []int) {
	return file_pkg_proto_executor_proto_rawDescGZIP(), []int{2}
}

func (x *ExecuteMessage) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *ExecuteMessage) GetKind() Kind {
	if x != nil {
		return x.Kind
	}
	return Kind_pod
}

func (x *ExecuteMessage) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *ExecuteMessage) GetExprVal() float64 {
	if x != nil {
		return x.ExprVal
	}
	return 0
}

func (x *ExecuteMessage) GetCondVal() bool {
	if x != nil {
		return x.CondVal
	}
	return false
}

func (x *ExecuteMessage) GetActionData() *runtime.RawExtension {
	if x != nil {
		return x.ActionData
	}
	return nil
}

func (x *ExecuteMessage) GetGroup() string {
	if x != nil {
		return x.Group
	}
	return ""
}

func (x *ExecuteMessage) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *ExecuteMessage) GetResources() string {
	if x != nil {
		return x.Resources
	}
	return ""
}

func (x *ExecuteMessage) GetAction() string {
	if x != nil {
		return x.Action
	}
	return ""
}

type ExecuteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data string `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ExecuteResponse) Reset() {
	*x = ExecuteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_executor_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExecuteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecuteResponse) ProtoMessage() {}

func (x *ExecuteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_executor_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecuteResponse.ProtoReflect.Descriptor instead.
func (*ExecuteResponse) Descriptor() ([]byte, []int) {
	return file_pkg_proto_executor_proto_rawDescGZIP(), []int{3}
}

func (x *ExecuteResponse) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

var File_pkg_proto_executor_proto protoreflect.FileDescriptor

var file_pkg_proto_executor_proto_rawDesc = []byte{
	0x0a, 0x18, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x65, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x1a, 0x2f, 0x6b, 0x38, 0x73, 0x2e, 0x69, 0x6f, 0x2f, 0x61, 0x70, 0x69, 0x6d,
	0x61, 0x63, 0x68, 0x69, 0x6e, 0x65, 0x72, 0x79, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x72, 0x75, 0x6e,
	0x74, 0x69, 0x6d, 0x65, 0x2f, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x16, 0x0a, 0x14, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x41,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x2f, 0x0a, 0x15,
	0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xf2, 0x02,
	0x0a, 0x0e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x2e, 0x4b, 0x69, 0x6e,
	0x64, 0x52, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x78, 0x70, 0x72, 0x56, 0x61, 0x6c,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x07, 0x65, 0x78, 0x70, 0x72, 0x56, 0x61, 0x6c, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x64, 0x56, 0x61, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x63, 0x6f, 0x6e, 0x64, 0x56, 0x61, 0x6c, 0x12, 0x52, 0x0a, 0x0a, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2d, 0x2e,
	0x6b, 0x38, 0x73, 0x2e, 0x69, 0x6f, 0x2e, 0x61, 0x70, 0x69, 0x6d, 0x61, 0x63, 0x68, 0x69, 0x6e,
	0x65, 0x72, 0x79, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x52, 0x61, 0x77, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x0a,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x12, 0x14, 0x0a,
	0x05, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a,
	0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x44, 0x61,
	0x74, 0x61, 0x22, 0x25, 0x0a, 0x0f, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x2a, 0x19, 0x0a, 0x04, 0x4b, 0x69, 0x6e,
	0x64, 0x12, 0x07, 0x0a, 0x03, 0x70, 0x6f, 0x64, 0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x6e, 0x6f,
	0x64, 0x65, 0x10, 0x01, 0x32, 0x9b, 0x01, 0x0a, 0x07, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x12, 0x50, 0x0a, 0x0d, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x1d, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x2e, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x1a, 0x1e, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75,
	0x74, 0x65, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x3e, 0x0a, 0x07, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x12, 0x17, 0x2e,
	0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x18, 0x2e, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65,
	0x2e, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x42, 0x0f, 0x5a, 0x0d, 0x6c, 0x69, 0x62, 0x2f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74,
	0x6f, 0x72, 0x2f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_proto_executor_proto_rawDescOnce sync.Once
	file_pkg_proto_executor_proto_rawDescData = file_pkg_proto_executor_proto_rawDesc
)

func file_pkg_proto_executor_proto_rawDescGZIP() []byte {
	file_pkg_proto_executor_proto_rawDescOnce.Do(func() {
		file_pkg_proto_executor_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_proto_executor_proto_rawDescData)
	})
	return file_pkg_proto_executor_proto_rawDescData
}

var file_pkg_proto_executor_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_pkg_proto_executor_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pkg_proto_executor_proto_goTypes = []interface{}{
	(Kind)(0),                     // 0: execute.Kind
	(*ExecuteActionMessage)(nil),  // 1: execute.ExecuteActionMessage
	(*ExecuteActionResponse)(nil), // 2: execute.ExecuteActionResponse
	(*ExecuteMessage)(nil),        // 3: execute.ExecuteMessage
	(*ExecuteResponse)(nil),       // 4: execute.ExecuteResponse
	(*runtime.RawExtension)(nil),  // 5: k8s.io.apimachinery.pkg.runtime.RawExtension
}
var file_pkg_proto_executor_proto_depIdxs = []int32{
	0, // 0: execute.ExecuteMessage.kind:type_name -> execute.Kind
	5, // 1: execute.ExecuteMessage.actionData:type_name -> k8s.io.apimachinery.pkg.runtime.RawExtension
	1, // 2: execute.Execute.ExecuteAction:input_type -> execute.ExecuteActionMessage
	3, // 3: execute.Execute.Execute:input_type -> execute.ExecuteMessage
	2, // 4: execute.Execute.ExecuteAction:output_type -> execute.ExecuteActionResponse
	4, // 5: execute.Execute.Execute:output_type -> execute.ExecuteResponse
	4, // [4:6] is the sub-list for method output_type
	2, // [2:4] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_proto_executor_proto_init() }
func file_pkg_proto_executor_proto_init() {
	if File_pkg_proto_executor_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_proto_executor_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteActionMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_proto_executor_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteActionResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_proto_executor_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_proto_executor_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExecuteResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_pkg_proto_executor_proto_msgTypes[2].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_proto_executor_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_proto_executor_proto_goTypes,
		DependencyIndexes: file_pkg_proto_executor_proto_depIdxs,
		EnumInfos:         file_pkg_proto_executor_proto_enumTypes,
		MessageInfos:      file_pkg_proto_executor_proto_msgTypes,
	}.Build()
	File_pkg_proto_executor_proto = out.File
	file_pkg_proto_executor_proto_rawDesc = nil
	file_pkg_proto_executor_proto_goTypes = nil
	file_pkg_proto_executor_proto_depIdxs = nil
}