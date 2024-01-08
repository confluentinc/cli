package ccloud

import (
	proto "github.com/golang/protobuf/proto"
)

type CreateExternalIdentityResponse struct {
	IdentityName         string   `protobuf:"bytes,1,opt,name=identity_name,json=identityName,proto3" json:"identity_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateExternalIdentityResponse) Reset()         { *m = CreateExternalIdentityResponse{} }
func (m *CreateExternalIdentityResponse) String() string { return proto.CompactTextString(m) }
func (*CreateExternalIdentityResponse) ProtoMessage()    {}

func (m *CreateExternalIdentityResponse) GetIdentityName() string {
	if m != nil {
		return m.IdentityName
	}
	return ""
}

type CreateExternalIdentityRequest struct {
	AccountId            string   `protobuf:"bytes,1,opt,name=account_id,json=accountId,proto3" json:"account_id,omitempty"`
	Cloud                string   `protobuf:"bytes,2,opt,name=cloud,proto3" json:"cloud,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateExternalIdentityRequest) Reset()         { *m = CreateExternalIdentityRequest{} }
func (m *CreateExternalIdentityRequest) String() string { return proto.CompactTextString(m) }
func (*CreateExternalIdentityRequest) ProtoMessage()    {}

func (m *CreateExternalIdentityRequest) GetAccountId() string {
	if m != nil {
		return m.AccountId
	}
	return ""
}

func (m *CreateExternalIdentityRequest) GetCloud() string {
	if m != nil {
		return m.Cloud
	}
	return ""
}
