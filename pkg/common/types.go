package common

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
)

type AmpKVDataType int

const (
	TypeUnknown AmpKVDataType = iota
	TypeString
	TypeInt
	TypeFloat
	TypeBool
	TypeJSON
	TypeBinary
)

func (t AmpKVDataType) String() string {
	switch t {
	case TypeUnknown:
		return "Unknown"
	case TypeString:
		return "String"
	case TypeInt:
		return "Int"
	case TypeFloat:
		return "Float"
	case TypeBool:
		return "Bool"
	case TypeJSON:
		return "JSON"
	case TypeBinary:
		return "Binary"
	default:
		return fmt.Sprintf("AmpKVDataType(%d)", t)
	}
}

type AmpKVValue struct {
	Type AmpKVDataType
	Data []byte
}

func (t *AmpKVValue) ToByteSlice() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(t)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode AmpKV Type to byte slice: %w", err)
	}

	return buffer.Bytes(), nil
}

func AmpKVValueFrom(data []byte) (*AmpKVValue, error) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	var decodedAmpKVValue AmpKVValue
	err := decoder.Decode(&decodedAmpKVValue)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode AmpKV Type from data: %w", err)
	}

	return &decodedAmpKVValue, nil
}

func NewAmpKVValue(value any) (*AmpKVValue, error) {
	if value == nil {
		return nil, fmt.Errorf("value cannot be nil")
	}

	valType := reflect.TypeOf(value)
	valValue := reflect.ValueOf(value)

	var (
		ampKVData []byte
		ampKVType AmpKVDataType
	)

	switch valType.Kind() {
	case reflect.String:
		ampKVType = TypeString
		ampKVData = []byte(valValue.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ampKVType = TypeInt
		var buf [8]byte
		uval := uint64(valValue.Int())
		binary.BigEndian.PutUint64(buf[:], uval)
		ampKVData = buf[:]
	case reflect.Float32, reflect.Float64:
		ampKVType = TypeFloat
		var buf [8]byte
		f64Val := valValue.Float()
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(f64Val))
		ampKVData = buf[:]
	case reflect.Bool:
		ampKVType = TypeBool
		ampKVData = make([]byte, 1)
		if valValue.Bool() {
			ampKVData[0] = 1
		} else {
			ampKVData[0] = 0
		}
	case reflect.Slice:
		if valType.Elem().Kind() == reflect.Uint8 {
			ampKVType = TypeBinary
			ampKVData = value.([]byte)
		} else {
			ampKVType = TypeJSON
			jsonData, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value: %w", err)
			}
			ampKVData = jsonData
		}
	case reflect.Map, reflect.Struct:
		ampKVType = TypeJSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}
		ampKVData = jsonData
	case reflect.Ptr:
		if valValue.IsNil() {
			return nil, fmt.Errorf("nil pointer value provided")
		}
		return NewAmpKVValue(valValue.Elem().Interface())
	default:
		ampKVType = TypeJSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("unsupported type %T for direct conversion; failed to marshal to JSON: %w", value, err)
		}
		ampKVData = jsonData
	}

	return &AmpKVValue{
		Data: ampKVData,
		Type: ampKVType,
	}, nil
}

func (v *AmpKVValue) AsString() (string, error) {
	if v.Type != TypeString {
		return "", fmt.Errorf("data is not a string, but %s", v.Type.String())
	}
	return BytesToString(v.Data)
}

func (v *AmpKVValue) AsInt() (int, error) {
	if v.Type != TypeInt {
		return 0, fmt.Errorf("data is not an int, but %s", v.Type.String())
	}
	return BytesToInt(v.Data)
}

func (v *AmpKVValue) AsInt8() (int8, error) {
	if v.Type != TypeInt {
		return 0, fmt.Errorf("data is not an int, but %s", v.Type.String())
	}
	return BytesToInt8(v.Data)
}

func (v *AmpKVValue) AsInt16() (int16, error) {
	if v.Type != TypeInt {
		return 0, fmt.Errorf("data is not an int, but %s", v.Type.String())
	}
	return BytesToInt16(v.Data)
}

func (v *AmpKVValue) AsInt32() (int32, error) {
	if v.Type != TypeInt {
		return 0, fmt.Errorf("data is not an int, but %s", v.Type.String())
	}
	return BytesToInt32(v.Data)
}

func (v *AmpKVValue) AsInt64() (int64, error) {
	if v.Type != TypeInt {
		return 0, fmt.Errorf("data is not an int, but %s", v.Type.String())
	}
	return BytesToInt64(v.Data)
}

func (v *AmpKVValue) AsFloat32() (float32, error) {
	if v.Type != TypeFloat {
		return 0, fmt.Errorf("data is not a float, but %s", v.Type.String())
	}
	return BytesToFloat32(v.Data)
}

func (v *AmpKVValue) AsFloat64() (float64, error) {
	if v.Type != TypeFloat {
		return 0, fmt.Errorf("data is not a float, but %s", v.Type.String())
	}
	return BytesToFloat64(v.Data)
}

func (v *AmpKVValue) AsBool() (bool, error) {
	if v.Type != TypeBool {
		return false, fmt.Errorf("data is not a bool, but %s", v.Type.String())
	}

	return BytesToBool(v.Data)
}

func (_v *AmpKVValue) AsJson(v any) error {
	if _v.Type != TypeJSON {
		return fmt.Errorf("data is not json, but %s", _v.Type.String())
	}
	return BytesToJson(_v.Data, v)
}

func (v *AmpKVValue) AsBinary() ([]byte, error) {
	if v.Type != TypeBinary {
		return nil, fmt.Errorf("data is not binary, but %s", v.Type.String())
	}
	return v.Data, nil
}

func (v *AmpKVValue) Bytes() ([]byte, error) {
	return v.Data, nil
}

func BytesToString(bytes []byte) (string, error) {
	return string(bytes), nil
}

func getInt64FromBinary(bytes []byte) (int64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("invalid binary data length for int64: expected 8 bytes, got %d", len(bytes))
	}
	uval := binary.BigEndian.Uint64(bytes)
	val := int64(uval)
	return val, nil
}

func BytesToInt(bytes []byte) (int, error) {
	val, err := getInt64FromBinary(bytes)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt || val > math.MaxInt {
		return 0, fmt.Errorf("int overflow: %d is out of range for type int", val)
	}
	return int(val), nil
}

func BytesToInt8(bytes []byte) (int8, error) {
	val, err := getInt64FromBinary(bytes)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt8 || val > math.MaxInt8 {
		return 0, fmt.Errorf("int overflow: %d is out of range for type int8", val)
	}
	return int8(val), nil
}

func BytesToInt16(bytes []byte) (int16, error) {
	val, err := getInt64FromBinary(bytes)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt16 || val > math.MaxInt16 {
		return 0, fmt.Errorf("int overflow: %d is out of range for type int16", val)
	}
	return int16(val), nil
}

func BytesToInt32(bytes []byte) (int32, error) {
	val, err := getInt64FromBinary(bytes)
	if err != nil {
		return 0, err
	}
	if val < math.MinInt32 || val > math.MaxInt32 {
		return 0, fmt.Errorf("int overflow: %d is out of range for type int32", val)
	}
	return int32(val), nil
}

func BytesToInt64(bytes []byte) (int64, error) {
	return getInt64FromBinary(bytes)
}

func getFloat64FromBinary(bytes []byte) (float64, error) {
	if len(bytes) != 8 {
		return 0, fmt.Errorf("invalid binary data length for float64: expected 8 bytes, got %d", len(bytes))
	}
	uval := binary.BigEndian.Uint64(bytes)
	val := math.Float64frombits(uval)
	return val, nil
}

func BytesToFloat32(bytes []byte) (float32, error) {
	f64Val, err := getFloat64FromBinary(bytes)
	if err != nil {
		return 0, err
	}

	return float32(f64Val), nil
}

func BytesToFloat64(bytes []byte) (float64, error) {
	return getFloat64FromBinary(bytes)
}

func BytesToBool(bytes []byte) (bool, error) {
	if len(bytes) != 1 {
		return false, fmt.Errorf("invalid binary data length for bool: expected 1 byte, got %d", len(bytes))
	}

	return bytes[0] == 1, nil
}

func BytesToJson(bytes []byte, v any) error {
	return json.Unmarshal(bytes, v)
}
