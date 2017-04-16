package entity

import "bytes"
import (
	"encoding/binary"
	"fmt"
)

const (
	CMD_INFO_EX = 0x000f
	CMD_STOCK_LIST = 0x0524
	CMD_INSTANT_TRANS = 0x0fc5
	CMD_HIS_TRANS = 0x0fb5
	CMD_HEART_BEAT = 0x0523
)

const (
	BLOCK_SH_A = 0
	BLOCK_SH_B = 1
	BLOCK_SZ_A = 2
	BLOCK_SZ_B = 3
	BLOCK_INDEX = 11
)

type Header struct {
	Zip 	byte
	SeqId 	uint32
	PacketType byte
	Len 	uint16
	Len1 	uint16
	Cmd 	uint16
}

type StockDef struct {
	MarketLocation 	byte
	StockCode string
}

type InfoExReq struct {
	Header Header
	Count uint16
	Stocks []*StockDef
}

type StockListReq struct {
	Header Header
	Block uint16
	Unknown1 uint16
	Offset uint16
	Count uint16
	Unknown2 uint16
}

func MarketLocationFromCode(stockCode string) byte {
	data := []byte(stockCode)
	fmt.Println(data[0])
	if data[0] <= 0x34 {
		return 0
	}
	return 1
}

func (this *Header) Write(writer *bytes.Buffer) {
	binary.Write(writer, binary.LittleEndian, *this)
}

func (this *Header) SetLength(length uint16) {
	this.Len = length
	this.Len1 = length
}

func (this *StockDef) Write(writer *bytes.Buffer) {
	writer.Write([]byte{this.MarketLocation})
	writer.Write([]byte(this.StockCode))
}

func (this *InfoExReq) Write(writer *bytes.Buffer) {
	this.Header.Write(writer)
	var int16buf2 [2]byte
	binary.LittleEndian.PutUint16(int16buf2[:], this.Count)
	writer.Write(int16buf2[:])

	for _, o := range this.Stocks {
		o.Write(writer)
	}
}

func (this *InfoExReq) Size() int {
	return 4 + 7 * len(this.Stocks)
}

func (this *InfoExReq) AddCode(stockCode string) {
	v := &StockDef{
		MarketLocationFromCode(stockCode),
		stockCode,
	}

	this.Stocks = append(this.Stocks, v)
	this.Count = uint16(len(this.Stocks))
	this.Header.SetLength(uint16(this.Size()))
}

func NewInfoExReq(seqId uint32) *InfoExReq {
	req := &InfoExReq{
		Header{
			Zip: 0xc,
			SeqId: seqId,
			PacketType: 0x1,
			Len: 0,
			Len1: 0,
			Cmd: CMD_INFO_EX,
		},
		0,
		[]*StockDef {},
	}
	return req
}

func writeUInt16(writer *bytes.Buffer, v uint16) {
	var int16buf2 [2]byte
	binary.LittleEndian.PutUint16(int16buf2[:], v)
	writer.Write(int16buf2[:])
}

func (this *StockListReq) Write(writer *bytes.Buffer) {
	this.Header.Write(writer)
	writeUInt16(writer, this.Block)
	writeUInt16(writer, this.Unknown1)
	writeUInt16(writer, this.Offset)
	writeUInt16(writer, this.Count)
	writeUInt16(writer, this.Unknown2)
}

func (this *StockListReq) Size() int {
	return 12
}

func NewStockListReq(seqId uint32, block uint16, offset uint16, count uint16) *StockListReq {
	req := &StockListReq{
		Header{
			Zip: 0xc,
			SeqId: seqId,
			PacketType: 0x1,
			Len: 0,
			Len1: 0,
			Cmd: CMD_STOCK_LIST,
		},
		block,
		0,
		offset,
		count,
		0,
	}

	req.Header.Len = uint16(req.Size())
	req.Header.Len1 = req.Header.Len

	return req
}
