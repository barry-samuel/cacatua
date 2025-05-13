package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

func SendResponseJSON(w http.ResponseWriter, r *http.Request, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func nonZero(v reflect.Value) bool {
	// A zero Value is “invalid” or equal to its Type’s zero.
	return v.IsValid() && !reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// MapStruct copies exported, non-zero fields from src to dst by matching field names.
// Both src and dst must be pointers to structs; dst receives updates.
func MapStruct(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)
	// Ensure pointers
	if srcVal.Kind() != reflect.Ptr || dstVal.Kind() != reflect.Ptr {
		return errors.New("src and dst must be pointers to structs")
	}
	srcElem := reflect.Indirect(srcVal)
	dstElem := reflect.Indirect(dstVal)
	// Ensure underlying structs
	if srcElem.Kind() != reflect.Struct || dstElem.Kind() != reflect.Struct {
		return errors.New("src and dst must point to structs")
	}

	dstType := dstElem.Type()
	for i := 0; i < dstElem.NumField(); i++ {
		field := dstType.Field(i)
		// Only exported fields (start with uppercase)
		if !field.IsExported() {
			continue
		}
		dstField := dstElem.Field(i)
		if !dstField.CanSet() {
			continue
		}
		// Attempt to find matching field in src
		srcField := srcElem.FieldByName(field.Name)
		if !srcField.IsValid() {
			continue
		}
		// Only copy non-zero values
		if nonZero(srcField) {
			dstField.Set(srcField)
		}
	}
	return nil
}
