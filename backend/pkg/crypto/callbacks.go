package crypto

import (
	"context"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func RegisterEncryptionCallbacks(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("crypto:encrypt_create", encryptCallback)
	db.Callback().Update().Before("gorm:update").Register("crypto:encrypt_update", encryptCallback)
	db.Callback().Query().After("gorm:query").Register("crypto:decrypt_query", decryptCallback)
}

func encryptCallback(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	for _, field := range db.Statement.Schema.Fields {
		if !strings.Contains(field.Tag.Get("fayhub"), "encrypt") {
			continue
		}

		val, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
		if isZero || val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}

		if !IsEncryptionEnabled() {
			continue
		}

		encrypted, err := Encrypt(strVal)
		if err != nil {
			continue
		}

		field.Set(db.Statement.Context, db.Statement.ReflectValue, encrypted)
	}
}

func decryptCallback(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	if !IsEncryptionEnabled() {
		return
	}

	rv := db.Statement.ReflectValue
	if rv.Kind() == reflect.Slice {
		for i := 0; i < rv.Len(); i++ {
			decryptStruct(rv.Index(i), db.Statement.Schema)
		}
	} else if rv.Kind() == reflect.Struct {
		decryptStruct(rv, db.Statement.Schema)
	}
}

func decryptStruct(structVal reflect.Value, sch *schema.Schema) {
	for _, field := range sch.Fields {
		if !strings.Contains(field.Tag.Get("fayhub"), "encrypt") {
			continue
		}

		val, isZero := field.ValueOf(context.Background(), structVal)
		if isZero || val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok || strVal == "" {
			continue
		}

		decrypted, err := Decrypt(strVal)
		if err != nil {
			continue
		}

		if decrypted != strVal {
			field.Set(context.Background(), structVal, decrypted)
		}
	}
}
