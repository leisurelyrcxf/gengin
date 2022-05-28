package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"

	"github.com/leisurelyrcxf/gengin/example/helloworld/types"
)

const (
	sessionIDInputLen = 16
	sessionIDLen      = 22
)

func NewServiceImpl() *Impl {
	return &Impl{users: []User{{
		ID:       1,
		Email:    "john@gmail.com",
		Password: "123456",
		Phone:    "11111111",
	}, {
		ID:       2,
		Email:    "jim@gmail.com",
		Password: "123456",
		Phone:    "22222222",
	}}, sessions: make(map[string]*types.Session)}
}

type User struct {
	ID int64

	Email    string
	Password string

	Phone string
}

type Impl struct {
	users []User

	sync.RWMutex
	sessions map[string]*types.Session
}

func genSessionID() string {
	var (
		srcBuf [sessionIDInputLen]byte
		dstBuf [sessionIDLen]byte

		src = srcBuf[0:]
		dst = dstBuf[0:]
	)
	if _, err := io.ReadFull(rand.Reader, src); err != nil {
		panic(err)
	}
	base64.RawURLEncoding.Encode(dst, src)
	return string(dst)
}

func (s *Impl) Auth(_ context.Context, token string) (*types.Session, error) {
	s.RLock()
	session, ok := s.sessions[token]
	s.RUnlock()

	if !ok {
		return nil, errors.New("session not found")
	}
	return session, nil
}

func (s *Impl) SignIn(_ context.Context, req types.SigninRequest) (types.SigninResponse, error) {
	for _, u := range s.users {
		if u.Email == req.Email && u.Password == req.Password {
			sessionID := genSessionID()

			s.Lock()
			s.sessions[sessionID] = &types.Session{UID: u.ID}
			s.Unlock()

			return types.SigninResponse{SessionID: sessionID}, nil
		}
	}
	return types.SigninResponse{}, errors.New("user not found")
}

func (s *Impl) Profile(_ context.Context, _ types.ProfileRequest, session *types.Session) (types.ProfileResponse, error) {
	for _, u := range s.users {
		if u.ID == session.UID {
			return types.ProfileResponse{
				ID:    u.ID,
				Phone: u.Phone,
				Email: u.Email,
			}, nil
		}
	}
	return types.ProfileResponse{}, errors.New("user not found")
}
