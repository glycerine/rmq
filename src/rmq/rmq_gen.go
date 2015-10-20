package main

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/glycerine/msgp)
// DO NOT EDIT

import (
	"github.com/glycerine/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Payload) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Sub":
			err = z.Sub.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "D":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.D) >= int(xsz) {
				z.D = z.D[:xsz]
			} else {
				z.D = make([]string, xsz)
			}
			for xvk := range z.D {
				z.D[xvk], err = dc.ReadString()
				if err != nil {
					return
				}
			}
		case "E":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.E) >= int(xsz) {
				z.E = z.E[:xsz]
			} else {
				z.E = make([]int32, xsz)
			}
			for bzg := range z.E {
				z.E[bzg], err = dc.ReadInt32()
				if err != nil {
					return
				}
			}
		case "G":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.G) >= int(xsz) {
				z.G = z.G[:xsz]
			} else {
				z.G = make([]float64, xsz)
			}
			for bai := range z.G {
				z.G[bai], err = dc.ReadFloat64()
				if err != nil {
					return
				}
			}
		case "Blob":
			z.Blob, err = dc.ReadBytes(z.Blob)
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Payload) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "Sub"
	err = en.Append(0x85, 0xa3, 0x53, 0x75, 0x62)
	if err != nil {
		return err
	}
	err = z.Sub.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "D"
	err = en.Append(0xa1, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.D)))
	if err != nil {
		return
	}
	for xvk := range z.D {
		err = en.WriteString(z.D[xvk])
		if err != nil {
			return
		}
	}
	// write "E"
	err = en.Append(0xa1, 0x45)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.E)))
	if err != nil {
		return
	}
	for bzg := range z.E {
		err = en.WriteInt32(z.E[bzg])
		if err != nil {
			return
		}
	}
	// write "G"
	err = en.Append(0xa1, 0x47)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.G)))
	if err != nil {
		return
	}
	for bai := range z.G {
		err = en.WriteFloat64(z.G[bai])
		if err != nil {
			return
		}
	}
	// write "Blob"
	err = en.Append(0xa4, 0x42, 0x6c, 0x6f, 0x62)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.Blob)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Payload) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "Sub"
	o = append(o, 0x85, 0xa3, 0x53, 0x75, 0x62)
	o, err = z.Sub.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "D"
	o = append(o, 0xa1, 0x44)
	o = msgp.AppendArrayHeader(o, uint32(len(z.D)))
	for xvk := range z.D {
		o = msgp.AppendString(o, z.D[xvk])
	}
	// string "E"
	o = append(o, 0xa1, 0x45)
	o = msgp.AppendArrayHeader(o, uint32(len(z.E)))
	for bzg := range z.E {
		o = msgp.AppendInt32(o, z.E[bzg])
	}
	// string "G"
	o = append(o, 0xa1, 0x47)
	o = msgp.AppendArrayHeader(o, uint32(len(z.G)))
	for bai := range z.G {
		o = msgp.AppendFloat64(o, z.G[bai])
	}
	// string "Blob"
	o = append(o, 0xa4, 0x42, 0x6c, 0x6f, 0x62)
	o = msgp.AppendBytes(o, z.Blob)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Payload) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Sub":
			bts, err = z.Sub.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "D":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.D) >= int(xsz) {
				z.D = z.D[:xsz]
			} else {
				z.D = make([]string, xsz)
			}
			for xvk := range z.D {
				z.D[xvk], bts, err = msgp.ReadStringBytes(bts)
				if err != nil {
					return
				}
			}
		case "E":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.E) >= int(xsz) {
				z.E = z.E[:xsz]
			} else {
				z.E = make([]int32, xsz)
			}
			for bzg := range z.E {
				z.E[bzg], bts, err = msgp.ReadInt32Bytes(bts)
				if err != nil {
					return
				}
			}
		case "G":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.G) >= int(xsz) {
				z.G = z.G[:xsz]
			} else {
				z.G = make([]float64, xsz)
			}
			for bai := range z.G {
				z.G[bai], bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
			}
		case "Blob":
			z.Blob, bts, err = msgp.ReadBytesBytes(bts, z.Blob)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Payload) Msgsize() (s int) {
	s = 1 + 4 + z.Sub.Msgsize() + 2 + msgp.ArrayHeaderSize
	for xvk := range z.D {
		s += msgp.StringPrefixSize + len(z.D[xvk])
	}
	s += 2 + msgp.ArrayHeaderSize + (len(z.E) * (msgp.Int32Size)) + 2 + msgp.ArrayHeaderSize + (len(z.G) * (msgp.Float64Size)) + 5 + msgp.BytesPrefixSize + len(z.Blob)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Subload) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "A":
			z.A, err = dc.ReadString()
			if err != nil {
				return
			}
		case "B":
			z.B, err = dc.ReadInt()
			if err != nil {
				return
			}
		case "F":
			var xsz uint32
			xsz, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.F) >= int(xsz) {
				z.F = z.F[:xsz]
			} else {
				z.F = make([]float64, xsz)
			}
			for cmr := range z.F {
				z.F[cmr], err = dc.ReadFloat64()
				if err != nil {
					return
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Subload) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "A"
	err = en.Append(0x83, 0xa1, 0x41)
	if err != nil {
		return err
	}
	err = en.WriteString(z.A)
	if err != nil {
		return
	}
	// write "B"
	err = en.Append(0xa1, 0x42)
	if err != nil {
		return err
	}
	err = en.WriteInt(z.B)
	if err != nil {
		return
	}
	// write "F"
	err = en.Append(0xa1, 0x46)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.F)))
	if err != nil {
		return
	}
	for cmr := range z.F {
		err = en.WriteFloat64(z.F[cmr])
		if err != nil {
			return
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Subload) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "A"
	o = append(o, 0x83, 0xa1, 0x41)
	o = msgp.AppendString(o, z.A)
	// string "B"
	o = append(o, 0xa1, 0x42)
	o = msgp.AppendInt(o, z.B)
	// string "F"
	o = append(o, 0xa1, 0x46)
	o = msgp.AppendArrayHeader(o, uint32(len(z.F)))
	for cmr := range z.F {
		o = msgp.AppendFloat64(o, z.F[cmr])
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Subload) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var isz uint32
	isz, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for isz > 0 {
		isz--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "A":
			z.A, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "B":
			z.B, bts, err = msgp.ReadIntBytes(bts)
			if err != nil {
				return
			}
		case "F":
			var xsz uint32
			xsz, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.F) >= int(xsz) {
				z.F = z.F[:xsz]
			} else {
				z.F = make([]float64, xsz)
			}
			for cmr := range z.F {
				z.F[cmr], bts, err = msgp.ReadFloat64Bytes(bts)
				if err != nil {
					return
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Subload) Msgsize() (s int) {
	s = 1 + 2 + msgp.StringPrefixSize + len(z.A) + 2 + msgp.IntSize + 2 + msgp.ArrayHeaderSize + (len(z.F) * (msgp.Float64Size))
	return
}
