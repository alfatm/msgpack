package msgpack_test

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/alfatm/msgpack/v4"
)

type nameStruct struct {
	Name string
}

func TestGocheck(t *testing.T) { TestingT(t) }

type MsgpackTest struct {
	buf *bytes.Buffer
	enc *msgpack.Encoder
	dec *msgpack.Decoder
}

var _ = Suite(&MsgpackTest{})

func (t *MsgpackTest) SetUpTest(c *C) {
	t.buf = &bytes.Buffer{}
	t.enc = msgpack.NewEncoder(t.buf)
	t.dec = msgpack.NewDecoder(bufio.NewReader(t.buf))
}

func (t *MsgpackTest) TestDecodeNil(c *C) {
	c.Assert(t.dec.Decode(nil), NotNil)
}

func (t *MsgpackTest) TestTime(c *C) {
	in := time.Now()
	var out time.Time
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Equal(in), Equals, true)

	var zero time.Time
	c.Assert(t.enc.Encode(zero), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Equal(zero), Equals, true)
	c.Assert(out.IsZero(), Equals, true)
}

func (t *MsgpackTest) TestLargeBytes(c *C) {
	N := int(1e6)

	src := bytes.Repeat([]byte{'1'}, N)
	c.Assert(t.enc.Encode(src), IsNil)
	var dst []byte
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, DeepEquals, src)
}

func (t *MsgpackTest) TestLargeString(c *C) {
	N := int(1e6)

	src := string(bytes.Repeat([]byte{'1'}, N))
	c.Assert(t.enc.Encode(src), IsNil)
	var dst string
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, Equals, src)
}

func (t *MsgpackTest) TestSliceOfStructs(c *C) {
	in := []*nameStruct{&nameStruct{"hello"}}
	var out []*nameStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out, DeepEquals, in)
}

func (t *MsgpackTest) TestMap(c *C) {
	for _, i := range []struct {
		m map[string]string
		b []byte
	}{
		{map[string]string{}, []byte{0x80}},
		{map[string]string{"hello": "world"}, []byte{0x81, 0xa5, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64}},
	} {
		c.Assert(t.enc.Encode(i.m), IsNil)
		c.Assert(t.buf.Bytes(), DeepEquals, i.b, Commentf("err encoding %v", i.m))
		var m map[string]string
		c.Assert(t.dec.Decode(&m), IsNil)
		c.Assert(m, DeepEquals, i.m)
	}
}

func (t *MsgpackTest) TestStructNil(c *C) {
	var dst *nameStruct

	c.Assert(t.enc.Encode(nameStruct{Name: "foo"}), IsNil)
	c.Assert(t.dec.Decode(&dst), IsNil)
	c.Assert(dst, Not(IsNil))
	c.Assert(dst.Name, Equals, "foo")
}

func (t *MsgpackTest) TestStructUnknownField(c *C) {
	in := struct {
		Field1 string
		Field2 string
		Field3 string
	}{
		Field1: "value1",
		Field2: "value2",
		Field3: "value3",
	}
	c.Assert(t.enc.Encode(in), IsNil)

	out := struct {
		Field2 string
	}{}
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Field2, Equals, "value2")
}

//------------------------------------------------------------------------------

type coderStruct struct {
	name string
}

type wrapperStruct struct {
	coderStruct
}

var (
	_ msgpack.CustomEncoder = (*coderStruct)(nil)
	_ msgpack.CustomDecoder = (*coderStruct)(nil)
)

func (s *coderStruct) Name() string {
	return s.name
}

func (s *coderStruct) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.Encode(s.name)
}

func (s *coderStruct) DecodeMsgpack(dec *msgpack.Decoder) error {
	return dec.Decode(&s.name)
}

func (t *MsgpackTest) TestCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestNilCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out *coderStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestNilCoderValue(c *C) {
	in := &coderStruct{name: "hello"}
	var out *coderStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.DecodeValue(reflect.ValueOf(&out)), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestPtrToCoder(c *C) {
	in := &coderStruct{name: "hello"}
	var out coderStruct
	out2 := &out
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out2), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

func (t *MsgpackTest) TestWrappedCoder(c *C) {
	in := &wrapperStruct{coderStruct: coderStruct{name: "hello"}}
	var out wrapperStruct
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name(), Equals, "hello")
}

//------------------------------------------------------------------------------

type struct2 struct {
	Name string
}

type struct1 struct {
	Name    string
	Struct2 struct2
}

func (t *MsgpackTest) TestNestedStructs(c *C) {
	in := &struct1{Name: "hello", Struct2: struct2{Name: "world"}}
	var out struct1
	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out.Name, Equals, in.Name)
	c.Assert(out.Struct2.Name, Equals, in.Struct2.Name)
}

type Struct4 struct {
	Name2 string
}

type Struct3 struct {
	Struct4
	Name1 string
}

func TestEmbedding(t *testing.T) {
	in := &Struct3{
		Name1: "hello",
		Struct4: Struct4{
			Name2: "world",
		},
	}
	var out Struct3

	b, err := msgpack.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	err = msgpack.Unmarshal(b, &out)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name1 != in.Name1 {
		t.Fatalf("")
	}
	if out.Name2 != in.Name2 {
		t.Fatalf("")
	}
}

func (t *MsgpackTest) TestSliceNil(c *C) {
	in := [][]*int{nil}
	var out [][]*int

	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)
	c.Assert(out, DeepEquals, in)
}

//------------------------------------------------------------------------------

func (t *MsgpackTest) TestMapStringInterface(c *C) {
	in := map[string]interface{}{
		"foo": "bar",
		"hello": map[string]interface{}{
			"foo": "bar",
		},
	}
	var out map[string]interface{}

	c.Assert(t.enc.Encode(in), IsNil)
	c.Assert(t.dec.Decode(&out), IsNil)

	c.Assert(out["foo"], Equals, "bar")
	mm := out["hello"].(map[string]interface{})
	c.Assert(mm["foo"], Equals, "bar")
}

//------------------------------------------------------------------------------

func TestDisallowUnknownFields(t *testing.T) {
	type Base struct {
		FooOne   string
		FooTwo   string
		FooThree string
	}

	type Derived struct {
		FooOne string
		// FooTwo string // field missed
		FooThree string // exist
	}

	base := &Base{
		FooOne:   "barOne",
		FooTwo:   "barTwo",
		FooThree: "barThree",
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err := enc.Encode(base)
	if err != nil {
		t.Fatal(err)
	}

	derived := &Derived{}
	dec := msgpack.NewDecoder(bytes.NewReader(buf.Bytes()))
	dec = dec.DisallowUnknownFields(true)
	err = dec.Decode(&derived)
	if err == nil {
		t.Fatalf("expecting error")
	}

	if !strings.HasPrefix(err.Error(), "msgpack: unknown field") {
		t.Fatalf("unexpected error")
	}

	t.Logf("correct error %q", err)
}

//------------------------------------------------------------------------------

func TestSortMapKeys(t *testing.T) {
	type Foo struct {
		Bar string
	}

	// map int by pointer
	{
		mapIntPFoo := map[int]*Foo{
			5:      &Foo{Bar: "five"},
			3:      &Foo{Bar: "three"},
			100500: &Foo{Bar: "a lot"},
			42:     &Foo{Bar: "meaning of live"},
		}

		var buf bytes.Buffer
		enc := msgpack.NewEncoder(&buf)
		enc = enc.SortMapKeys(true)
		err := enc.Encode(&mapIntPFoo)
		if err != nil {
			t.Fatal(err)
		}

		want := "7e252d542e3f72b82f9ee84dfd150f90"
		got := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
		t.Logf("map[int]*Foo got md5 value: %q\n", got)

		// must be consisten on every encoding
		if got != want {
			t.Fatalf("map[int]*Foo inconsistent encoding, sample not match, %v != %v", got, want)
		}

		// decode
		var decMap map[int]Foo
		dec := msgpack.NewDecoder(bytes.NewReader(buf.Bytes()))
		err = dec.Decode(&decMap)
		if err != nil {
			t.Fatalf("decode sorted map failed")
		}

		if len(decMap) == 0 {
			t.Fatalf("unable decode sorted map")
		}
		t.Logf("decoded result %#+v", decMap)
	}

	// map int64 by value
	{
		mapInt64Foo := map[int64]Foo{
			5:      Foo{Bar: "five"},
			42:     Foo{Bar: "meaning of live"},
			100500: Foo{Bar: "a lot"},
			3:      Foo{Bar: "three"},
		}

		var buf bytes.Buffer
		enc := msgpack.NewEncoder(&buf)
		enc = enc.SortMapKeys(true)
		err := enc.Encode(&mapInt64Foo)
		if err != nil {
			t.Fatal(err)
		}

		want := "7e252d542e3f72b82f9ee84dfd150f90"
		got := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
		t.Logf("map[int64]Foo got md5 value: %q\n", got)

		// must be consisten on every encoding
		if got != want {
			t.Fatalf("map[int64]Foo inconsistent encoding, sample not match, %v != %v", got, want)
		}
	}

	// mixed keys
	{
		mapII := map[interface{}]interface{}{
			42:    "meaning of live",
			"foo": "bar",
			3.14:  "pi",
		}

		var buf bytes.Buffer
		enc := msgpack.NewEncoder(&buf)
		enc = enc.SortMapKeys(true)
		err := enc.Encode(&mapII)
		if err != nil {
			t.Fatal(err)
		}

		want := "8aeffbc24293bb09cd23901ae3d57ba7"
		got := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
		t.Logf("map[interface{}]interface{} got md5 value: %q\n", got)

		// must be consisten on every encoding
		if got != want {
			t.Fatalf("map[interface{}]interface{} inconsistent encoding, sample not match, %v != %v", got, want)
		}
	}
}
