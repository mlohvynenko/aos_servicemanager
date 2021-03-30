// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.5.0
// source: systemcore.proto

package systemcore

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type ServerCode int32

const (
	ServerCode_OK ServerCode = 0
	// The server is unable to fulfil the client request because the request is malformed.
	ServerCode_BAD_REQUEST ServerCode = 400
	// access_token: [token expired, token invalid, token missing]
	ServerCode_UNAUTHORIZED ServerCode = 401
	// service is not permitted to access the requested resource
	ServerCode_FORBIDDEN ServerCode = 403
	// The specified data path does not exist.
	ServerCode_NOT_FOUND ServerCode = 404
	// The client has sent the server too many requests in a given amount of time.
	ServerCode_TOO_MANY_REQUEST ServerCode = 429
	// The server is currently unable to handle the request due to a temporary overload or scheduled maintenance
	ServerCode_SERVICE_UNAVALIBLE ServerCode = 503
)

// Enum value maps for ServerCode.
var (
	ServerCode_name = map[int32]string{
		0:   "OK",
		400: "BAD_REQUEST",
		401: "UNAUTHORIZED",
		403: "FORBIDDEN",
		404: "NOT_FOUND",
		429: "TOO_MANY_REQUEST",
		503: "SERVICE_UNAVALIBLE",
	}
	ServerCode_value = map[string]int32{
		"OK":                 0,
		"BAD_REQUEST":        400,
		"UNAUTHORIZED":       401,
		"FORBIDDEN":          403,
		"NOT_FOUND":          404,
		"TOO_MANY_REQUEST":   429,
		"SERVICE_UNAVALIBLE": 503,
	}
)

func (x ServerCode) Enum() *ServerCode {
	p := new(ServerCode)
	*p = x
	return p
}

func (x ServerCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ServerCode) Descriptor() protoreflect.EnumDescriptor {
	return file_systemcore_proto_enumTypes[0].Descriptor()
}

func (ServerCode) Type() protoreflect.EnumType {
	return &file_systemcore_proto_enumTypes[0]
}

func (x ServerCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ServerCode.Descriptor instead.
func (ServerCode) EnumDescriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{0}
}

// Request execution status
type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// request code
	Code ServerCode `protobuf:"varint,1,opt,name=code,proto3,enum=systemcore.ServerCode" json:"code,omitempty"`
	// message containing the result of the request
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{0}
}

func (x *Status) GetCode() ServerCode {
	if x != nil {
		return x.Code
	}
	return ServerCode_OK
}

func (x *Status) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// structure containing the path to the requested resource
type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// path to resource on system, similar to VIS
	// for example, system.services.restart - for restarting service per Service Provider
	PathToResource string `protobuf:"bytes,1,opt,name=path_to_resource,json=pathToResource,proto3" json:"path_to_resource,omitempty"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{1}
}

func (x *Resource) GetPathToResource() string {
	if x != nil {
		return x.PathToResource
	}
	return ""
}

type ResourceSetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// unique id to get access token
	// this value can be taken from env variable AOS_SECRET
	// inside of container
	Secret string `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	// path to resource
	Resource *Resource `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
	// the value to be set for the resource
	Value *any.Any `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ResourceSetRequest) Reset() {
	*x = ResourceSetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceSetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceSetRequest) ProtoMessage() {}

func (x *ResourceSetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceSetRequest.ProtoReflect.Descriptor instead.
func (*ResourceSetRequest) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{2}
}

func (x *ResourceSetRequest) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *ResourceSetRequest) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

func (x *ResourceSetRequest) GetValue() *any.Any {
	if x != nil {
		return x.Value
	}
	return nil
}

type ResourceSetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status *Status `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *ResourceSetResponse) Reset() {
	*x = ResourceSetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceSetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceSetResponse) ProtoMessage() {}

func (x *ResourceSetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceSetResponse.ProtoReflect.Descriptor instead.
func (*ResourceSetResponse) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{3}
}

func (x *ResourceSetResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

type ResourceGetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// unique id to get access token
	// this value can be taken from env variable AOS_SECRET
	// inside of container
	Secret string `protobuf:"bytes,1,opt,name=secret,proto3" json:"secret,omitempty"`
	// path to resource
	Resource *Resource `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
}

func (x *ResourceGetRequest) Reset() {
	*x = ResourceGetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceGetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceGetRequest) ProtoMessage() {}

func (x *ResourceGetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceGetRequest.ProtoReflect.Descriptor instead.
func (*ResourceGetRequest) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{4}
}

func (x *ResourceGetRequest) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *ResourceGetRequest) GetResource() *Resource {
	if x != nil {
		return x.Resource
	}
	return nil
}

type ResourceGetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// the value of the current state of the resource
	Value  *any.Any `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Status *Status  `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *ResourceGetResponse) Reset() {
	*x = ResourceGetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_systemcore_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceGetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceGetResponse) ProtoMessage() {}

func (x *ResourceGetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_systemcore_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceGetResponse.ProtoReflect.Descriptor instead.
func (*ResourceGetResponse) Descriptor() ([]byte, []int) {
	return file_systemcore_proto_rawDescGZIP(), []int{5}
}

func (x *ResourceGetResponse) GetValue() *any.Any {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *ResourceGetResponse) GetStatus() *Status {
	if x != nil {
		return x.Status
	}
	return nil
}

var File_systemcore_proto protoreflect.FileDescriptor

var file_systemcore_proto_rawDesc = []byte{
	0x0a, 0x10, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0a, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x1a, 0x19,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4e, 0x0a, 0x06, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x2a, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x16, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x34, 0x0a, 0x08, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x28, 0x0a, 0x10, 0x70, 0x61, 0x74, 0x68, 0x5f, 0x74, 0x6f,
	0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0e, 0x70, 0x61, 0x74, 0x68, 0x54, 0x6f, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22,
	0x8a, 0x01, 0x0a, 0x12, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x30,
	0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x41, 0x0a, 0x13,
	0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22,
	0x5e, 0x0a, 0x12, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x12, 0x30, 0x0a,
	0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x22,
	0x6d, 0x0a, 0x13, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0x89,
	0x01, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x06, 0x0a,
	0x02, 0x4f, 0x4b, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0b, 0x42, 0x41, 0x44, 0x5f, 0x52, 0x45, 0x51,
	0x55, 0x45, 0x53, 0x54, 0x10, 0x90, 0x03, 0x12, 0x11, 0x0a, 0x0c, 0x55, 0x4e, 0x41, 0x55, 0x54,
	0x48, 0x4f, 0x52, 0x49, 0x5a, 0x45, 0x44, 0x10, 0x91, 0x03, 0x12, 0x0e, 0x0a, 0x09, 0x46, 0x4f,
	0x52, 0x42, 0x49, 0x44, 0x44, 0x45, 0x4e, 0x10, 0x93, 0x03, 0x12, 0x0e, 0x0a, 0x09, 0x4e, 0x4f,
	0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x94, 0x03, 0x12, 0x15, 0x0a, 0x10, 0x54, 0x4f,
	0x4f, 0x5f, 0x4d, 0x41, 0x4e, 0x59, 0x5f, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0xad,
	0x03, 0x12, 0x17, 0x0a, 0x12, 0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x5f, 0x55, 0x4e, 0x41,
	0x56, 0x41, 0x4c, 0x49, 0x42, 0x4c, 0x45, 0x10, 0xf7, 0x03, 0x32, 0xa0, 0x01, 0x0a, 0x0a, 0x53,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x43, 0x6f, 0x72, 0x65, 0x12, 0x48, 0x0a, 0x03, 0x53, 0x65, 0x74,
	0x12, 0x1e, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1f, 0x2e, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x53, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x48, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12, 0x1e, 0x2e, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x73, 0x79, 0x73,
	0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3a, 0x5a,
	0x38, 0x67, 0x69, 0x74, 0x70, 0x63, 0x74, 0x2e, 0x65, 0x70, 0x61, 0x6d, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x65, 0x70, 0x6d, 0x64, 0x2d, 0x61, 0x65, 0x70, 0x72, 0x2f, 0x61, 0x6f, 0x73, 0x5f, 0x73,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73,
	0x79, 0x73, 0x74, 0x65, 0x6d, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_systemcore_proto_rawDescOnce sync.Once
	file_systemcore_proto_rawDescData = file_systemcore_proto_rawDesc
)

func file_systemcore_proto_rawDescGZIP() []byte {
	file_systemcore_proto_rawDescOnce.Do(func() {
		file_systemcore_proto_rawDescData = protoimpl.X.CompressGZIP(file_systemcore_proto_rawDescData)
	})
	return file_systemcore_proto_rawDescData
}

var file_systemcore_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_systemcore_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_systemcore_proto_goTypes = []interface{}{
	(ServerCode)(0),             // 0: systemcore.ServerCode
	(*Status)(nil),              // 1: systemcore.Status
	(*Resource)(nil),            // 2: systemcore.Resource
	(*ResourceSetRequest)(nil),  // 3: systemcore.ResourceSetRequest
	(*ResourceSetResponse)(nil), // 4: systemcore.ResourceSetResponse
	(*ResourceGetRequest)(nil),  // 5: systemcore.ResourceGetRequest
	(*ResourceGetResponse)(nil), // 6: systemcore.ResourceGetResponse
	(*any.Any)(nil),             // 7: google.protobuf.Any
}
var file_systemcore_proto_depIdxs = []int32{
	0, // 0: systemcore.Status.code:type_name -> systemcore.ServerCode
	2, // 1: systemcore.ResourceSetRequest.resource:type_name -> systemcore.Resource
	7, // 2: systemcore.ResourceSetRequest.value:type_name -> google.protobuf.Any
	1, // 3: systemcore.ResourceSetResponse.status:type_name -> systemcore.Status
	2, // 4: systemcore.ResourceGetRequest.resource:type_name -> systemcore.Resource
	7, // 5: systemcore.ResourceGetResponse.value:type_name -> google.protobuf.Any
	1, // 6: systemcore.ResourceGetResponse.status:type_name -> systemcore.Status
	3, // 7: systemcore.SystemCore.Set:input_type -> systemcore.ResourceSetRequest
	5, // 8: systemcore.SystemCore.Get:input_type -> systemcore.ResourceGetRequest
	4, // 9: systemcore.SystemCore.Set:output_type -> systemcore.ResourceSetResponse
	6, // 10: systemcore.SystemCore.Get:output_type -> systemcore.ResourceGetResponse
	9, // [9:11] is the sub-list for method output_type
	7, // [7:9] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_systemcore_proto_init() }
func file_systemcore_proto_init() {
	if File_systemcore_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_systemcore_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
		file_systemcore_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Resource); i {
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
		file_systemcore_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceSetRequest); i {
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
		file_systemcore_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceSetResponse); i {
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
		file_systemcore_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceGetRequest); i {
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
		file_systemcore_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceGetResponse); i {
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
			RawDescriptor: file_systemcore_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_systemcore_proto_goTypes,
		DependencyIndexes: file_systemcore_proto_depIdxs,
		EnumInfos:         file_systemcore_proto_enumTypes,
		MessageInfos:      file_systemcore_proto_msgTypes,
	}.Build()
	File_systemcore_proto = out.File
	file_systemcore_proto_rawDesc = nil
	file_systemcore_proto_goTypes = nil
	file_systemcore_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// SystemCoreClient is the client API for SystemCore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SystemCoreClient interface {
	// API for executing an authorized request for a set specified resource (reset, stop etс.).
	// The request must contain the path to the resource and the value that needs to be set for this resource.
	Set(ctx context.Context, in *ResourceSetRequest, opts ...grpc.CallOption) (*ResourceSetResponse, error)
	// API for executing an authorized request for a get specified resource (reset, stop etс.).
	// The request must contain the path to the resource.
	Get(ctx context.Context, in *ResourceGetRequest, opts ...grpc.CallOption) (*ResourceGetResponse, error)
}

type systemCoreClient struct {
	cc grpc.ClientConnInterface
}

func NewSystemCoreClient(cc grpc.ClientConnInterface) SystemCoreClient {
	return &systemCoreClient{cc}
}

func (c *systemCoreClient) Set(ctx context.Context, in *ResourceSetRequest, opts ...grpc.CallOption) (*ResourceSetResponse, error) {
	out := new(ResourceSetResponse)
	err := c.cc.Invoke(ctx, "/systemcore.SystemCore/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *systemCoreClient) Get(ctx context.Context, in *ResourceGetRequest, opts ...grpc.CallOption) (*ResourceGetResponse, error) {
	out := new(ResourceGetResponse)
	err := c.cc.Invoke(ctx, "/systemcore.SystemCore/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SystemCoreServer is the server API for SystemCore service.
type SystemCoreServer interface {
	// API for executing an authorized request for a set specified resource (reset, stop etс.).
	// The request must contain the path to the resource and the value that needs to be set for this resource.
	Set(context.Context, *ResourceSetRequest) (*ResourceSetResponse, error)
	// API for executing an authorized request for a get specified resource (reset, stop etс.).
	// The request must contain the path to the resource.
	Get(context.Context, *ResourceGetRequest) (*ResourceGetResponse, error)
}

// UnimplementedSystemCoreServer can be embedded to have forward compatible implementations.
type UnimplementedSystemCoreServer struct {
}

func (*UnimplementedSystemCoreServer) Set(context.Context, *ResourceSetRequest) (*ResourceSetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (*UnimplementedSystemCoreServer) Get(context.Context, *ResourceGetRequest) (*ResourceGetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}

func RegisterSystemCoreServer(s *grpc.Server, srv SystemCoreServer) {
	s.RegisterService(&_SystemCore_serviceDesc, srv)
}

func _SystemCore_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemCoreServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/systemcore.SystemCore/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemCoreServer).Set(ctx, req.(*ResourceSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SystemCore_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceGetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SystemCoreServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/systemcore.SystemCore/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SystemCoreServer).Get(ctx, req.(*ResourceGetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SystemCore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "systemcore.SystemCore",
	HandlerType: (*SystemCoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _SystemCore_Set_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _SystemCore_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "systemcore.proto",
}
