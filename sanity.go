package sanity

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// less than
// greater than
// regex
// length
// Nested struct (objects)
// not null / empty

var (
	NotNullValidationFailedError         = errors.New("blank value sent to field marked as notnull")
	GTConditionValidationFailedError     = errors.New("value does not satisfy greater than condition")
	LTConditionValidationFailedError     = errors.New("value does not satisfy less than condition")
	RegexValidationFailedError           = errors.New("does not match pattern")
	MaxLenConditionValidationFailedError = errors.New("exceeds maximum length")
	MinLenConditionValidationFailedError = errors.New("len is smaller than minimum length")
)

var (
	Label       = "sanity"
	GreaterThan = "gt"
	LessThan    = "lt"
	Regex       = "regex"
	Optional    = "optional"
	NotNull     = "notnull"
	MaxLen      = "maxlen"
	MinLen      = "minlen"
)

type SanityInfo struct {
	Errors map[string]error `json:"errors"`
}

func Check(o interface{}) (SanityInfo, bool) {

	// todo: check for empty structs
	// todo: check if nil

	feedback := make(map[string]error)

	sanitize(reflect.ValueOf(o), &feedback)

	return SanityInfo{
		Errors: feedback,
	}, len(feedback) == 0

}

func sanitize(obj interface{}, fb *map[string]error, paths ...string) {
	val := obj.(reflect.Value)

	switch val.Kind() {
	case reflect.Ptr:
		sanitize(val.Elem(), fb, paths...)
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			tag := val.Type().Field(i).Tag
			field := val.Field(i)
			fieldValue := field.Interface()
			jsonKey := tag.Get("json")
			if jsonKey == "" {
				jsonKey = val.Type().Field(i).Name
			}
			currentPath := strings.Join(append(paths, jsonKey), ".")
			sanitizeTermString := tag.Get(Label)
			if sanitizeTermString == "" {
				return
			}
			sanitizeTerms := strings.Split(sanitizeTermString, ",")
			instructions := make(map[string]string)
			for _, term := range sanitizeTerms {
				if s := tag.Get(term); s != "" {
					instructions[term] = s
				}
			}

			// handle struct field
			switch field.Kind() {
			case reflect.Struct:
				sanitize(fieldValue, fb, append(paths, jsonKey)...)
			case reflect.Ptr:

				if v, ok := instructions[NotNull]; ok && v == "true" && reflect.ValueOf(fieldValue).IsNil() {
					(*fb)[currentPath] = NotNullValidationFailedError
					return
				}
				if reflect.ValueOf(fieldValue).IsNil() {
					return
				}

				if field.Elem().Kind() == reflect.Struct {
					sanitize(field.Elem(), fb, append(paths, jsonKey)...)
					return
				}
				// terminal field
				handleField(field.Elem(), instructions, fb, append(paths, jsonKey)...)
			default:
				handleField(field, instructions, fb, append(paths, jsonKey)...)

			}
		}

	default:
		// for all other cases
		return

	}
}

func handleField(field reflect.Value, instructions map[string]string, fb *map[string]error, path ...string) {
	switch field.Kind() {
	case reflect.String:
		s := field.Interface().(string)
		handleString(s, instructions, fb, path...)
	case reflect.Int64:
		handleInt64(field.Interface().(int64), instructions, fb, path...)
	case reflect.Int32:
		handleInt32(field.Interface().(int32), instructions, fb, path...)
	case reflect.Int16:
		handleInt16(field.Interface().(int16), instructions, fb, path...)
	case reflect.Int8:
		handleInt8(field.Interface().(int8), instructions, fb, path...)
	case reflect.Int:
		handleInt(field.Interface().(int), instructions, fb, path...)
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
	case reflect.Bool:
		handleBool(field.Interface().(bool), instructions, fb, path...)
	case reflect.Chan:
	case reflect.Array:
	case reflect.Map:
	case reflect.Slice:
		handleSlice(field.Interface().([]interface{}), instructions, fb, path...)
	}
}

func handleBool(b bool, instructions map[string]string, fb *map[string]error, path ...string) {
	completePath := strings.Join(path, ".")
	if v, ok := instructions[NotNull]; ok && v == "true" && b == false {
		(*fb)[completePath] = NotNullValidationFailedError
	}

}

func handleString(str string, instructions map[string]string, fb *map[string]error, path ...string) {
	completePath := strings.Join(path, ".")
	if v, ok := instructions[NotNull]; ok && v == "true" && str == "" {
		(*fb)[completePath] = NotNullValidationFailedError
	}

	if v, ok := instructions[Regex]; ok {
		pattern := regexp.MustCompile(v)
		if !pattern.Match([]byte(str)) {
			(*fb)[completePath] = RegexValidationFailedError
		}
	}

	if v, ok := instructions[MaxLen]; ok {
		n, _ := strconv.Atoi(v)
		if len(str) > n {
			(*fb)[completePath] = MaxLenConditionValidationFailedError
		}
	}

	if v, ok := instructions[MinLen]; ok {
		n, _ := strconv.Atoi(v)
		if len(str) < n {
			(*fb)[completePath] = MinLenConditionValidationFailedError
		}
	}
}

func handleSlice(slice []interface{}, instructions map[string]string, fb *map[string]error, path ...string) {
	completePath := strings.Join(path, ".")
	if v, ok := instructions[NotNull]; ok && v == "true" && len(slice) == 0 {
		(*fb)[completePath] = NotNullValidationFailedError
	}
	for i, v := range slice {
		sanitize(v, fb, append(path, strconv.Itoa(i))...)
	}
}

func handleInt64(n int64, instructions map[string]string, fb *map[string]error, path ...string) {
	if v, ok := instructions[NotNull]; ok && v == "true" && n == 0 {
		(*fb)[strings.Join(path, ".")] = NotNullValidationFailedError
	}

	if v, ok := instructions[GreaterThan]; ok {
		gt, _ := strconv.ParseInt(v, 10, 64)
		if n < gt {
			(*fb)[strings.Join(path, ".")] = GTConditionValidationFailedError
		}
	}

	if v, ok := instructions[LessThan]; ok {
		lt, _ := strconv.ParseInt(v, 10, 64)
		if n > lt {
			(*fb)[strings.Join(path, ".")] = LTConditionValidationFailedError
		}
	}
}

func handleInt32(n int32, instructions map[string]string, fb *map[string]error, path ...string) {
	if v, ok := instructions[NotNull]; ok && v == "true" && n == 0 {
		(*fb)[strings.Join(path, ".")] = NotNullValidationFailedError
	}

	if v, ok := instructions[GreaterThan]; ok {
		gt, _ := strconv.ParseInt(v, 10, 32)
		if n < int32(gt) {
			(*fb)[strings.Join(path, ".")] = GTConditionValidationFailedError
		}
	}

	if v, ok := instructions[LessThan]; ok {
		lt, _ := strconv.ParseInt(v, 10, 32)
		if n > int32(lt) {
			(*fb)[strings.Join(path, ".")] = LTConditionValidationFailedError
		}
	}
}

func handleInt16(n int16, instructions map[string]string, fb *map[string]error, path ...string) {
	if v, ok := instructions[NotNull]; ok && v == "true" && n == 0 {
		(*fb)[strings.Join(path, ".")] = NotNullValidationFailedError
	}

	if v, ok := instructions[GreaterThan]; ok {
		gt, _ := strconv.ParseInt(v, 10, 16)
		if n < int16(gt) {
			(*fb)[strings.Join(path, ".")] = GTConditionValidationFailedError
		}
	}

	if v, ok := instructions[LessThan]; ok {
		lt, _ := strconv.ParseInt(v, 10, 32)
		if n > int16(lt) {
			(*fb)[strings.Join(path, ".")] = LTConditionValidationFailedError
		}
	}
}

func handleInt8(n int8, instructions map[string]string, fb *map[string]error, path ...string) {
	if v, ok := instructions[NotNull]; ok && v == "true" && n == 0 {
		(*fb)[strings.Join(path, ".")] = NotNullValidationFailedError
	}

	if v, ok := instructions[GreaterThan]; ok {
		gt, _ := strconv.ParseInt(v, 10, 16)
		if n < int8(gt) {
			(*fb)[strings.Join(path, ".")] = GTConditionValidationFailedError
		}
	}

	if v, ok := instructions[LessThan]; ok {
		lt, _ := strconv.ParseInt(v, 10, 32)
		if n > int8(lt) {
			(*fb)[strings.Join(path, ".")] = LTConditionValidationFailedError
		}
	}
}

func handleInt(n int, instructions map[string]string, fb *map[string]error, path ...string) {
	if v, ok := instructions[NotNull]; ok && v == "true" && n == 0 {
		(*fb)[strings.Join(path, ".")] = NotNullValidationFailedError
	}

	if v, ok := instructions[GreaterThan]; ok {
		gt, _ := strconv.Atoi(v)
		if n < gt {
			(*fb)[strings.Join(path, ".")] = GTConditionValidationFailedError
		}
	}

	if v, ok := instructions[LessThan]; ok {
		lt, _ := strconv.Atoi(v)
		if n > lt {
			(*fb)[strings.Join(path, ".")] = LTConditionValidationFailedError
		}
	}
}
