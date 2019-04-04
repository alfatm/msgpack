package msgpack

import (
	"reflect"
	"sort"

	"github.com/alfatm/msgpack/codes"
)

func encodeMapValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	keys := v.MapKeys()
	if e.sortMapKeys {
		sort.Sort(genericSort(keys))
	}

	for _, key := range keys {
		if err := e.EncodeValue(key); err != nil {
			return err
		}
		if err := e.EncodeValue(v.MapIndex(key)); err != nil {
			return err
		}
	}

	return nil
}

type genericSort []reflect.Value

func (a genericSort) Len() int      { return len(a) }
func (a genericSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a genericSort) Less(i, j int) bool {
	iV := a[i]
	jV := a[j]
	iK := iV.Kind()
	jK := jV.Kind()

	switch iK {
	case reflect.String:
		if jK == reflect.String {
			return iV.String() < jV.String()
		}
		return false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch jK {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return iV.Int() < jV.Int()
		}
		return false

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch jK {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return iV.Uint() < jV.Uint()
		}
		return false

	case reflect.Float32, reflect.Float64:
		switch jK {
		case reflect.Float32, reflect.Float64:
			return iV.Float() < jV.Float()

		}
		return false
	}

	return false
}

func encodeMapStringStringValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringStringType).Interface().(map[string]string)
	if e.sortMapKeys {
		return e.encodeSortedMapStringString(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.EncodeString(mv); err != nil {
			return err
		}
	}

	return nil
}

func encodeMapStringInterfaceValue(e *Encoder, v reflect.Value) error {
	if v.IsNil() {
		return e.EncodeNil()
	}

	if err := e.EncodeMapLen(v.Len()); err != nil {
		return err
	}

	m := v.Convert(mapStringInterfaceType).Interface().(map[string]interface{})
	if e.sortMapKeys {
		return e.encodeSortedMapStringInterface(m)
	}

	for mk, mv := range m {
		if err := e.EncodeString(mk); err != nil {
			return err
		}
		if err := e.Encode(mv); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringString(m map[string]string) error {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.EncodeString(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSortedMapStringInterface(m map[string]interface{}) error {
	keys := make([]string, 0, len(m))
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		err := e.EncodeString(k)
		if err != nil {
			return err
		}
		if err = e.Encode(m[k]); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) EncodeMapLen(l int) error {
	if l < 16 {
		return e.writeCode(codes.FixedMapLow | codes.Code(l))
	}
	if l < 65536 {
		return e.write2(codes.Map16, uint16(l))
	}
	return e.write4(codes.Map32, uint32(l))
}

func encodeStructValue(e *Encoder, strct reflect.Value) error {
	var structFields *fields
	if e.useJSONTag {
		structFields = jsonStructs.Fields(strct.Type())
	} else {
		structFields = structs.Fields(strct.Type())
	}

	if e.structAsArray || structFields.AsArray {
		return encodeStructValueAsArray(e, strct, structFields.List)
	}
	fields := structFields.OmitEmpty(strct)

	if err := e.EncodeMapLen(len(fields)); err != nil {
		return err
	}

	for _, f := range fields {
		if err := e.EncodeString(f.name); err != nil {
			return err
		}
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}

	return nil
}

func encodeStructValueAsArray(e *Encoder, strct reflect.Value, fields []*field) error {
	if err := e.EncodeArrayLen(len(fields)); err != nil {
		return err
	}
	for _, f := range fields {
		if err := f.EncodeValue(e, strct); err != nil {
			return err
		}
	}
	return nil
}
