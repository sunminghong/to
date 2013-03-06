/*
  Copyright (c) 2012-2013 José Carlos Nieto, http://xiam.menteslibres.org/

  Permission is hereby granted, free of charge, to any person obtaining
  a copy of this software and associated documentation files (the
  "Software"), to deal in the Software without restriction, including
  without limitation the rights to use, copy, modify, merge, publish,
  distribute, sublicense, and/or sell copies of the Software, and to
  permit persons to whom the Software is furnished to do so, subject to
  the following conditions:

  The above copyright notice and this permission notice shall be
  included in all copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
  MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
  LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
  OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
  WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

/*
	A helper package for converting between datatypes.

	If a certain datatype could not be directly converted to another, the
	zero value of the destination type would be returned instead.

	This is a experimental package, it may change anytime without warning.
*/
package to

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"time"
)

const (
	digits     = "0123456789"
	uintbuflen = 20
)

var strToTimeFormats = []string{
	"2006-01-02",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.000000000",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02T15:04:05.000000000",
	"01/02/2006",
	"01/02/2006 15:04",
	"01/02/2006 15:04:05",
	"01/02/2006 15:04:05.000000000",
	"01/02/06",
	"01/02/06 15:04",
	"01/02/06 15:04:05",
	"01/02/06 15:04:05.000000000",
	"Mon Jan _2 15:04:05.000000000 -0700 MST 2006",
	"_2/Jan/2006 15:04:05",
	"Jan _2, 2006",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC850,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	time.Kitchen,
	time.Stamp,
	time.StampMilli,
	time.StampMicro,
	time.StampNano,
}

var strToDurationMatches = map[*regexp.Regexp]func([][][]byte) (time.Duration, error){
	regexp.MustCompile(`^(\-?\d+):(\d+)$`): func(m [][][]byte) (time.Duration, error) {
		sign := 1

		hrs := time.Hour * time.Duration(Int64(m[0][1]))

		if hrs < 0 {
			hrs = -1 * hrs
			sign = -1
		}

		min := time.Minute * time.Duration(Int64(m[0][2]))

		return time.Duration(sign) * (hrs + min), nil
	},
	regexp.MustCompile(`^(\-?\d+):(\d+):(\d+)$`): func(m [][][]byte) (time.Duration, error) {
		sign := 1

		hrs := time.Hour * time.Duration(Int64(m[0][1]))

		if hrs < 0 {
			hrs = -1 * hrs
			sign = -1
		}

		min := time.Minute * time.Duration(Int64(m[0][2]))
		sec := time.Second * time.Duration(Int64(m[0][3]))

		return time.Duration(sign) * (hrs + min + sec), nil
	},
	regexp.MustCompile(`^(\-?\d+):(\d+):(\d+).(\d+)$`): func(m [][][]byte) (time.Duration, error) {
		sign := 1

		hrs := time.Hour * time.Duration(Int64(m[0][1]))

		if hrs < 0 {
			hrs = -1 * hrs
			sign = -1
		}

		min := time.Minute * time.Duration(Int64(m[0][2]))
		sec := time.Second * time.Duration(Int64(m[0][3]))
		lst := m[0][4]

		for len(lst) < 9 {
			lst = append(lst, '0')
		}
		lst = lst[0:9]

		return time.Duration(sign) * (hrs + min + sec + time.Duration(Int64(lst))), nil
	},
}

func strToDuration(v string) time.Duration {

	var err error
	var d time.Duration

	d, err = time.ParseDuration(v)

	if err == nil {
		return d
	}

	b := []byte(v)

	for re, fn := range strToDurationMatches {
		m := re.FindAllSubmatch(b, -1)
		if m != nil {
			r, err := fn(m)
			if err == nil {
				return r
			}
		}
	}

	return time.Duration(0)
}

func uint64ToBytes(v uint64) []byte {
	buf := make([]byte, uintbuflen)

	i := len(buf)

	for v >= 10 {
		i--
		buf[i] = digits[v%10]
		v = v / 10
	}

	i--
	buf[i] = digits[v%10]

	return buf[i:]
}

func int64ToBytes(v int64) []byte {
	negative := false

	if v < 0 {
		negative = true
		v = -v
	}

	uv := uint64(v)

	buf := uint64ToBytes(uv)

	if negative {
		buf2 := []byte{'-'}
		buf2 = append(buf2, buf...)
		return buf2
	}

	return buf
}

func float32ToBytes(v float32) []byte {
	slice := strconv.AppendFloat(nil, float64(v), 'g', -1, 32)
	return slice
}

func float64ToBytes(v float64) []byte {
	slice := strconv.AppendFloat(nil, v, 'g', -1, 64)
	return slice
}

func complex128ToBytes(v complex128) []byte {
	buf := []byte{'('}

	r := strconv.AppendFloat(buf, real(v), 'g', -1, 64)

	im := imag(v)
	if im >= 0 {
		buf = append(r, '+')
	} else {
		buf = r
	}

	i := strconv.AppendFloat(buf, im, 'g', -1, 64)

	buf = append(i, []byte{'i', ')'}...)

	return buf
}

func Time(val interface{}) time.Time {
	switch t := val.(type) {
	// We could use this later.
	default:
		s := String(t)
		for _, format := range strToTimeFormats {
			r, err := time.Parse(format, s)
			if err == nil {
				return r
			}
		}
	}
	return time.Time{}
}

func Duration(val interface{}) time.Duration {
	switch t := val.(type) {
	case int:
		return time.Duration(int64(t))
	case int8:
		return time.Duration(int64(t))
	case int16:
		return time.Duration(int64(t))
	case int32:
		return time.Duration(int64(t))
	case int64:
		return time.Duration(t)
	case uint:
		return time.Duration(int64(t))
	case uint8:
		return time.Duration(int64(t))
	case uint16:
		return time.Duration(int64(t))
	case uint32:
		return time.Duration(int64(t))
	case uint64:
		return time.Duration(int64(t))
	default:
		return strToDuration(String(val))
	}
	panic("Reached")
}

func Bytes(val interface{}) []byte {
	if val == nil {
		return []byte{}
	}

	switch val.(type) {

	case int:
		return int64ToBytes(int64(val.(int)))

	case int8:
		return int64ToBytes(int64(val.(int8)))
	case int16:
		return int64ToBytes(int64(val.(int16)))
	case int32:
		return int64ToBytes(int64(val.(int32)))
	case int64:
		return int64ToBytes(val.(int64))

	case uint:
		return uint64ToBytes(uint64(val.(uint)))
	case uint8:
		return uint64ToBytes(uint64(val.(uint8)))
	case uint16:
		return uint64ToBytes(uint64(val.(uint16)))
	case uint32:
		return uint64ToBytes(uint64(val.(uint32)))
	case uint64:
		return uint64ToBytes(val.(uint64))

	case float32:
		return float32ToBytes(val.(float32))
	case float64:
		return float64ToBytes(val.(float64))

	case complex128:
		return complex128ToBytes(val.(complex128))
	case complex64:
		return complex128ToBytes(complex128(val.(complex64)))

	case bool:
		if val.(bool) == true {
			return []byte("true")
		} else {
			return []byte("false")
		}

	case string:
		return []byte(val.(string))

	case []byte:
		return val.([]byte)

	default:
		return []byte(fmt.Sprintf("%v", val))
	}

	panic("Not reached.")
}

func String(val interface{}) string {
	var buf []byte

	if val == nil {
		return ""
	}

	switch val.(type) {

	case int:
		buf = int64ToBytes(int64(val.(int)))
	case int8:
		buf = int64ToBytes(int64(val.(int8)))
	case int16:
		buf = int64ToBytes(int64(val.(int16)))
	case int32:
		buf = int64ToBytes(int64(val.(int32)))
	case int64:
		buf = int64ToBytes(val.(int64))

	case uint:
		buf = uint64ToBytes(uint64(val.(uint)))
	case uint8:
		buf = uint64ToBytes(uint64(val.(uint8)))
	case uint16:
		buf = uint64ToBytes(uint64(val.(uint16)))
	case uint32:
		buf = uint64ToBytes(uint64(val.(uint32)))
	case uint64:
		buf = uint64ToBytes(val.(uint64))

	case float32:
		buf = float32ToBytes(val.(float32))
	case float64:
		buf = float64ToBytes(val.(float64))

	case complex128:
		buf = complex128ToBytes(val.(complex128))
	case complex64:
		buf = complex128ToBytes(complex128(val.(complex64)))

	case bool:
		if val.(bool) == true {
			return "true"
		} else {
			return "false"
		}

	case string:
		return val.(string)

	case []byte:
		return string(val.([]byte))

	default:
		return fmt.Sprintf("%v", val)
	}

	return string(buf)
}

func List(val interface{}) []interface{} {
	list := []interface{}{}

	if val == nil {
		return list
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Slice:
		vval := reflect.ValueOf(val)

		size := vval.Len()
		list := make([]interface{}, size)
		vlist := reflect.ValueOf(list)

		for i := 0; i < size; i++ {
			vlist.Index(i).Set(vval.Index(i))
		}

		return list
	}

	return list
}

func Map(val interface{}) map[string]interface{} {

	list := map[string]interface{}{}

	if val == nil {
		return list
	}

	switch reflect.TypeOf(val).Kind() {
	case reflect.Map:
		vval := reflect.ValueOf(val)
		vlist := reflect.ValueOf(list)

		for _, vkey := range vval.MapKeys() {
			key := String(vkey.Interface())
			vlist.SetMapIndex(reflect.ValueOf(key), vval.MapIndex(vkey))
		}

		return list
	}

	return list
}

func Int(val interface{}) int {
	return int(Int64(val))
}

func Int8(val interface{}) int8 {
	return int8(Int64(val))
}

func Int16(val interface{}) int16 {
	return int16(Int64(val))
}

func Int32(val interface{}) int32 {
	return int32(Int64(val))
}

func Int64(val interface{}) int64 {
	var i int64

	switch val.(type) {
	case int:
		i = int64(val.(int))
	case int8:
		i = int64(val.(int8))
	case int16:
		i = int64(val.(int16))
	case int32:
		i = int64(val.(int32))
	case int64:
		i = val.(int64)
	case uint:
		i = int64(val.(uint))
	case uint8:
		i = int64(val.(uint8))
	case uint16:
		i = int64(val.(uint16))
	case uint32:
		i = int64(val.(uint32))
	case uint64:
		i = int64(val.(uint64))
	case bool:
		if val.(bool) == true {
			i = int64(1)
		} else {
			i = int64(0)
		}
	case float32:
		i = int64(val.(float32))
	case float64:
		i = int64(val.(float64))
	default:
		i, _ = strconv.ParseInt(String(val), 10, 64)
	}

	return i
}

func Uint(val interface{}) uint {
	return uint(Uint64(val))
}

func Uint8(val interface{}) uint8 {
	return uint8(Uint64(val))
}

func Uint16(val interface{}) uint16 {
	return uint16(Uint64(val))
}

func Uint32(val interface{}) uint32 {
	return uint32(Uint64(val))
}

func Uint64(val interface{}) uint64 {
	var i uint64

	switch val.(type) {
	case int:
		i = uint64(val.(int))
	case int8:
		i = uint64(val.(int8))
	case int16:
		i = uint64(val.(int16))
	case int32:
		i = uint64(val.(int32))
	case int64:
		i = uint64(val.(int64))
	case uint:
		i = uint64(val.(uint))
	case uint8:
		i = uint64(val.(uint8))
	case uint16:
		i = uint64(val.(uint16))
	case uint32:
		i = uint64(val.(uint32))
	case uint64:
		i = val.(uint64)
	case bool:
		if val.(bool) == true {
			i = uint64(1)
		} else {
			i = uint64(0)
		}
	case string:
		i, _ = strconv.ParseUint(val.(string), 10, 64)
	case float32:
		i = uint64(val.(float32))
	case float64:
		i = uint64(val.(float64))
	}

	return i
}

func Float32(val interface{}) float32 {
	return float32(Float64(val))
}

func Float64(val interface{}) float64 {
	var f float64

	switch val.(type) {
	case int:
		f = float64(val.(int))
	case int8:
		f = float64(val.(int8))
	case int16:
		f = float64(val.(int16))
	case int32:
		f = float64(val.(int32))
	case int64:
		f = float64(val.(int64))
	case uint:
		f = float64(val.(uint))
	case uint8:
		f = float64(val.(uint8))
	case uint16:
		f = float64(val.(uint16))
	case uint32:
		f = float64(val.(uint32))
	case uint64:
		f = float64(val.(uint64))
	case bool:
		if val.(bool) == true {
			f = float64(1)
		} else {
			f = float64(0)
		}
	case string:
		f, _ = strconv.ParseFloat(val.(string), 64)
	case float32:
		f = float64(val.(float32))
	case float64:
		f = val.(float64)
	}

	return f
}

func Bool(value interface{}) bool {
	b, _ := strconv.ParseBool(String(value))
	return b
}

func Convert(value interface{}, t reflect.Kind) (interface{}, error) {

	switch reflect.TypeOf(value).Kind() {
	case reflect.Slice:
		switch t {
		case reflect.String:
			if reflect.TypeOf(value).Elem().Kind() == reflect.Uint8 {
				return string(value.([]byte)), nil
			} else {
				return String(value), nil
			}
		case reflect.Slice:
		default:
			return nil, fmt.Errorf("Could not convert slice into non-slice.")
		}
	}

	switch t {

	case reflect.String:
		return String(value), nil

	case reflect.Uint64:
		return Uint64(value), nil

	case reflect.Uint32:
		return Uint32(value), nil

	case reflect.Uint16:
		return Uint16(value), nil

	case reflect.Uint8:
		return Uint8(value), nil

	case reflect.Uint:
		return Uint(value), nil

	case reflect.Int64:
		return Int64(value), nil

	case reflect.Int32:
		return Int32(value), nil

	case reflect.Int16:
		return Int16(value), nil

	case reflect.Int8:
		return Int8(value), nil

	case reflect.Int:
		return Int(value), nil

	case reflect.Float64:
		return Float64(value), nil

	case reflect.Float32:
		return Float32(value), nil

	case reflect.Bool:
		return Bool(value), nil

	case reflect.Interface:
		return value, nil

	}

	return nil, fmt.Errorf("Could not convert %s into %s.", reflect.TypeOf(value).Kind(), t)
}
