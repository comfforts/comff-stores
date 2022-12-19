// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.20.1
// source: api/v1/stores.proto

package store_v1

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

type Store struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name      string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Org       string  `protobuf:"bytes,3,opt,name=org,proto3" json:"org,omitempty"`
	Longitude float32 `protobuf:"fixed32,4,opt,name=longitude,proto3" json:"longitude,omitempty"`
	Latitude  float32 `protobuf:"fixed32,5,opt,name=latitude,proto3" json:"latitude,omitempty"`
	City      string  `protobuf:"bytes,6,opt,name=city,proto3" json:"city,omitempty"`
	Country   string  `protobuf:"bytes,7,opt,name=country,proto3" json:"country,omitempty"`
	StoreId   uint64  `protobuf:"varint,8,opt,name=storeId,proto3" json:"storeId,omitempty"`
}

func (x *Store) Reset() {
	*x = Store{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Store) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Store) ProtoMessage() {}

func (x *Store) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Store.ProtoReflect.Descriptor instead.
func (*Store) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{0}
}

func (x *Store) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Store) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Store) GetOrg() string {
	if x != nil {
		return x.Org
	}
	return ""
}

func (x *Store) GetLongitude() float32 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *Store) GetLatitude() float32 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *Store) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *Store) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *Store) GetStoreId() uint64 {
	if x != nil {
		return x.StoreId
	}
	return 0
}

type Point struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Latitude  float32 `protobuf:"fixed32,1,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude float32 `protobuf:"fixed32,2,opt,name=longitude,proto3" json:"longitude,omitempty"`
}

func (x *Point) Reset() {
	*x = Point{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Point) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Point) ProtoMessage() {}

func (x *Point) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Point.ProtoReflect.Descriptor instead.
func (*Point) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{1}
}

func (x *Point) GetLatitude() float32 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *Point) GetLongitude() float32 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

type StoreGeo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Store    *Store  `protobuf:"bytes,1,opt,name=store,proto3" json:"store,omitempty"`
	Distance float32 `protobuf:"fixed32,2,opt,name=distance,proto3" json:"distance,omitempty"`
}

func (x *StoreGeo) Reset() {
	*x = StoreGeo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreGeo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreGeo) ProtoMessage() {}

func (x *StoreGeo) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreGeo.ProtoReflect.Descriptor instead.
func (*StoreGeo) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{2}
}

func (x *StoreGeo) GetStore() *Store {
	if x != nil {
		return x.Store
	}
	return nil
}

func (x *StoreGeo) GetDistance() float32 {
	if x != nil {
		return x.Distance
	}
	return 0
}

type AddStoreRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StoreId   uint64  `protobuf:"varint,1,opt,name=storeId,proto3" json:"storeId,omitempty"`
	Name      string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Org       string  `protobuf:"bytes,3,opt,name=org,proto3" json:"org,omitempty"`
	Longitude float32 `protobuf:"fixed32,4,opt,name=longitude,proto3" json:"longitude,omitempty"`
	Latitude  float32 `protobuf:"fixed32,5,opt,name=latitude,proto3" json:"latitude,omitempty"`
	City      string  `protobuf:"bytes,6,opt,name=city,proto3" json:"city,omitempty"`
	Country   string  `protobuf:"bytes,7,opt,name=country,proto3" json:"country,omitempty"`
}

func (x *AddStoreRequest) Reset() {
	*x = AddStoreRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddStoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddStoreRequest) ProtoMessage() {}

func (x *AddStoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddStoreRequest.ProtoReflect.Descriptor instead.
func (*AddStoreRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{3}
}

func (x *AddStoreRequest) GetStoreId() uint64 {
	if x != nil {
		return x.StoreId
	}
	return 0
}

func (x *AddStoreRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *AddStoreRequest) GetOrg() string {
	if x != nil {
		return x.Org
	}
	return ""
}

func (x *AddStoreRequest) GetLongitude() float32 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *AddStoreRequest) GetLatitude() float32 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *AddStoreRequest) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *AddStoreRequest) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

type AddStoreResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok    bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Store *Store `protobuf:"bytes,2,opt,name=store,proto3,oneof" json:"store,omitempty"`
}

func (x *AddStoreResponse) Reset() {
	*x = AddStoreResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddStoreResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddStoreResponse) ProtoMessage() {}

func (x *AddStoreResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddStoreResponse.ProtoReflect.Descriptor instead.
func (*AddStoreResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{4}
}

func (x *AddStoreResponse) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *AddStoreResponse) GetStore() *Store {
	if x != nil {
		return x.Store
	}
	return nil
}

type GetStoreRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetStoreRequest) Reset() {
	*x = GetStoreRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStoreRequest) ProtoMessage() {}

func (x *GetStoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStoreRequest.ProtoReflect.Descriptor instead.
func (*GetStoreRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{5}
}

func (x *GetStoreRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetStoreResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Store *Store `protobuf:"bytes,1,opt,name=store,proto3,oneof" json:"store,omitempty"`
}

func (x *GetStoreResponse) Reset() {
	*x = GetStoreResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetStoreResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetStoreResponse) ProtoMessage() {}

func (x *GetStoreResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetStoreResponse.ProtoReflect.Descriptor instead.
func (*GetStoreResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{6}
}

func (x *GetStoreResponse) GetStore() *Store {
	if x != nil {
		return x.Store
	}
	return nil
}

type SearchStoreRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Latitude   float32 `protobuf:"fixed32,1,opt,name=latitude,proto3" json:"latitude,omitempty"`
	Longitude  float32 `protobuf:"fixed32,2,opt,name=longitude,proto3" json:"longitude,omitempty"`
	PostalCode string  `protobuf:"bytes,3,opt,name=postalCode,proto3" json:"postalCode,omitempty"`
	Distance   uint32  `protobuf:"varint,4,opt,name=distance,proto3" json:"distance,omitempty"`
}

func (x *SearchStoreRequest) Reset() {
	*x = SearchStoreRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchStoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchStoreRequest) ProtoMessage() {}

func (x *SearchStoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchStoreRequest.ProtoReflect.Descriptor instead.
func (*SearchStoreRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{7}
}

func (x *SearchStoreRequest) GetLatitude() float32 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *SearchStoreRequest) GetLongitude() float32 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *SearchStoreRequest) GetPostalCode() string {
	if x != nil {
		return x.PostalCode
	}
	return ""
}

func (x *SearchStoreRequest) GetDistance() uint32 {
	if x != nil {
		return x.Distance
	}
	return 0
}

type SearchStoreResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Stores []*StoreGeo `protobuf:"bytes,1,rep,name=stores,proto3" json:"stores,omitempty"`
	Geo    *Point      `protobuf:"bytes,2,opt,name=geo,proto3,oneof" json:"geo,omitempty"`
}

func (x *SearchStoreResponse) Reset() {
	*x = SearchStoreResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SearchStoreResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchStoreResponse) ProtoMessage() {}

func (x *SearchStoreResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchStoreResponse.ProtoReflect.Descriptor instead.
func (*SearchStoreResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{8}
}

func (x *SearchStoreResponse) GetStores() []*StoreGeo {
	if x != nil {
		return x.Stores
	}
	return nil
}

func (x *SearchStoreResponse) GetGeo() *Point {
	if x != nil {
		return x.Geo
	}
	return nil
}

type StatsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StatsRequest) Reset() {
	*x = StatsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsRequest) ProtoMessage() {}

func (x *StatsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatsRequest.ProtoReflect.Descriptor instead.
func (*StatsRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{9}
}

type StatsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count     uint32 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	HashCount uint32 `protobuf:"varint,2,opt,name=hashCount,proto3" json:"hashCount,omitempty"`
	Ready     bool   `protobuf:"varint,3,opt,name=ready,proto3" json:"ready,omitempty"`
}

func (x *StatsResponse) Reset() {
	*x = StatsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatsResponse) ProtoMessage() {}

func (x *StatsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatsResponse.ProtoReflect.Descriptor instead.
func (*StatsResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{10}
}

func (x *StatsResponse) GetCount() uint32 {
	if x != nil {
		return x.Count
	}
	return 0
}

func (x *StatsResponse) GetHashCount() uint32 {
	if x != nil {
		return x.HashCount
	}
	return 0
}

func (x *StatsResponse) GetReady() bool {
	if x != nil {
		return x.Ready
	}
	return false
}

type GeoLocationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PostalCode string `protobuf:"bytes,1,opt,name=postalCode,proto3" json:"postalCode,omitempty"`
}

func (x *GeoLocationRequest) Reset() {
	*x = GeoLocationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeoLocationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoLocationRequest) ProtoMessage() {}

func (x *GeoLocationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoLocationRequest.ProtoReflect.Descriptor instead.
func (*GeoLocationRequest) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{11}
}

func (x *GeoLocationRequest) GetPostalCode() string {
	if x != nil {
		return x.PostalCode
	}
	return ""
}

type GeoLocationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Point *Point `protobuf:"bytes,1,opt,name=point,proto3" json:"point,omitempty"`
}

func (x *GeoLocationResponse) Reset() {
	*x = GeoLocationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_v1_stores_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GeoLocationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GeoLocationResponse) ProtoMessage() {}

func (x *GeoLocationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_v1_stores_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GeoLocationResponse.ProtoReflect.Descriptor instead.
func (*GeoLocationResponse) Descriptor() ([]byte, []int) {
	return file_api_v1_stores_proto_rawDescGZIP(), []int{12}
}

func (x *GeoLocationResponse) GetPoint() *Point {
	if x != nil {
		return x.Point
	}
	return nil
}

var File_api_v1_stores_proto protoreflect.FileDescriptor

var file_api_v1_stores_proto_rawDesc = []byte{
	0x0a, 0x13, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x22,
	0xbf, 0x01, 0x0a, 0x05, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x6f, 0x72, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6f, 0x72, 0x67, 0x12,
	0x1c, 0x0a, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x02, 0x52, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52,
	0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69, 0x74,
	0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x18, 0x0a,
	0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x49, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x49,
	0x64, 0x22, 0x41, 0x0a, 0x05, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74,
	0x75, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69,
	0x74, 0x75, 0x64, 0x65, 0x22, 0x4d, 0x0a, 0x08, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x47, 0x65, 0x6f,
	0x12, 0x25, 0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x52, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x22, 0xb9, 0x01, 0x0a, 0x0f, 0x41, 0x64, 0x64, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x49,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6f, 0x72, 0x67, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6f, 0x72, 0x67, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69,
	0x74, 0x75, 0x64, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x6c, 0x6f, 0x6e, 0x67,
	0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64,
	0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69, 0x74, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x22,
	0x58, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x02, 0x6f, 0x6b, 0x12, 0x2a, 0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x6f, 0x72, 0x65, 0x48, 0x00, 0x52, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x88, 0x01, 0x01, 0x42,
	0x08, 0x0a, 0x06, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x22, 0x21, 0x0a, 0x0f, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x48, 0x0a, 0x10,
	0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2a, 0x0a, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x48, 0x00, 0x52, 0x05, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x88, 0x01, 0x01, 0x42, 0x08, 0x0a, 0x06,
	0x5f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x22, 0x8a, 0x01, 0x0a, 0x12, 0x53, 0x65, 0x61, 0x72, 0x63,
	0x68, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52,
	0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x6f, 0x6e,
	0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09, 0x6c, 0x6f,
	0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6f, 0x73, 0x74, 0x61,
	0x6c, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6f, 0x73,
	0x74, 0x61, 0x6c, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x64, 0x69, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x22, 0x71, 0x0a, 0x13, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x53, 0x74, 0x6f,
	0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x47, 0x65, 0x6f, 0x52, 0x06,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x12, 0x26, 0x0a, 0x03, 0x67, 0x65, 0x6f, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50,
	0x6f, 0x69, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x03, 0x67, 0x65, 0x6f, 0x88, 0x01, 0x01, 0x42, 0x06,
	0x0a, 0x04, 0x5f, 0x67, 0x65, 0x6f, 0x22, 0x0e, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x59, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x68, 0x61, 0x73, 0x68, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x09, 0x68, 0x61, 0x73, 0x68, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x72,
	0x65, 0x61, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x72, 0x65, 0x61, 0x64,
	0x79, 0x22, 0x34, 0x0a, 0x12, 0x47, 0x65, 0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6f, 0x73, 0x74, 0x61,
	0x6c, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6f, 0x73,
	0x74, 0x61, 0x6c, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x3c, 0x0a, 0x13, 0x47, 0x65, 0x6f, 0x4c, 0x6f,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25,
	0x0a, 0x05, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x52, 0x05,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x32, 0xeb, 0x02, 0x0a, 0x06, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x73,
	0x12, 0x43, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x19, 0x2e, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72,
	0x65, 0x12, 0x19, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4c, 0x0a, 0x0b, 0x53, 0x65,
	0x61, 0x72, 0x63, 0x68, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x1c, 0x2e, 0x73, 0x74, 0x6f, 0x72,
	0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x53, 0x74, 0x6f, 0x72, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x53,
	0x74, 0x61, 0x74, 0x73, 0x12, 0x16, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4a, 0x0a, 0x09, 0x47, 0x65, 0x6f, 0x4c, 0x6f,
	0x63, 0x61, 0x74, 0x65, 0x12, 0x1c, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x47, 0x65, 0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x6f, 0x4c, 0x6f, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0x30, 0x5a, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x63, 0x6f, 0x6d, 0x66, 0x66, 0x6f, 0x72, 0x74, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x66,
	0x66, 0x2d, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_v1_stores_proto_rawDescOnce sync.Once
	file_api_v1_stores_proto_rawDescData = file_api_v1_stores_proto_rawDesc
)

func file_api_v1_stores_proto_rawDescGZIP() []byte {
	file_api_v1_stores_proto_rawDescOnce.Do(func() {
		file_api_v1_stores_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_v1_stores_proto_rawDescData)
	})
	return file_api_v1_stores_proto_rawDescData
}

var file_api_v1_stores_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_api_v1_stores_proto_goTypes = []interface{}{
	(*Store)(nil),               // 0: store.v1.Store
	(*Point)(nil),               // 1: store.v1.Point
	(*StoreGeo)(nil),            // 2: store.v1.StoreGeo
	(*AddStoreRequest)(nil),     // 3: store.v1.AddStoreRequest
	(*AddStoreResponse)(nil),    // 4: store.v1.AddStoreResponse
	(*GetStoreRequest)(nil),     // 5: store.v1.GetStoreRequest
	(*GetStoreResponse)(nil),    // 6: store.v1.GetStoreResponse
	(*SearchStoreRequest)(nil),  // 7: store.v1.SearchStoreRequest
	(*SearchStoreResponse)(nil), // 8: store.v1.SearchStoreResponse
	(*StatsRequest)(nil),        // 9: store.v1.StatsRequest
	(*StatsResponse)(nil),       // 10: store.v1.StatsResponse
	(*GeoLocationRequest)(nil),  // 11: store.v1.GeoLocationRequest
	(*GeoLocationResponse)(nil), // 12: store.v1.GeoLocationResponse
}
var file_api_v1_stores_proto_depIdxs = []int32{
	0,  // 0: store.v1.StoreGeo.store:type_name -> store.v1.Store
	0,  // 1: store.v1.AddStoreResponse.store:type_name -> store.v1.Store
	0,  // 2: store.v1.GetStoreResponse.store:type_name -> store.v1.Store
	2,  // 3: store.v1.SearchStoreResponse.stores:type_name -> store.v1.StoreGeo
	1,  // 4: store.v1.SearchStoreResponse.geo:type_name -> store.v1.Point
	1,  // 5: store.v1.GeoLocationResponse.point:type_name -> store.v1.Point
	3,  // 6: store.v1.Stores.AddStore:input_type -> store.v1.AddStoreRequest
	5,  // 7: store.v1.Stores.GetStore:input_type -> store.v1.GetStoreRequest
	7,  // 8: store.v1.Stores.SearchStore:input_type -> store.v1.SearchStoreRequest
	9,  // 9: store.v1.Stores.GetStats:input_type -> store.v1.StatsRequest
	11, // 10: store.v1.Stores.GeoLocate:input_type -> store.v1.GeoLocationRequest
	4,  // 11: store.v1.Stores.AddStore:output_type -> store.v1.AddStoreResponse
	6,  // 12: store.v1.Stores.GetStore:output_type -> store.v1.GetStoreResponse
	8,  // 13: store.v1.Stores.SearchStore:output_type -> store.v1.SearchStoreResponse
	10, // 14: store.v1.Stores.GetStats:output_type -> store.v1.StatsResponse
	12, // 15: store.v1.Stores.GeoLocate:output_type -> store.v1.GeoLocationResponse
	11, // [11:16] is the sub-list for method output_type
	6,  // [6:11] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_api_v1_stores_proto_init() }
func file_api_v1_stores_proto_init() {
	if File_api_v1_stores_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_v1_stores_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Store); i {
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
		file_api_v1_stores_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Point); i {
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
		file_api_v1_stores_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreGeo); i {
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
		file_api_v1_stores_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddStoreRequest); i {
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
		file_api_v1_stores_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddStoreResponse); i {
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
		file_api_v1_stores_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStoreRequest); i {
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
		file_api_v1_stores_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetStoreResponse); i {
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
		file_api_v1_stores_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchStoreRequest); i {
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
		file_api_v1_stores_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SearchStoreResponse); i {
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
		file_api_v1_stores_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatsRequest); i {
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
		file_api_v1_stores_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatsResponse); i {
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
		file_api_v1_stores_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeoLocationRequest); i {
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
		file_api_v1_stores_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GeoLocationResponse); i {
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
	file_api_v1_stores_proto_msgTypes[4].OneofWrappers = []interface{}{}
	file_api_v1_stores_proto_msgTypes[6].OneofWrappers = []interface{}{}
	file_api_v1_stores_proto_msgTypes[8].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_v1_stores_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_v1_stores_proto_goTypes,
		DependencyIndexes: file_api_v1_stores_proto_depIdxs,
		MessageInfos:      file_api_v1_stores_proto_msgTypes,
	}.Build()
	File_api_v1_stores_proto = out.File
	file_api_v1_stores_proto_rawDesc = nil
	file_api_v1_stores_proto_goTypes = nil
	file_api_v1_stores_proto_depIdxs = nil
}
