// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.3
// source: observiability.proto

package observer

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetPluginNameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetPluginNameRequest) Reset() {
	*x = GetPluginNameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPluginNameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPluginNameRequest) ProtoMessage() {}

func (x *GetPluginNameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPluginNameRequest.ProtoReflect.Descriptor instead.
func (*GetPluginNameRequest) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{0}
}

type GetPluginNameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Attr map[string]string `protobuf:"bytes,2,rep,name=attr,proto3" json:"attr,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *GetPluginNameResponse) Reset() {
	*x = GetPluginNameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPluginNameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPluginNameResponse) ProtoMessage() {}

func (x *GetPluginNameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPluginNameResponse.ProtoReflect.Descriptor instead.
func (*GetPluginNameResponse) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{1}
}

func (x *GetPluginNameResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *GetPluginNameResponse) GetAttr() map[string]string {
	if x != nil {
		return x.Attr
	}
	return nil
}

type PluginCapabilitiesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PluginCapabilitiesRequest) Reset() {
	*x = PluginCapabilitiesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginCapabilitiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginCapabilitiesRequest) ProtoMessage() {}

func (x *PluginCapabilitiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginCapabilitiesRequest.ProtoReflect.Descriptor instead.
func (*PluginCapabilitiesRequest) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{2}
}

type MetricInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// cpu_usage
	MetricUnit  string   `protobuf:"bytes,1,opt,name=metric_unit,json=metricUnit,proto3" json:"metric_unit,omitempty"`
	Description string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	Aggregation []string `protobuf:"bytes,3,rep,name=aggregation,proto3" json:"aggregation,omitempty"`
}

func (x *MetricInfo) Reset() {
	*x = MetricInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricInfo) ProtoMessage() {}

func (x *MetricInfo) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricInfo.ProtoReflect.Descriptor instead.
func (*MetricInfo) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{3}
}

func (x *MetricInfo) GetMetricUnit() string {
	if x != nil {
		return x.MetricUnit
	}
	return ""
}

func (x *MetricInfo) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *MetricInfo) GetAggregation() []string {
	if x != nil {
		return x.Aggregation
	}
	return nil
}

type PluginCapabilitiesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// {"cpu_usage": {"metric_unit":"c","aggregation": ["a","b"]}}
	MetricInfo map[string]*MetricInfo `protobuf:"bytes,2,rep,name=metric_info,json=metricInfo,proto3" json:"metric_info,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *PluginCapabilitiesResponse) Reset() {
	*x = PluginCapabilitiesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PluginCapabilitiesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PluginCapabilitiesResponse) ProtoMessage() {}

func (x *PluginCapabilitiesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PluginCapabilitiesResponse.ProtoReflect.Descriptor instead.
func (*PluginCapabilitiesResponse) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{4}
}

func (x *PluginCapabilitiesResponse) GetMetricInfo() map[string]*MetricInfo {
	if x != nil {
		return x.MetricInfo
	}
	return nil
}

type GetMetricsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// pod name or node name.
	// if there is a query field, ignore the resource_name field
	ResourceNames []string `protobuf:"bytes,1,rep,name=resource_names,json=resourceNames,proto3" json:"resource_names,omitempty"`
	Namespace     string   `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// cpu_usage, memory_usage eg.
	MetricName string `protobuf:"bytes,3,opt,name=metric_name,json=metricName,proto3" json:"metric_name,omitempty"`
	// max, min, avg. And so on, some ways to aggregate data
	Aggregation []string `protobuf:"bytes,4,rep,name=aggregation,proto3" json:"aggregation,omitempty"`
	// such as promql
	Query string `protobuf:"bytes,5,opt,name=query,proto3" json:"query,omitempty"`
	// 1650867976563 million second
	// 这个请求时间做一个保留
	StartTime int64 `protobuf:"varint,6,opt,name=start_time,json=startTime,proto3" json:"start_time,omitempty"`
	EndTime   int64 `protobuf:"varint,7,opt,name=end_time,json=endTime,proto3" json:"end_time,omitempty"`
	// resource kind, Pod, Node etc.
	Kind string `protobuf:"bytes,8,opt,name=kind,proto3" json:"kind,omitempty"`
	Unit string `protobuf:"bytes,9,opt,name=unit,proto3" json:"unit,omitempty"`
}

func (x *GetMetricsRequest) Reset() {
	*x = GetMetricsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMetricsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetricsRequest) ProtoMessage() {}

func (x *GetMetricsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetricsRequest.ProtoReflect.Descriptor instead.
func (*GetMetricsRequest) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{5}
}

func (x *GetMetricsRequest) GetResourceNames() []string {
	if x != nil {
		return x.ResourceNames
	}
	return nil
}

func (x *GetMetricsRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *GetMetricsRequest) GetMetricName() string {
	if x != nil {
		return x.MetricName
	}
	return ""
}

func (x *GetMetricsRequest) GetAggregation() []string {
	if x != nil {
		return x.Aggregation
	}
	return nil
}

func (x *GetMetricsRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *GetMetricsRequest) GetStartTime() int64 {
	if x != nil {
		return x.StartTime
	}
	return 0
}

func (x *GetMetricsRequest) GetEndTime() int64 {
	if x != nil {
		return x.EndTime
	}
	return 0
}

func (x *GetMetricsRequest) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *GetMetricsRequest) GetUnit() string {
	if x != nil {
		return x.Unit
	}
	return ""
}

type GetMetricsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// maybe empty,
	// For example, metrics-server can return resource_name,
	// which is the acquisition of the instantaneous value of the data.
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	Namespace    string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// maybe empty
	Unit    string                      `protobuf:"bytes,3,opt,name=unit,proto3" json:"unit,omitempty"`
	Records []*GetMetricsResponseRecord `protobuf:"bytes,4,rep,name=records,proto3" json:"records,omitempty"` //map<string,double> values = 4;
}

func (x *GetMetricsResponse) Reset() {
	*x = GetMetricsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMetricsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetricsResponse) ProtoMessage() {}

func (x *GetMetricsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetricsResponse.ProtoReflect.Descriptor instead.
func (*GetMetricsResponse) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{6}
}

func (x *GetMetricsResponse) GetResourceName() string {
	if x != nil {
		return x.ResourceName
	}
	return ""
}

func (x *GetMetricsResponse) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *GetMetricsResponse) GetUnit() string {
	if x != nil {
		return x.Unit
	}
	return ""
}

func (x *GetMetricsResponse) GetRecords() []*GetMetricsResponseRecord {
	if x != nil {
		return x.Records
	}
	return nil
}

// the value field is string that is serialized by json.
// {"cpu_usage": 0.1}
type GetMetricsResponseRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp int64  `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Value     string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *GetMetricsResponseRecord) Reset() {
	*x = GetMetricsResponseRecord{}
	if protoimpl.UnsafeEnabled {
		mi := &file_observiability_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMetricsResponseRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetricsResponseRecord) ProtoMessage() {}

func (x *GetMetricsResponseRecord) ProtoReflect() protoreflect.Message {
	mi := &file_observiability_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetricsResponseRecord.ProtoReflect.Descriptor instead.
func (*GetMetricsResponseRecord) Descriptor() ([]byte, []int) {
	return file_observiability_proto_rawDescGZIP(), []int{6, 0}
}

func (x *GetMetricsResponseRecord) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *GetMetricsResponseRecord) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_observiability_proto protoreflect.FileDescriptor

var file_observiability_proto_rawDesc = []byte{
	0x0a, 0x14, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x69, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x22, 0x16,
	0x0a, 0x14, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xa1, 0x01, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x04, 0x61, 0x74, 0x74, 0x72, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x27, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x41, 0x74, 0x74, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x04, 0x61, 0x74, 0x74,
	0x72, 0x1a, 0x37, 0x0a, 0x09, 0x41, 0x74, 0x74, 0x72, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x1b, 0x0a, 0x19, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x71, 0x0a, 0x0a, 0x4d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f,
	0x75, 0x6e, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x55, 0x6e, 0x69, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x61, 0x67, 0x67, 0x72,
	0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x61,
	0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xc4, 0x01, 0x0a, 0x1a, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x53, 0x0a, 0x0b, 0x6d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x32,
	0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x61,
	0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x0a, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x51,
	0x0a, 0x0f, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x28, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x93, 0x02, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x25, 0x0a, 0x0e, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x12, 0x1c,
	0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a,
	0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x0b, 0x61, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x65, 0x6e, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b,
	0x69, 0x6e, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x6e, 0x69, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x75, 0x6e, 0x69, 0x74, 0x22, 0xe6, 0x01, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x23,
	0x0a, 0x0d, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x6e, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x75, 0x6e, 0x69, 0x74, 0x12, 0x3b, 0x0a, 0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73,
	0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x2e, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x52, 0x07, 0x72, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x73, 0x1a, 0x3c, 0x0a, 0x06, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x32, 0xfe, 0x01, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x4e, 0x0a, 0x0d, 0x47,
	0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x2e, 0x6f,
	0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e,
	0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x6f, 0x62, 0x69,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x4e, 0x61, 0x6d,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x5d, 0x0a, 0x12, 0x50,
	0x6c, 0x75, 0x67, 0x69, 0x6e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65,
	0x73, 0x12, 0x21, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x22, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6c,
	0x75, 0x67, 0x69, 0x6e, 0x43, 0x61, 0x70, 0x61, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x45, 0x0a, 0x0a, 0x47, 0x65,
	0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x19, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76,
	0x31, 0x2e, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x6f, 0x62, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0x0e, 0x5a, 0x0c, 0x6c, 0x69, 0x62, 0x2f, 0x6f, 0x62, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_observiability_proto_rawDescOnce sync.Once
	file_observiability_proto_rawDescData = file_observiability_proto_rawDesc
)

func file_observiability_proto_rawDescGZIP() []byte {
	file_observiability_proto_rawDescOnce.Do(func() {
		file_observiability_proto_rawDescData = protoimpl.X.CompressGZIP(file_observiability_proto_rawDescData)
	})
	return file_observiability_proto_rawDescData
}

var file_observiability_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_observiability_proto_goTypes = []interface{}{
	(*GetPluginNameRequest)(nil),       // 0: obi.v1.GetPluginNameRequest
	(*GetPluginNameResponse)(nil),      // 1: obi.v1.GetPluginNameResponse
	(*PluginCapabilitiesRequest)(nil),  // 2: obi.v1.PluginCapabilitiesRequest
	(*MetricInfo)(nil),                 // 3: obi.v1.MetricInfo
	(*PluginCapabilitiesResponse)(nil), // 4: obi.v1.PluginCapabilitiesResponse
	(*GetMetricsRequest)(nil),          // 5: obi.v1.GetMetricsRequest
	(*GetMetricsResponse)(nil),         // 6: obi.v1.GetMetricsResponse
	nil,                                // 7: obi.v1.GetPluginNameResponse.AttrEntry
	nil,                                // 8: obi.v1.PluginCapabilitiesResponse.MetricInfoEntry
	(*GetMetricsResponseRecord)(nil),   // 9: obi.v1.GetMetricsResponse.record
}
var file_observiability_proto_depIdxs = []int32{
	7, // 0: obi.v1.GetPluginNameResponse.attr:type_name -> obi.v1.GetPluginNameResponse.AttrEntry
	8, // 1: obi.v1.PluginCapabilitiesResponse.metric_info:type_name -> obi.v1.PluginCapabilitiesResponse.MetricInfoEntry
	9, // 2: obi.v1.GetMetricsResponse.records:type_name -> obi.v1.GetMetricsResponse.record
	3, // 3: obi.v1.PluginCapabilitiesResponse.MetricInfoEntry.value:type_name -> obi.v1.MetricInfo
	0, // 4: obi.v1.Server.GetPluginName:input_type -> obi.v1.GetPluginNameRequest
	2, // 5: obi.v1.Server.PluginCapabilities:input_type -> obi.v1.PluginCapabilitiesRequest
	5, // 6: obi.v1.Server.GetMetrics:input_type -> obi.v1.GetMetricsRequest
	1, // 7: obi.v1.Server.GetPluginName:output_type -> obi.v1.GetPluginNameResponse
	4, // 8: obi.v1.Server.PluginCapabilities:output_type -> obi.v1.PluginCapabilitiesResponse
	6, // 9: obi.v1.Server.GetMetrics:output_type -> obi.v1.GetMetricsResponse
	7, // [7:10] is the sub-list for method output_type
	4, // [4:7] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_observiability_proto_init() }
func file_observiability_proto_init() {
	if File_observiability_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_observiability_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPluginNameRequest); i {
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
		file_observiability_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPluginNameResponse); i {
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
		file_observiability_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginCapabilitiesRequest); i {
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
		file_observiability_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricInfo); i {
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
		file_observiability_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PluginCapabilitiesResponse); i {
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
		file_observiability_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMetricsRequest); i {
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
		file_observiability_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMetricsResponse); i {
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
		file_observiability_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMetricsResponseRecord); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_observiability_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_observiability_proto_goTypes,
		DependencyIndexes: file_observiability_proto_depIdxs,
		MessageInfos:      file_observiability_proto_msgTypes,
	}.Build()
	File_observiability_proto = out.File
	file_observiability_proto_rawDesc = nil
	file_observiability_proto_goTypes = nil
	file_observiability_proto_depIdxs = nil
}
