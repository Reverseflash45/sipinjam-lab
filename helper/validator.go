package helper

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

// ValidateStruct mengubah hasil validator menjadi map field -> pesan yang
// bisa langsung ditampilkan frontend. nil berarti valid.
func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return map[string]string{"_": err.Error()}
	}
	out := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		out[jsonName(fe.Namespace())] = message(fe)
	}
	return out
}

// jsonName mengubah "CreateBookingRequest.Items[0].ItemID" menjadi
// "items[0].item_id" supaya cocok dengan nama field di body JSON.
func jsonName(ns string) string {
	parts := strings.Split(ns, ".")[1:]
	for i, p := range parts {
		parts[i] = snake(p)
	}
	return strings.Join(parts, ".")
}

func snake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && s[i-1] != '[' && !(s[i-1] >= 'A' && s[i-1] <= 'Z') {
				b.WriteByte('_')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "numeric":
		return "hanya boleh berisi angka"
	case "min":
		return "minimal " + fe.Param()
	case "max":
		return "maksimal " + fe.Param()
	case "oneof":
		return "harus salah satu dari: " + fe.Param()
	default:
		return "tidak valid (" + fe.Tag() + ")"
	}
}
