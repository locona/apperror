package apperror

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	ER_DUPLICATE_ENTRY     = 1062
	ER_NOT_NULL_VIOLATION  = 1048
	ER_NO_REFERENCED_ROW_2 = 1452
	ER_DATA_TOO_LONG       = 1406
	ER_OUT_OF_RANGE        = 1264
)

var (
	columnRegexp = regexp.MustCompile("'.+?'")
	ERR_MESSAGES = map[int]string{
		ER_DUPLICATE_ENTRY:     "has already been taken",
		ER_NOT_NULL_VIOLATION:  "cant't be blank",
		ER_NO_REFERENCED_ROW_2: "cannot add or update a child row",
		ER_DATA_TOO_LONG:       "data too long",
		ER_OUT_OF_RANGE:        "out of range value",
	}
)

type RecordError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (this RecordError) Error() string {
	return fmt.Sprintf("Error %s: %s", this.Field, this.Message)
}

func CustomError(field, message string) error {
	return RecordError{Field: field, Message: message}
}

func MysqlError(err error, obj interface{}) error {
	me, ok := err.(*mysql.MySQLError)
	if !ok {
		return RecordError{Field: "", Message: err.Error()}
	}
	messages := columnRegexp.FindAllString(me.Message, -1)
	reflectType := reflect.ValueOf(obj).Type()
	for i := 0; i < reflectType.NumField(); i++ {
		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			field := structField{
				Name:        convertCamelToLower(fieldStruct.Name),
				Tag:         fieldStruct,
				TagSettings: parseTagSetting(fieldStruct.Tag),
			}

			for idx := range messages {
				if strings.Replace(messages[idx], "'", "", -1) == field.Name {
					return RecordError{
						Field:   field.Name,
						Message: ERR_MESSAGES[int(me.Number)],
					}
				}
			}
		}
	}

	return RecordError{Field: "", Message: me.Error()}
}

type structField struct {
	Name        string
	Tag         reflect.StructField
	TagSettings map[string]string
}

func parseTagSetting(fieldTag reflect.StructTag) map[string]string {
	setting := map[string]string{}
	tags := strings.Split(fieldTag.Get("gorm"), ";")
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) >= 2 {
			setting[k] = strings.Join(v[1:], ":")
		} else {
			setting[k] = k
		}
	}

	return setting
}

func convertCamelToLower(s string) string {
	camel := regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")
	var a []string
	for _, sub := range camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}
