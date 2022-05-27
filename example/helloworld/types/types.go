package types

import (
    "errors"

    "github.com/leisurelyrcxf/gengin"
)

type Session struct {
    UID int64
}

type SigninRequest struct {
    Email    string
    Password string
}

func (p SigninRequest) Sanitize() gengin.Request { return p }

func (p SigninRequest) Validate() error {
    if p.Email == "" {
        return errors.New("email empty")
    }
    if p.Password == "" {
        return errors.New("password empty")
    }
    return nil
}

type SigninResponse struct {
    SessionID string
}

type ProfileRequest struct{}

func (p ProfileRequest) Sanitize() gengin.Request { return p }

func (p ProfileRequest) Validate() error { return nil }

type ProfileResponse struct {
    ID    int64
    Phone string
    Email string
}
