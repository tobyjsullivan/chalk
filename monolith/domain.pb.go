// Code generated by protoc-gen-go. DO NOT EDIT.
// source: domain.proto

package monolith

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Error struct {
	Message              string   `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Error) Reset()         { *m = Error{} }
func (m *Error) String() string { return proto.CompactTextString(m) }
func (*Error) ProtoMessage()    {}
func (*Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_73e6234e76dbdb84, []int{0}
}

func (m *Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Error.Unmarshal(m, b)
}
func (m *Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Error.Marshal(b, m, deterministic)
}
func (m *Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Error.Merge(m, src)
}
func (m *Error) XXX_Size() int {
	return xxx_messageInfo_Error.Size(m)
}
func (m *Error) XXX_DiscardUnknown() {
	xxx_messageInfo_Error.DiscardUnknown(m)
}

var xxx_messageInfo_Error proto.InternalMessageInfo

func (m *Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type Session struct {
	SessionId            string   `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Session) Reset()         { *m = Session{} }
func (m *Session) String() string { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()    {}
func (*Session) Descriptor() ([]byte, []int) {
	return fileDescriptor_73e6234e76dbdb84, []int{1}
}

func (m *Session) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Session.Unmarshal(m, b)
}
func (m *Session) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Session.Marshal(b, m, deterministic)
}
func (m *Session) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Session.Merge(m, src)
}
func (m *Session) XXX_Size() int {
	return xxx_messageInfo_Session.Size(m)
}
func (m *Session) XXX_DiscardUnknown() {
	xxx_messageInfo_Session.DiscardUnknown(m)
}

var xxx_messageInfo_Session proto.InternalMessageInfo

func (m *Session) GetSessionId() string {
	if m != nil {
		return m.SessionId
	}
	return ""
}

type Page struct {
	PageId               string   `protobuf:"bytes,1,opt,name=page_id,json=pageId,proto3" json:"page_id,omitempty"`
	Session              string   `protobuf:"bytes,2,opt,name=session,proto3" json:"session,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Page) Reset()         { *m = Page{} }
func (m *Page) String() string { return proto.CompactTextString(m) }
func (*Page) ProtoMessage()    {}
func (*Page) Descriptor() ([]byte, []int) {
	return fileDescriptor_73e6234e76dbdb84, []int{2}
}

func (m *Page) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Page.Unmarshal(m, b)
}
func (m *Page) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Page.Marshal(b, m, deterministic)
}
func (m *Page) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Page.Merge(m, src)
}
func (m *Page) XXX_Size() int {
	return xxx_messageInfo_Page.Size(m)
}
func (m *Page) XXX_DiscardUnknown() {
	xxx_messageInfo_Page.DiscardUnknown(m)
}

var xxx_messageInfo_Page proto.InternalMessageInfo

func (m *Page) GetPageId() string {
	if m != nil {
		return m.PageId
	}
	return ""
}

func (m *Page) GetSession() string {
	if m != nil {
		return m.Session
	}
	return ""
}

type Variable struct {
	VariableId           string   `protobuf:"bytes,1,opt,name=variable_id,json=variableId,proto3" json:"variable_id,omitempty"`
	Page                 string   `protobuf:"bytes,2,opt,name=page,proto3" json:"page,omitempty"`
	Name                 string   `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Formula              string   `protobuf:"bytes,4,opt,name=formula,proto3" json:"formula,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Variable) Reset()         { *m = Variable{} }
func (m *Variable) String() string { return proto.CompactTextString(m) }
func (*Variable) ProtoMessage()    {}
func (*Variable) Descriptor() ([]byte, []int) {
	return fileDescriptor_73e6234e76dbdb84, []int{3}
}

func (m *Variable) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Variable.Unmarshal(m, b)
}
func (m *Variable) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Variable.Marshal(b, m, deterministic)
}
func (m *Variable) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Variable.Merge(m, src)
}
func (m *Variable) XXX_Size() int {
	return xxx_messageInfo_Variable.Size(m)
}
func (m *Variable) XXX_DiscardUnknown() {
	xxx_messageInfo_Variable.DiscardUnknown(m)
}

var xxx_messageInfo_Variable proto.InternalMessageInfo

func (m *Variable) GetVariableId() string {
	if m != nil {
		return m.VariableId
	}
	return ""
}

func (m *Variable) GetPage() string {
	if m != nil {
		return m.Page
	}
	return ""
}

func (m *Variable) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Variable) GetFormula() string {
	if m != nil {
		return m.Formula
	}
	return ""
}

func init() {
	proto.RegisterType((*Error)(nil), "monolith.Error")
	proto.RegisterType((*Session)(nil), "monolith.Session")
	proto.RegisterType((*Page)(nil), "monolith.Page")
	proto.RegisterType((*Variable)(nil), "monolith.Variable")
}

func init() { proto.RegisterFile("domain.proto", fileDescriptor_73e6234e76dbdb84) }

var fileDescriptor_73e6234e76dbdb84 = []byte{
	// 199 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x90, 0x3f, 0x6f, 0xc2, 0x30,
	0x10, 0xc5, 0x95, 0x36, 0xcd, 0x9f, 0x6b, 0x27, 0x2f, 0xf5, 0x52, 0x15, 0x3c, 0x65, 0x62, 0x61,
	0xe2, 0x03, 0x30, 0x64, 0x43, 0x20, 0xb1, 0xa2, 0x8b, 0x6c, 0x82, 0xa5, 0xd8, 0x17, 0xd9, 0x81,
	0xcf, 0x8f, 0x1c, 0x3b, 0xca, 0xf6, 0xde, 0x4f, 0xa7, 0xf7, 0x93, 0x0e, 0x7e, 0x24, 0x19, 0xd4,
	0x76, 0x37, 0x3a, 0x9a, 0x88, 0x55, 0x86, 0x2c, 0x0d, 0x7a, 0x7a, 0x88, 0x2d, 0x7c, 0x1d, 0x9d,
	0x23, 0xc7, 0x38, 0x94, 0x46, 0x79, 0x8f, 0xbd, 0xe2, 0xd9, 0x26, 0x6b, 0xea, 0xf3, 0x52, 0x45,
	0x03, 0xe5, 0x45, 0x79, 0xaf, 0xc9, 0xb2, 0x3f, 0x00, 0x1f, 0xe3, 0x4d, 0xcb, 0x74, 0x57, 0x27,
	0xd2, 0x4a, 0x71, 0x80, 0xfc, 0x84, 0xbd, 0x62, 0xbf, 0x50, 0x8e, 0xd8, 0xab, 0xf5, 0xa6, 0x08,
	0xb5, 0x95, 0x41, 0x92, 0xae, 0xf9, 0x47, 0x94, 0xa4, 0x2a, 0x0c, 0x54, 0x57, 0x74, 0x1a, 0xbb,
	0x41, 0xb1, 0x7f, 0xf8, 0x7e, 0xa5, 0xbc, 0x4e, 0xc0, 0x82, 0x5a, 0xc9, 0x18, 0xe4, 0x61, 0x30,
	0x6d, 0xcc, 0x39, 0x30, 0x8b, 0x46, 0xf1, 0xcf, 0xc8, 0x42, 0x0e, 0xba, 0x3b, 0x39, 0xf3, 0x1c,
	0x90, 0xe7, 0x51, 0x97, 0x6a, 0x57, 0xcc, 0x7f, 0xd8, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0xcf,
	0x5e, 0x33, 0x14, 0x17, 0x01, 0x00, 0x00,
}
