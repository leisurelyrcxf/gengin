A web framework using generic and gin.

Gengin allows you to focus on your business logic. No need to write any code regard to http handler or router any more.

Gengin can also generate restful inteface descriptions automatically.

You can use just a few lines code to create a web application, sth like below:

```
    type Service interface {
        Auth(ctx context.Context, token string) (*types.Session, error)
    
        SignIn(ctx context.Context, req types.SigninRequest) (types.SigninResponse, error)
        Profile(ctx context.Context, req types.ProfileRequest, session *types.Session) (types.ProfileResponse, error)
    }
    
    ....
    
    var srv Service = your_service_impl
    
    v1 := s.Group("v1")
    services := gengin.NewServices("usr", v1, "User", srv.Auth, nil)
	
    gengin.RegisterService(services, "SignIn", "", "POST", "login", srv.SignIn)
    gengin.RegisterAuthenticatedService(services, "Profile", "", "GET", "get user profile", srv.Profile)
```

Refer to example to find more detail.
