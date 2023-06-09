package serialisers

import (
	"fmt"
//	"strconv"
	"math"
)

type Serialiser interface {
	IoI(*int32)
	IoI64(*int64)
	IoS(*string)
	IoF(*float32)
	IoF64(*float64)
}

type Serialisable interface {
	Serialise(Serialiser)
}

type Sizer struct {
	Size *uint64
}

func (re *Sizer) IoI(i *int32) {
	*(re.Size) += 4
}

func (re *Sizer) IoI64(i *int64) {
	*(re.Size) += 8
}

func (re *Sizer) IoS(s *string) {
	*(re.Size) += uint64(len(*s))
	*(re.Size) += 4
	fmt.Sprintln("%i", re.Size)
}

func (re *Sizer) IoF(f *float32) {
	*(re.Size) += 4
}

func (re *Sizer) IoF64(f *float64) {
	*(re.Size) += 8
}

type Saver struct {
	Array *[]byte
	index uint64
}

func (re *Saver) IoI(i *int32) {
	a := byte(*i >> 24)
	b := byte(*i >> 16)
	c := byte(*i >> 8)
	d := byte(*i >> 0)
	(*(re.Array))[re.index + 0] = byte(a)
	(*(re.Array))[re.index + 1] = byte(b)
	(*(re.Array))[re.index + 2] = byte(c)
	(*(re.Array))[re.index + 3] = byte(d)
	re.index += 4
}

func (re *Saver) IoI64(i *int64) {
	a := byte(*i >> 56)
	b := byte(*i >> 48)
	c := byte(*i >> 40)
	d := byte(*i >> 32)
	e := byte(*i >> 24)
	f := byte(*i >> 16)
	g := byte(*i >> 8)
	h := byte(*i >> 0)
	(*(re.Array))[re.index + 0] = byte(a)
	(*(re.Array))[re.index + 1] = byte(b)
	(*(re.Array))[re.index + 2] = byte(c)
	(*(re.Array))[re.index + 3] = byte(d)
	(*(re.Array))[re.index + 4] = byte(e)
	(*(re.Array))[re.index + 5] = byte(f)
	(*(re.Array))[re.index + 6] = byte(g)
	(*(re.Array))[re.index + 7] = byte(h)
	re.index += 8
}

func (re *Saver) IoS(s *string) {
	var val int32 = int32(len(*s))
	re.IoI(&val)
	for _, value := range *s {
		(*(re.Array))[re.index] = byte(value)
		re.index++
 	}
}

func (re *Saver) IoF(f *float32) {
	var i int32 = int32(math.Float32bits(*f))
	re.IoI(&i)
}

func (re *Saver) IoF64(f *float64) {
	var i int64 = int64(math.Float64bits(*f))
	re.IoI64(&i)
}

type Loader struct {
	Array *[]byte
	index uint64
}

func (re *Loader) IoI(i *int32) {
	var a byte = (*re.Array)[re.index + 0]
	var b byte = (*re.Array)[re.index + 1]
	var c byte = (*re.Array)[re.index + 2]
	var d byte = (*re.Array)[re.index + 3]
	// casts required because byte is not a normal integer
	*i = (int32(a) << 24) + (int32(b) << 16) + (int32(c) << 8)
	*i += int32(d)
	re.index += 4
}

func (re *Loader) IoI64(i *int64) {
	var a byte = (*re.Array)[re.index + 0]
	var b byte = (*re.Array)[re.index + 1]
	var c byte = (*re.Array)[re.index + 2]
	var d byte = (*re.Array)[re.index + 3]
	var e byte = (*re.Array)[re.index + 4]
	var f byte = (*re.Array)[re.index + 5]
	var g byte = (*re.Array)[re.index + 6]
	var h byte = (*re.Array)[re.index + 7]
	// casts required because byte is not a normal integer
	*i =  (int64(a) << 56) + (int64(b) << 48) + (int64(c) << 40)
	*i += (int64(d) << 32) + (int64(e) << 24) + (int64(f) << 16)
	*i += (int64(g) << 8) + int64(h)
	re.index += 8
}

func (re *Loader) IoS(s *string) {
	var val int32 = 0
	re.IoI(&val)
	for i := int32(0); i < val; i++ {
		*s += string(rune((*(re.Array))[re.index]))
		re.index++
	}
}

func (re *Loader) IoF(f *float32) {
	var i int32 = 0
	re.IoI(&i)
	*f = math.Float32frombits(uint32(i))
//	fmt.Printf("Float32: %f\n", *f)
}

func (re *Loader) IoF64(f *float64) {
	var i int64 = 0
	re.IoI64(&i)
	*f = math.Float64frombits(uint64(i))
//	fmt.Printf("Float64: %f\n", *f)
}

/*
type Printer struct {
}

func (re *Printer) IoI(i *int32) {
	fmt.Println(strconv.Itoa(int(*i)))
}

func (re *Printer) IoI64(i *int64) {
	fmt.Println(strconv.Itoa(int(*i)))
}

func (re *Printer) IoS(s *string) {
	fmt.Println(*s)
}

func (re *Printer) IoF(f *float32) {
	fmt.Println(*f)
}

func (re *Printer) IoF64(f *float64) {
	fmt.Println(*f)
}
*/
