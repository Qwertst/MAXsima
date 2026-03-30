package model

import (
	"errors"
	"strings"
)

type User struct {
	Name string
}

func (u User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return errors.New("username cannot be empty")
	}
	return nil
}
