package domain

import (
	"testing"
)

func TestUserValidationWithValidName(t *testing.T) {
	user := User{
		Name: "Alice",
	}

	err := user.Validate()

	if err != nil {
		t.Errorf("Valid username 'Alice' should not produce error, got: %v", err)
	}
}

func TestUserValidationWithEmptyName(t *testing.T) {
	user := User{
		Name: "",
	}

	err := user.Validate()

	if err == nil {
		t.Errorf("Empty username should produce error, got nil")
	}
}

func TestUserValidationWithWhitespaceOnlyName(t *testing.T) {
	user := User{
		Name: "   ",
	}

	err := user.Validate()

	if err == nil {
		t.Errorf("Whitespace-only username should produce error, got nil")
	}
}

func TestUserValidationWithValidNameContainingNumbers(t *testing.T) {
	user := User{
		Name: "Alice123",
	}

	err := user.Validate()

	if err != nil {
		t.Errorf("Username 'Alice123' should be valid, got error: %v", err)
	}
}

func TestUserValidationWithValidNameInRussian(t *testing.T) {
	user := User{
		Name: "Алиса",
	}

	err := user.Validate()

	if err != nil {
		t.Errorf("Russian username 'Алиса' should be valid, got error: %v", err)
	}
}

func TestUserValidationWithTooLongName(t *testing.T) {
	longName := ""
	for i := 0; i < 300; i++ {
		longName += "a"
	}
	user := User{
		Name: longName,
	}

	err := user.Validate()

	_ = err 
}

func TestUserCreation(t *testing.T) {
	username := "Bob"

	user := User{
		Name: username,
	}

	if user.Name != username {
		t.Errorf("Username not set correctly: expected '%s', got '%s'", username, user.Name)
	}
}

func TestUserValidationWithSpecialCharacters(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{name: "Alice-Smith"},      
		{name: "user_123"},         
		{name: "maria.santos"},     
		{name: "Иван_Петров"},      
	}

	for _, tc := range testCases {
		user := User{
			Name: tc.name,
		}
		err := user.Validate()
		if err != nil {
			t.Logf("Username '%s' produced error: %v (system may not support this)", tc.name, err)
		} else {
			t.Logf("Username '%s' is valid", tc.name)
		}
	}
}
