package service

import (
    "context"
    "github.com/leisurelyrcxf/gengin/example/helloworld/types"
)

type Service interface {
    Auth(ctx context.Context, token string) (*types.Session, error)

    SignIn(ctx context.Context, req types.SigninRequest) (types.SigninResponse, error)
    Profile(ctx context.Context, req types.ProfileRequest, session *types.Session) (types.ProfileResponse, error)
}
