package dist

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"

	"github.com/buildpacks/pack/internal/style"
)

type Labeled interface {
	Label(name string) (value string, err error)
}

type Labelable interface {
	SetLabel(name string, value string) error
}

func SetLabel(labelable Labelable, label string, data interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "marshalling data to JSON for label %s", style.Symbol(label))
	}
	if err := labelable.SetLabel(label, string(dataBytes)); err != nil {
		return errors.Wrapf(err, "setting label %s", style.Symbol(label))
	}
	return nil
}

func id_from(labeled Labeled, label string) (string, error) {
	labelData, err := labeled.Label(label)
	if err != nil {
		return "", errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	return fmt.Sprintf("%s-%s", labelData, label), nil
}

var cache map[string]interface{}

func GetLabel(labeled Labeled, label string, obj interface{}) (ok bool, err error) {
	labelData, err := labeled.Label(label)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		if err := json.Unmarshal([]byte(labelData), obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		return true, nil
	}
	return false, nil
}

func getLabel(labeled Labeled, label string, obj interface{}) (ok bool, err error) {
	labelData, err := labeled.Label(label)
	if err != nil {
		return false, errors.Wrapf(err, "retrieving label %s", style.Symbol(label))
	}
	if labelData != "" {
		key := fmt.Sprintf("%s-%s", labelData, label)
		if cache == nil {
			cache = make(map[string]interface{})
		}
		if cachedObj, ok := cache[key]; ok {
			Copy(cachedObj, obj)
			return true, nil
		}
		if err := json.Unmarshal([]byte(labelData), obj); err != nil {
			return false, errors.Wrapf(err, "unmarshalling label %s", style.Symbol(label))
		}
		cache[key] = obj
		return true, nil
	}
	return false, nil
}

// Interface for delegating copy process to type
type Interface interface {
	DeepCopy() interface{}
}

// Iface is an alias to Copy; this exists for backwards compatibility reasons.
func Iface(src interface{}, des interface{}) {
	Copy(src, des)
}

// Copy creates a deep copy of whatever is passed to it and returns the copy
// in an interface{}.  The returned value will need to be asserted to the
// correct type.
func Copy(src interface{}, des interface{}) {
	if src == nil {
		return
	}
	original := reflect.ValueOf(src)
	cpy := reflect.ValueOf(des)
	copyRecursive(original, cpy)
}

// copyRecursive does the actual copying of the interface. It currently has
// limited support for what it can handle. Add as needed.
func copyRecursive(original reflect.Value, cpy reflect.Value) {
	switch original.Kind() {
	case reflect.Pointer:
		originalValue := original.Elem()

		if !originalValue.IsValid() {
			return
		}
		if cpy.IsNil() {
			cpy.Set(reflect.New(originalValue.Type()))
		}
		copyRecursive(originalValue, cpy.Elem())

	case reflect.Interface:
		// If this is a nil, don't do anything
		if original.IsNil() {
			return
		}
		// Get the value for the interface, not the pointer.
		originalValue := original.Elem()

		// Get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		cpy.Set(copyValue)

	case reflect.Struct:
		t, ok := original.Interface().(time.Time)
		if ok {
			cpy.Set(reflect.ValueOf(t))
			return
		}
		// Go through each field of the struct and copy it.
		for i := 0; i < original.NumField(); i++ {
			// The Type's StructField for a given field is checked to see if StructField.PkgPath
			// is set to determine if the field is exported or not because CanSet() returns false
			// for settable fields.  I'm not sure why.  -mohae
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(original.Field(i), cpy.Field(i))
		}

	case reflect.Slice:
		if original.IsNil() {
			return
		}
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i))
		}

	case reflect.Map:
		if original.IsNil() {
			return
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := reflect.New(key.Type()).Elem()
			copyRecursive(key, copyKey)
			cpy.SetMapIndex(copyKey, copyValue)
		}

	default:
		cpy.Set(original)
	}
}
