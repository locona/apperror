package apperror

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	ER_BAD_FIELD_ERROR        = 1054
	ER_DUP_FIELDNAME          = 1060
	ER_DUP_ENTRY              = 1062
	ER_NOT_NULL_VIOLATION     = 1048
	ER_CANT_DROP_FIELD_OR_KEY = 1091
	ER_NO_REFERENCED_ROW_2    = 1452
	ER_DATA_TOO_LONG          = 1406
	ER_OUT_OF_RANGE           = 1264
)

var (
	ERR_MESSAGE_FORMAT = map[int]string{
		ER_BAD_FIELD_ERROR:        "Unknown column '%s' in '%s'",
		ER_DUP_FIELDNAME:          "Duplicate column name '%s'",
		ER_DUP_ENTRY:              "Duplicate entry '%s' for key %s",
		ER_NOT_NULL_VIOLATION:     "Column '%s' cannot be null",
		ER_CANT_DROP_FIELD_OR_KEY: "Can't DROP '%s'; check that column/key exists",
	}
	ERR_MESSAGES = map[int]string{
		ER_BAD_FIELD_ERROR:        "unknown column",
		ER_DUP_FIELDNAME:          "has already been taken",
		ER_DUP_ENTRY:              "has already been taken",
		ER_NOT_NULL_VIOLATION:     "cant't be blank",
		ER_CANT_DROP_FIELD_OR_KEY: "check that column exists",
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

func MysqlError(errMessage error) error {
	me, ok := errMessage.(*mysql.MySQLError)
	if !ok {
		return RecordError{Field: "", Message: errMessage.Error()}
	}

	in := strings.Replace(me.Message, "'", "", -1)
	f := strings.Replace(ERR_MESSAGE_FORMAT[int(me.Number)], "'", "", -1)
	var field string
	fmt.Sscanf(in, f, &field)
	return RecordError{Field: field, Message: ERR_MESSAGES[int(me.Number)]}
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
