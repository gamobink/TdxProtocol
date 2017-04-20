package entity

import (
	"encoding/binary"
	"math"
	"errors"
	"compress/zlib"
	"bytes"
	"io"
)

const (
	BS_BUY = 0
	BS_SELL = 1
)

type Transaction struct {
	Date uint32
	Minute uint16
	Price uint32
	Volume uint32
	Count uint32
	BS byte
}

type InfoExItem struct {
	Date uint32
	Bonus float32
	DeliveredShares float32
	RationedSharePrice float32
	RationedShares float32
}

type parser struct {
	RawBuffer []byte
	Current int
	Data []byte
}

type InstantTransParser struct {
	parser
	Req *InstantTransReq
}

func (this *parser) getCmd() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[10:12])
}

func (this *parser) getHeaderLen() int {
	return 16
}

func (this *parser) getLen() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[12:14])
}

func (this *parser) getLen1() uint16 {
	return binary.LittleEndian.Uint16(this.RawBuffer[14:16])
}

func (this *parser) getSeqId() uint32 {
	return binary.LittleEndian.Uint32(this.RawBuffer[5:9])
}

func (this *parser) skipByte(count int) {
	this.Current += count
}

func (this *parser) skipData(count int) {
	for count >= 0 {
		if this.Data[this.Current] < 0x80 {
			this.skipByte(1)
		} else if this.Data[this.Current + 1] < 0x80 {
			this.skipByte(2)
		} else {
			this.skipByte(3)
		}

		count--
	}
}

func (this *parser) getByte() byte {
	ret := this.Data[this.Current]
	this.Current++
	return ret
}

func (this *parser) getUint16() uint16 {
	ret := binary.LittleEndian.Uint16(this.Data[this.Current:this.Current + 2])
	this.Current += 2
	return ret
}

func (this *parser) getUint32() uint32 {
	ret := binary.LittleEndian.Uint32(this.Data[this.Current:this.Current + 4])
	this.Current += 4
	return ret
}

func (this *parser) getFloat32() float32 {
	bits := binary.LittleEndian.Uint32(this.Data[this.Current:this.Current + 4])
	ret := math.Float32frombits(bits)
	this.Current += 4
	return ret
}

func (this *parser) parseData() int {
	v := this.Data[this.Current]
	if v >= 0x40 && v < 0x80 || v >= 0xc0 {
		return 0x40 - this.parseData2()
	} else {
		return this.parseData2()
	}
}

func (this *parser) parseData2() int {
	 //8f ff ff ff 1f == -49
	 //bd ff ff ff 1f == -3
	 //b0 fe ff ff 1f == -80
	 //8c 01		 == 76
	 //a8 fb b6 01 == 1017 万
	 //a3 8e 11    == 14.02 万
	 //82 27         == 2498
	var v int
	var nBytes int = 0
	for this.Data[this.Current + nBytes] >= 0x80 {
		nBytes++
	}
	nBytes++

	switch(nBytes){
	case 1:
		v = int(this.Data[this.Current])
	case 2:
		v = int(this.Data[this.Current+1]) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 3:
		v = (int(this.Data[this.Current+2]) * 0x80 + int(this.Data[this.Current+1]) - 0x80) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 4:
		v = ((int(this.Data[this.Current+3]) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1] - 0x80)) * 0x40 + int(this.Data[this.Current]) - 0x80;
	case 5:
		// over flow, positive to negative
		v = (((int(this.Data[this.Current+4]) * 0x80 + int(this.Data[this.Current+3]) - 0x80) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1]) - 0x80)* 0x40 + int(this.Data[this.Current]) - 0x80;
	case 6:
		// over flow, positive to negative
		v = ((((int(this.Data[this.Current+5]) * 0x80 + int(this.Data[this.Current+4]) -0x80) * 0x80 +  int(this.Data[this.Current+3]) - 0x80) * 0x80 + int(this.Data[this.Current+2]) - 0x80) * 0x80 + int(this.Data[this.Current+1]) - 0x80) * 0x40 + int(this.Data[this.Current]) - 0x80;
	default:
		panic(errors.New("bad data"))
	}
	this.skipByte(nBytes)
	return v
}

func (this *parser) uncompressIf() {
	if this.getLen() == this.getLen1() {
		this.Data = this.RawBuffer[this.getHeaderLen():]
	} else {
		b := bytes.NewReader(this.RawBuffer[this.getHeaderLen():])
		var out bytes.Buffer
		r, _ := zlib.NewReader(b)
		io.Copy(&out, r)
		this.Data = out.Bytes()
	}

	this.Current = 0
}

func (this *InstantTransParser) Parse() []*Transaction {
	this.uncompressIf()

	var result []*Transaction

	count := this.getUint16()

	first := true
	var priceBase uint32

	for ; count > 0; count-- {
		trans := &Transaction{}
		trans.Minute = this.getUint16()
		if first {
			priceBase = uint32(this.parseData())
			trans.Price = priceBase
			first = false
		} else {
			priceBase = uint32(this.parseData()) + priceBase
			trans.Price = priceBase
		}
		trans.Volume = uint32(this.parseData())
		trans.Count = uint32(this.parseData())
		trans.BS = this.getByte()
		this.skipByte(1)
		result = append(result, trans)
	}
	return result
}

func NewInstantTransParser(req *InstantTransReq, data []byte) *InstantTransParser {
	return &InstantTransParser{
		parser: parser{
			RawBuffer: data,
		},
		Req: req,
	}
}
