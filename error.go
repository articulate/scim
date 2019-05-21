package scim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/elimity-com/scim/errors"
)

type scimType string

const (
	// One or more of the attribute values are already in use or are reserved.
	scimTypeUniqueness = "uniqueness"
	// The attempted modification is not compatible with the target attribute's mutability or current state (e.g.,
	// modification of an "immutable" attribute with an existing value).
	scimTypeMutability = "mutability"
	// The request body message structure was invalid or did not conform to the request schema.
	scimTypeInvalidSyntax = "invalidSyntax"
	// A required value was missing, or the value specified was not compatible with the operation or attribute type,
	// or resource schema.
	scimTypeInvalidValue = "invalidValue"
)

func scimErrorResourceNotFound(id string) scimError {
	return scimError{
		detail: fmt.Sprintf("Resource %s not found.", id),
		status: http.StatusNotFound,
	}
}

var scimErrorNil scimError

var (
	scimErrorUniqueness = scimError{
		scimType: scimTypeUniqueness,
		detail:   "One or more of the attribute values are already in use or are reserved.",
		status:   http.StatusConflict,
	}
	scimErrorMutability = scimError{
		scimType: scimTypeMutability,
		detail:   "The attempted modification is not compatible with the target attribute's mutability or current state.",
		status:   http.StatusBadRequest,
	}
	scimErrorInvalidSyntax = scimError{
		scimType: scimTypeInvalidSyntax,
		detail:   "The request body message structure was invalid or did not conform to the request schema.",
		status:   http.StatusBadRequest,
	}
	scimErrorInvalidValue = scimError{
		scimType: scimTypeInvalidValue,
		detail:   "A required value was missing, or the value specified was not compatible with the operation or attribute type, or resource schema.",
		status:   http.StatusBadRequest,
	}
	scimErrorInternalServer = scimError{
		status: http.StatusInternalServerError,
	}
)

// RFC: https://tools.ietf.org/html/rfc7644#section-3.12
type scimError struct {
	// scimType is a SCIM detail error keyword. OPTIONAL.
	scimType scimType
	// detail is a detailed human-readable message. OPTIONAL.
	detail string
	// status is the HTTP status code expressed as a JSON string. REQUIRED.
	status int
}

func (e scimError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Schemas  []string `json:"schemas"`
		ScimType scimType `json:"scimType,omitempty"`
		Detail   string   `json:"detail,omitempty"`
		Status   string   `json:"status"`
	}{
		Schemas:  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		ScimType: e.scimType,
		Detail:   e.detail,
		Status:   strconv.Itoa(e.status),
	})
}

func (e *scimError) UnmarshalJSON(data []byte) error {
	var tmpScimError struct {
		ScimType scimType
		Detail   string
		Status   string
	}

	err := json.Unmarshal(data, &tmpScimError)
	if err != nil {
		return err
	}

	status, err := strconv.Atoi(tmpScimError.Status)
	if err != nil {
		return err
	}

	*e = scimError{
		scimType: tmpScimError.ScimType,
		detail:   tmpScimError.Detail,
		status:   status,
	}

	return nil
}

func scimGetError(getError errors.GetError, id string) scimError {
	switch getError {
	case errors.GetErrorInvalidValue:
		return scimErrorInvalidValue
	case errors.GetErrorResourceNotFound:
		return scimErrorResourceNotFound(id)
	default:
		return scimErrorInternalServer
	}
}

func scimGetAllError(getError errors.GetAllError) scimError {
	switch getError {
	case errors.GetAllErrorInvalidValue:
		return scimErrorInvalidValue
	default:
		return scimErrorInternalServer
	}
}

func scimPostError(postError errors.PostError) scimError {
	switch postError {
	case errors.PostErrorUniqueness:
		return scimErrorUniqueness
	case errors.PostErrorInvalidSyntax:
		return scimErrorInvalidSyntax
	case errors.PostErrorInvalidValue:
		return scimErrorInvalidValue
	default:
		return scimErrorInternalServer
	}
}

func scimPutError(putError errors.PutError, id string) scimError {
	switch putError {
	case errors.PutErrorUniqueness:
		return scimErrorUniqueness
	case errors.PutErrorMutability:
		return scimErrorMutability
	case errors.PutErrorInvalidSyntax:
		return scimErrorInvalidSyntax
	case errors.PutErrorInvalidValue:
		return scimErrorInvalidValue
	case errors.PutErrorResourceNotFound:
		return scimErrorResourceNotFound(id)
	default:
		return scimErrorInternalServer
	}
}

func scimDeleteError(deleteError errors.DeleteError, id string) scimError {
	switch deleteError {
	case errors.DeleteErrorResourceNotFound:
		return scimErrorResourceNotFound(id)
	default:
		return scimErrorInternalServer
	}
}
