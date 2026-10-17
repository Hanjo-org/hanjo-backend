package validator

import (
	"errors"
	"testing"
)

type SampleDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,strongpassword"`
}

func TestStruct_Success(t *testing.T) {
	dto := SampleDTO{
		Email:    "test@example.com",
		Password: "ValidPassword123!",
	}

	err := Struct(dto)
	if err != nil {
		t.Fatalf("expected valid struct, got: %v", err)
	}
}

func TestStruct_ValidationFailure(t *testing.T) {
	dto := SampleDTO{
		Email:    "not-an-email",
		Password: "weak",
	}

	err := Struct(dto)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	var fieldErrs FieldErrors
	if !errors.As(err, &fieldErrs) {
		t.Fatalf("expected FieldErrors type, got %T", err)
	}

	if _, ok := fieldErrs["email"]; !ok {
		t.Errorf("expected validation error on 'email', got: %v", fieldErrs)
	}
	if _, ok := fieldErrs["password"]; !ok {
		t.Errorf("expected validation error on 'password', got: %v", fieldErrs)
	}
}
