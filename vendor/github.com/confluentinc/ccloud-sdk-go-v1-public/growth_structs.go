package ccloud

import (
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type GrowthPromoCodeClaim struct {
	Amount               int64            `protobuf:"varint,1,opt,name=amount,proto3" json:"amount,omitempty" db:"amount,omitempty" url:"amount,omitempty"`
	Balance              int64            `protobuf:"varint,2,opt,name=balance,proto3" json:"balance,omitempty" db:"balance,omitempty" url:"balance,omitempty"`
	ClaimDate            *types.Timestamp `protobuf:"bytes,4,opt,name=claim_date,json=claimDate,proto3" json:"claim_date,omitempty" db:"claim_date,omitempty" url:"claim_date,omitempty"`
	CreditExpirationDate *types.Timestamp `protobuf:"bytes,5,opt,name=credit_expiration_date,json=creditExpirationDate,proto3" json:"credit_expiration_date,omitempty" db:"credit_expiration_date,omitempty" url:"credit_expiration_date,omitempty"`
	IsFreeTrialPromoCode bool             `protobuf:"varint,6,opt,name=is_free_trial_promo_code,json=isFreeTrialPromoCode,proto3" json:"is_free_trial_promo_code,omitempty" db:"is_free_trial_promo_code,omitempty" url:"is_free_trial_promo_code,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *GrowthPromoCodeClaim) Reset()         { *m = GrowthPromoCodeClaim{} }
func (m *GrowthPromoCodeClaim) String() string { return proto.CompactTextString(m) }
func (*GrowthPromoCodeClaim) ProtoMessage()    {}

func (m *GrowthPromoCodeClaim) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *GrowthPromoCodeClaim) GetBalance() int64 {
	if m != nil {
		return m.Balance
	}
	return 0
}

func (m *GrowthPromoCodeClaim) GetClaimDate() *types.Timestamp {
	if m != nil {
		return m.ClaimDate
	}
	return nil
}

func (m *GrowthPromoCodeClaim) GetCreditExpirationDate() *types.Timestamp {
	if m != nil {
		return m.CreditExpirationDate
	}
	return nil
}

func (m *GrowthPromoCodeClaim) GetIsFreeTrialPromoCode() bool {
	if m != nil {
		return m.IsFreeTrialPromoCode
	}
	return false
}

type GetFreeTrialInfoRequest struct {
	RequestCarrier       map[string]string `protobuf:"bytes,1,rep,name=request_carrier,json=requestCarrier,proto3" json:"request_carrier,omitempty" db:"request_carrier,omitempty" url:"request_carrier,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	OrgId                int32             `protobuf:"varint,2,opt,name=org_id,json=orgId,proto3" json:"org_id,omitempty" db:"org_id,omitempty" url:"org_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *GetFreeTrialInfoRequest) Reset()         { *m = GetFreeTrialInfoRequest{} }
func (m *GetFreeTrialInfoRequest) String() string { return proto.CompactTextString(m) }
func (*GetFreeTrialInfoRequest) ProtoMessage()    {}

type GetFreeTrialInfoReply struct {
	Error                *Error                  `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty" db:"error,omitempty" url:"error,omitempty"`
	PromoCodeClaims      []*GrowthPromoCodeClaim `protobuf:"bytes,2,rep,name=promo_code_claims,json=promoCodeClaims,proto3" json:"promo_code_claims,omitempty" db:"promo_code_claims,omitempty" url:"promo_code_claims,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *GetFreeTrialInfoReply) Reset()         { *m = GetFreeTrialInfoReply{} }
func (m *GetFreeTrialInfoReply) String() string { return proto.CompactTextString(m) }
func (*GetFreeTrialInfoReply) ProtoMessage()    {}

func (m *GetFreeTrialInfoReply) GetError() *Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *GrowthPromoCodeClaim) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GrowthPromoCodeClaim) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Amount != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.Amount))
	}
	if m.Balance != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.Balance))
	}
	if m.ClaimDate != nil {
		dAtA[i] = 0x22
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.ClaimDate.Size()))
		n20, err := m.ClaimDate.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n20
	}
	if m.CreditExpirationDate != nil {
		dAtA[i] = 0x2a
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.CreditExpirationDate.Size()))
		n21, err := m.CreditExpirationDate.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n21
	}
	if m.IsFreeTrialPromoCode {
		dAtA[i] = 0x30
		i++
		if m.IsFreeTrialPromoCode {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *GetFreeTrialInfoRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetFreeTrialInfoRequest) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.RequestCarrier) > 0 {
		for k, _ := range m.RequestCarrier {
			dAtA[i] = 0xa
			i++
			v := m.RequestCarrier[k]
			mapSize := 1 + len(k) + sovGrowth(uint64(len(k))) + 1 + len(v) + sovGrowth(uint64(len(v)))
			i = encodeVarintGrowth(dAtA, i, uint64(mapSize))
			dAtA[i] = 0xa
			i++
			i = encodeVarintGrowth(dAtA, i, uint64(len(k)))
			i += copy(dAtA[i:], k)
			dAtA[i] = 0x12
			i++
			i = encodeVarintGrowth(dAtA, i, uint64(len(v)))
			i += copy(dAtA[i:], v)
		}
	}
	if m.OrgId != 0 {
		dAtA[i] = 0x10
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.OrgId))
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *GetFreeTrialInfoReply) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GetFreeTrialInfoReply) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Error != nil {
		dAtA[i] = 0xa
		i++
		i = encodeVarintGrowth(dAtA, i, uint64(m.Error.Size()))
		n22, err := m.Error.MarshalTo(dAtA[i:])
		if err != nil {
			return 0, err
		}
		i += n22
	}
	if len(m.PromoCodeClaims) > 0 {
		for _, msg := range m.PromoCodeClaims {
			dAtA[i] = 0x12
			i++
			i = encodeVarintGrowth(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintGrowth(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *GrowthPromoCodeClaim) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Amount != 0 {
		n += 1 + sovGrowth(uint64(m.Amount))
	}
	if m.Balance != 0 {
		n += 1 + sovGrowth(uint64(m.Balance))
	}
	if m.ClaimDate != nil {
		l = m.ClaimDate.Size()
		n += 1 + l + sovGrowth(uint64(l))
	}
	if m.CreditExpirationDate != nil {
		l = m.CreditExpirationDate.Size()
		n += 1 + l + sovGrowth(uint64(l))
	}
	if m.IsFreeTrialPromoCode {
		n += 2
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *GetFreeTrialInfoRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.RequestCarrier) > 0 {
		for k, v := range m.RequestCarrier {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovGrowth(uint64(len(k))) + 1 + len(v) + sovGrowth(uint64(len(v)))
			n += mapEntrySize + 1 + sovGrowth(uint64(mapEntrySize))
		}
	}
	if m.OrgId != 0 {
		n += 1 + sovGrowth(uint64(m.OrgId))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *GetFreeTrialInfoReply) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Error != nil {
		l = m.Error.Size()
		n += 1 + l + sovGrowth(uint64(l))
	}
	if len(m.PromoCodeClaims) > 0 {
		for _, e := range m.PromoCodeClaims {
			l = e.Size()
			n += 1 + l + sovGrowth(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovGrowth(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func (m *GrowthPromoCodeClaim) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGrowth
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PromoCodeClaim: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PromoCodeClaim: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			m.Amount = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Amount |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Balance", wireType)
			}
			m.Balance = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Balance |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClaimDate", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGrowth
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGrowth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ClaimDate == nil {
				m.ClaimDate = &types.Timestamp{}
			}
			if err := m.ClaimDate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreditExpirationDate", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGrowth
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGrowth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CreditExpirationDate == nil {
				m.CreditExpirationDate = &types.Timestamp{}
			}
			if err := m.CreditExpirationDate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IsFreeTrialPromoCode", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.IsFreeTrialPromoCode = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipGrowth(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GetFreeTrialInfoRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGrowth
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetFreeTrialInfoRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetFreeTrialInfoRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequestCarrier", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGrowth
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGrowth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.RequestCarrier == nil {
				m.RequestCarrier = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGrowth
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGrowth
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthGrowth
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthGrowth
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGrowth
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthGrowth
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthGrowth
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipGrowth(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if skippy < 0 {
						return ErrInvalidLengthGrowth
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.RequestCarrier[mapkey] = mapvalue
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field OrgId", wireType)
			}
			m.OrgId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.OrgId |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipGrowth(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *GetFreeTrialInfoReply) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGrowth
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GetFreeTrialInfoReply: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GetFreeTrialInfoReply: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Error", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGrowth
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGrowth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Error == nil {
				m.Error = &Error{}
			}
			if err := m.Error.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PromoCodeClaims", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGrowth
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGrowth
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PromoCodeClaims = append(m.PromoCodeClaims, &GrowthPromoCodeClaim{})
			if err := m.PromoCodeClaims[len(m.PromoCodeClaims)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGrowth(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthGrowth
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGrowth(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGrowth
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGrowth
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGrowth
			}
			iNdEx += length
			if iNdEx < 0 {
				return 0, ErrInvalidLengthGrowth
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowGrowth
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipGrowth(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
				if iNdEx < 0 {
					return 0, ErrInvalidLengthGrowth
				}
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthGrowth = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGrowth   = fmt.Errorf("proto: integer overflow")
)
