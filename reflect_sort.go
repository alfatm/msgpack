package msgpack

import (
	"fmt"
	"reflect"
)

type reflectSort []reflect.Value

func (a reflectSort) Len() int      { return len(a) }
func (a reflectSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a reflectSort) Less(i, j int) bool {
	iV := reflect.Indirect(a[i])
	jV := reflect.Indirect(a[j])
	iK := iV.Kind()
	jK := jV.Kind()

	switch iK {
	case reflect.String:
		if jK == reflect.String {
			return iV.String() < jV.String()
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch jK {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return iV.Int() < jV.Int()
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch jK {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return iV.Uint() < jV.Uint()
		}

	case reflect.Float32, reflect.Float64:
		switch jK {
		case reflect.Float32, reflect.Float64:
			return iV.Float() < jV.Float()

		}
	}

	return fmt.Sprintf("%v", iV.Interface()) < fmt.Sprintf("%v", jV.Interface())
}
