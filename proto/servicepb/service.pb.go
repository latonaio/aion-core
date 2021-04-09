// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.15.6
// source: proto/servicepb/service.proto

package servicepb

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Microservice struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command             []string                `protobuf:"bytes,1,rep,name=command,proto3" json:"command,omitempty"`
	NextService         map[string]*NextService `protobuf:"bytes,2,rep,name=nextService,proto3" json:"nextService,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Scale               int32                   `protobuf:"varint,3,opt,name=Scale,proto3" json:"Scale,omitempty"`
	Env                 map[string]string       `protobuf:"bytes,4,rep,name=env,proto3" json:"env,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Position            string                  `protobuf:"bytes,5,opt,name=position,proto3" json:"position,omitempty"`
	Always              bool                    `protobuf:"varint,6,opt,name=always,proto3" json:"always,omitempty"`
	Multiple            bool                    `protobuf:"varint,7,opt,name=multiple,proto3" json:"multiple,omitempty"`
	Docker              bool                    `protobuf:"varint,8,opt,name=docker,proto3" json:"docker,omitempty"`
	Startup             bool                    `protobuf:"varint,9,opt,name=startup,proto3" json:"startup,omitempty"`
	Interval            int32                   `protobuf:"varint,10,opt,name=interval,proto3" json:"interval,omitempty"`
	Ports               []*PortConfig           `protobuf:"bytes,11,rep,name=ports,proto3" json:"ports,omitempty"`
	DirPath             string                  `protobuf:"bytes,12,opt,name=dirPath,proto3" json:"dirPath,omitempty"`
	ServiceAccount      string                  `protobuf:"bytes,13,opt,name=serviceAccount,proto3" json:"serviceAccount,omitempty"`
	Network             string                  `protobuf:"bytes,14,opt,name=network,proto3" json:"network,omitempty"`
	Tag                 string                  `protobuf:"bytes,15,opt,name=tag,proto3" json:"tag,omitempty"`
	VolumeMountPathList []string                `protobuf:"bytes,16,rep,name=volumeMountPathList,proto3" json:"volumeMountPathList,omitempty"`
	Privileged          bool                    `protobuf:"varint,17,opt,name=privileged,proto3" json:"privileged,omitempty"`
	WithoutKanban       bool                    `protobuf:"varint,18,opt,name=withoutKanban,proto3" json:"withoutKanban,omitempty"`
	TargetNode          string                  `protobuf:"bytes,19,opt,name=targetNode,proto3" json:"targetNode,omitempty"`
	Resources           *Resources              `protobuf:"bytes,20,opt,name=resources,proto3" json:"resources,omitempty"`
}

func (x *Microservice) Reset() {
	*x = Microservice{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Microservice) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Microservice) ProtoMessage() {}

func (x *Microservice) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Microservice.ProtoReflect.Descriptor instead.
func (*Microservice) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{0}
}

func (x *Microservice) GetCommand() []string {
	if x != nil {
		return x.Command
	}
	return nil
}

func (x *Microservice) GetNextService() map[string]*NextService {
	if x != nil {
		return x.NextService
	}
	return nil
}

func (x *Microservice) GetScale() int32 {
	if x != nil {
		return x.Scale
	}
	return 0
}

func (x *Microservice) GetEnv() map[string]string {
	if x != nil {
		return x.Env
	}
	return nil
}

func (x *Microservice) GetPosition() string {
	if x != nil {
		return x.Position
	}
	return ""
}

func (x *Microservice) GetAlways() bool {
	if x != nil {
		return x.Always
	}
	return false
}

func (x *Microservice) GetMultiple() bool {
	if x != nil {
		return x.Multiple
	}
	return false
}

func (x *Microservice) GetDocker() bool {
	if x != nil {
		return x.Docker
	}
	return false
}

func (x *Microservice) GetStartup() bool {
	if x != nil {
		return x.Startup
	}
	return false
}

func (x *Microservice) GetInterval() int32 {
	if x != nil {
		return x.Interval
	}
	return 0
}

func (x *Microservice) GetPorts() []*PortConfig {
	if x != nil {
		return x.Ports
	}
	return nil
}

func (x *Microservice) GetDirPath() string {
	if x != nil {
		return x.DirPath
	}
	return ""
}

func (x *Microservice) GetServiceAccount() string {
	if x != nil {
		return x.ServiceAccount
	}
	return ""
}

func (x *Microservice) GetNetwork() string {
	if x != nil {
		return x.Network
	}
	return ""
}

func (x *Microservice) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

func (x *Microservice) GetVolumeMountPathList() []string {
	if x != nil {
		return x.VolumeMountPathList
	}
	return nil
}

func (x *Microservice) GetPrivileged() bool {
	if x != nil {
		return x.Privileged
	}
	return false
}

func (x *Microservice) GetWithoutKanban() bool {
	if x != nil {
		return x.WithoutKanban
	}
	return false
}

func (x *Microservice) GetTargetNode() string {
	if x != nil {
		return x.TargetNode
	}
	return ""
}

func (x *Microservice) GetResources() *Resources {
	if x != nil {
		return x.Resources
	}
	return nil
}

type PortConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Protocol string `protobuf:"bytes,2,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Port     int32  `protobuf:"varint,3,opt,name=port,proto3" json:"port,omitempty"`
	NodePort int32  `protobuf:"varint,4,opt,name=nodePort,proto3" json:"nodePort,omitempty"`
}

func (x *PortConfig) Reset() {
	*x = PortConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PortConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PortConfig) ProtoMessage() {}

func (x *PortConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PortConfig.ProtoReflect.Descriptor instead.
func (*PortConfig) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{1}
}

func (x *PortConfig) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *PortConfig) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *PortConfig) GetPort() int32 {
	if x != nil {
		return x.Port
	}
	return 0
}

func (x *PortConfig) GetNodePort() int32 {
	if x != nil {
		return x.NodePort
	}
	return 0
}

type NextService struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NextServiceSetting []*NextServiceSetting `protobuf:"bytes,1,rep,name=nextServiceSetting,proto3" json:"nextServiceSetting,omitempty"`
}

func (x *NextService) Reset() {
	*x = NextService{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NextService) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NextService) ProtoMessage() {}

func (x *NextService) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NextService.ProtoReflect.Descriptor instead.
func (*NextService) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{2}
}

func (x *NextService) GetNextServiceSetting() []*NextServiceSetting {
	if x != nil {
		return x.NextServiceSetting
	}
	return nil
}

type NextServiceSetting struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NextServiceName string `protobuf:"bytes,1,opt,name=nextServiceName,proto3" json:"nextServiceName,omitempty"`
	NumberPattern   string `protobuf:"bytes,2,opt,name=numberPattern,proto3" json:"numberPattern,omitempty"`
	NextDevice      string `protobuf:"bytes,3,opt,name=nextDevice,proto3" json:"nextDevice,omitempty"`
}

func (x *NextServiceSetting) Reset() {
	*x = NextServiceSetting{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NextServiceSetting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NextServiceSetting) ProtoMessage() {}

func (x *NextServiceSetting) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NextServiceSetting.ProtoReflect.Descriptor instead.
func (*NextServiceSetting) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{3}
}

func (x *NextServiceSetting) GetNextServiceName() string {
	if x != nil {
		return x.NextServiceName
	}
	return ""
}

func (x *NextServiceSetting) GetNumberPattern() string {
	if x != nil {
		return x.NumberPattern
	}
	return ""
}

func (x *NextServiceSetting) GetNextDevice() string {
	if x != nil {
		return x.NextDevice
	}
	return ""
}

type Resources struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Requests *ResourceConfig `protobuf:"bytes,1,opt,name=requests,proto3" json:"requests,omitempty"`
	Limits   *ResourceConfig `protobuf:"bytes,2,opt,name=limits,proto3" json:"limits,omitempty"`
}

func (x *Resources) Reset() {
	*x = Resources{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Resources) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resources) ProtoMessage() {}

func (x *Resources) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resources.ProtoReflect.Descriptor instead.
func (*Resources) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{4}
}

func (x *Resources) GetRequests() *ResourceConfig {
	if x != nil {
		return x.Requests
	}
	return nil
}

func (x *Resources) GetLimits() *ResourceConfig {
	if x != nil {
		return x.Limits
	}
	return nil
}

type ResourceConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Memory string `protobuf:"bytes,1,opt,name=memory,proto3" json:"memory,omitempty"`
	Cpu    string `protobuf:"bytes,2,opt,name=cpu,proto3" json:"cpu,omitempty"`
}

func (x *ResourceConfig) Reset() {
	*x = ResourceConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_servicepb_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceConfig) ProtoMessage() {}

func (x *ResourceConfig) ProtoReflect() protoreflect.Message {
	mi := &file_proto_servicepb_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceConfig.ProtoReflect.Descriptor instead.
func (*ResourceConfig) Descriptor() ([]byte, []int) {
	return file_proto_servicepb_service_proto_rawDescGZIP(), []int{5}
}

func (x *ResourceConfig) GetMemory() string {
	if x != nil {
		return x.Memory
	}
	return ""
}

func (x *ResourceConfig) GetCpu() string {
	if x != nil {
		return x.Cpu
	}
	return ""
}

var File_proto_servicepb_service_proto protoreflect.FileDescriptor

var file_proto_servicepb_service_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70,
	0x62, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x22, 0xd3, 0x06, 0x0a, 0x0c, 0x4d,
	0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x4a, 0x0a, 0x0b, 0x6e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x2e, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x6e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x12, 0x14, 0x0a, 0x05, 0x53, 0x63, 0x61, 0x6c, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x05, 0x53, 0x63, 0x61, 0x6c, 0x65, 0x12, 0x32, 0x0a, 0x03, 0x65, 0x6e, 0x76, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62,
	0x2e, 0x4d, 0x69, 0x63, 0x72, 0x6f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x45, 0x6e,
	0x76, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x65, 0x6e, 0x76, 0x12, 0x1a, 0x0a, 0x08, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6c, 0x77, 0x61, 0x79,
	0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x61, 0x6c, 0x77, 0x61, 0x79, 0x73, 0x12,
	0x1a, 0x0a, 0x08, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x08, 0x6d, 0x75, 0x6c, 0x74, 0x69, 0x70, 0x6c, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x64,
	0x6f, 0x63, 0x6b, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x64, 0x6f, 0x63,
	0x6b, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x75, 0x70, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x75, 0x70, 0x12, 0x1a, 0x0a,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x2b, 0x0a, 0x05, 0x70, 0x6f, 0x72,
	0x74, 0x73, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x70, 0x62, 0x2e, 0x50, 0x6f, 0x72, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52,
	0x05, 0x70, 0x6f, 0x72, 0x74, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x64, 0x69, 0x72, 0x50, 0x61, 0x74,
	0x68, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x64, 0x69, 0x72, 0x50, 0x61, 0x74, 0x68,
	0x12, 0x26, 0x0a, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6e, 0x65, 0x74, 0x77,
	0x6f, 0x72, 0x6b, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x74, 0x61, 0x67, 0x12, 0x30, 0x0a, 0x13, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x4d, 0x6f,
	0x75, 0x6e, 0x74, 0x50, 0x61, 0x74, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x18, 0x10, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x13, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x4d, 0x6f, 0x75, 0x6e, 0x74, 0x50, 0x61,
	0x74, 0x68, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x69, 0x76, 0x69, 0x6c,
	0x65, 0x67, 0x65, 0x64, 0x18, 0x11, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x70, 0x72, 0x69, 0x76,
	0x69, 0x6c, 0x65, 0x67, 0x65, 0x64, 0x12, 0x24, 0x0a, 0x0d, 0x77, 0x69, 0x74, 0x68, 0x6f, 0x75,
	0x74, 0x4b, 0x61, 0x6e, 0x62, 0x61, 0x6e, 0x18, 0x12, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x77,
	0x69, 0x74, 0x68, 0x6f, 0x75, 0x74, 0x4b, 0x61, 0x6e, 0x62, 0x61, 0x6e, 0x12, 0x1e, 0x0a, 0x0a,
	0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x18, 0x13, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x32, 0x0a, 0x09,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x52, 0x09, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x1a, 0x56, 0x0a, 0x10, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70,
	0x62, 0x2e, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x36, 0x0a, 0x08, 0x45, 0x6e, 0x76, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0x6c, 0x0a, 0x0a, 0x50, 0x6f, 0x72, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x6f,
	0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6e, 0x6f, 0x64, 0x65, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x5c,
	0x0a, 0x0b, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4d, 0x0a,
	0x12, 0x6e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x70, 0x62, 0x2e, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x52, 0x12, 0x6e, 0x65, 0x78, 0x74, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x22, 0x84, 0x01, 0x0a,
	0x12, 0x4e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x53, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x12, 0x28, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6e, 0x65,
	0x78, 0x74, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a,
	0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x50, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x50, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x65, 0x78, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6e, 0x65, 0x78, 0x74, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x22, 0x75, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73,
	0x12, 0x35, 0x0a, 0x08, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x19, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x2e, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x08, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x12, 0x31, 0x0a, 0x06, 0x6c, 0x69, 0x6d, 0x69, 0x74,
	0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x52, 0x06, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x73, 0x22, 0x3a, 0x0a, 0x0e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x16, 0x0a, 0x06,
	0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65,
	0x6d, 0x6f, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x70, 0x75, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x63, 0x70, 0x75, 0x42, 0x32, 0x5a, 0x30, 0x62, 0x69, 0x74, 0x62, 0x75, 0x63,
	0x6b, 0x65, 0x74, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x61, 0x74, 0x6f, 0x6e, 0x61, 0x69, 0x6f,
	0x2f, 0x61, 0x69, 0x6f, 0x6e, 0x2d, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_proto_servicepb_service_proto_rawDescOnce sync.Once
	file_proto_servicepb_service_proto_rawDescData = file_proto_servicepb_service_proto_rawDesc
)

func file_proto_servicepb_service_proto_rawDescGZIP() []byte {
	file_proto_servicepb_service_proto_rawDescOnce.Do(func() {
		file_proto_servicepb_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_servicepb_service_proto_rawDescData)
	})
	return file_proto_servicepb_service_proto_rawDescData
}

var file_proto_servicepb_service_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_proto_servicepb_service_proto_goTypes = []interface{}{
	(*Microservice)(nil),       // 0: servicepb.Microservice
	(*PortConfig)(nil),         // 1: servicepb.PortConfig
	(*NextService)(nil),        // 2: servicepb.NextService
	(*NextServiceSetting)(nil), // 3: servicepb.NextServiceSetting
	(*Resources)(nil),          // 4: servicepb.Resources
	(*ResourceConfig)(nil),     // 5: servicepb.ResourceConfig
	nil,                        // 6: servicepb.Microservice.NextServiceEntry
	nil,                        // 7: servicepb.Microservice.EnvEntry
}
var file_proto_servicepb_service_proto_depIdxs = []int32{
	6, // 0: servicepb.Microservice.nextService:type_name -> servicepb.Microservice.NextServiceEntry
	7, // 1: servicepb.Microservice.env:type_name -> servicepb.Microservice.EnvEntry
	1, // 2: servicepb.Microservice.ports:type_name -> servicepb.PortConfig
	4, // 3: servicepb.Microservice.resources:type_name -> servicepb.Resources
	3, // 4: servicepb.NextService.nextServiceSetting:type_name -> servicepb.NextServiceSetting
	5, // 5: servicepb.Resources.requests:type_name -> servicepb.ResourceConfig
	5, // 6: servicepb.Resources.limits:type_name -> servicepb.ResourceConfig
	2, // 7: servicepb.Microservice.NextServiceEntry.value:type_name -> servicepb.NextService
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_proto_servicepb_service_proto_init() }
func file_proto_servicepb_service_proto_init() {
	if File_proto_servicepb_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_servicepb_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Microservice); i {
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
		file_proto_servicepb_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PortConfig); i {
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
		file_proto_servicepb_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NextService); i {
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
		file_proto_servicepb_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NextServiceSetting); i {
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
		file_proto_servicepb_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Resources); i {
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
		file_proto_servicepb_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceConfig); i {
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
			RawDescriptor: file_proto_servicepb_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_servicepb_service_proto_goTypes,
		DependencyIndexes: file_proto_servicepb_service_proto_depIdxs,
		MessageInfos:      file_proto_servicepb_service_proto_msgTypes,
	}.Build()
	File_proto_servicepb_service_proto = out.File
	file_proto_servicepb_service_proto_rawDesc = nil
	file_proto_servicepb_service_proto_goTypes = nil
	file_proto_servicepb_service_proto_depIdxs = nil
}
