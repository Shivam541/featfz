package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shivam/featfz/feat-manager/internal/http/response"
)

func RequiredHeader(r *http.Request, name string) (string, *response.ErrorDetail) {
	value := strings.TrimSpace(r.Header.Get(name))
	if value == "" {
		return "", &response.ErrorDetail{Field: name, Message: "is required"}
	}

	return value, nil
}

func RequiredQuery(r *http.Request, name string) (string, *response.ErrorDetail) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return "", &response.ErrorDetail{Field: name, Message: "is required"}
	}

	return value, nil
}

func RequiredPath(r *http.Request, name string) (string, *response.ErrorDetail) {
	value := strings.TrimSpace(r.PathValue(name))
	if value == "" {
		return "", &response.ErrorDetail{Field: name, Message: "is required"}
	}

	return value, nil
}

func DecodeJSONBody(r *http.Request, dst any) []response.ErrorDetail {
	if contentType := strings.TrimSpace(r.Header.Get("Content-Type")); contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		return []response.ErrorDetail{{Field: "body", Message: "must be application/json"}}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return []response.ErrorDetail{decodeError(err)}
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return []response.ErrorDetail{{Field: "body", Message: "must contain a single JSON object"}}
	}

	return nil
}

func decodeError(err error) response.ErrorDetail {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError

	switch {
	case errors.Is(err, io.EOF):
		return response.ErrorDetail{Field: "body", Message: "is required"}
	case errors.As(err, &syntaxErr):
		return response.ErrorDetail{Field: "body", Message: "contains invalid JSON"}
	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return response.ErrorDetail{Field: typeErr.Field, Message: fmt.Sprintf("must be %s", typeErr.Type.String())}
		}

		return response.ErrorDetail{Field: "body", Message: "contains an invalid value"}
	default:
		if strings.HasPrefix(err.Error(), "json: unknown field ") {
			field := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), "\"")
			return response.ErrorDetail{Field: field, Message: "is not allowed"}
		}

		return response.ErrorDetail{Field: "body", Message: "contains invalid JSON"}
	}
}
