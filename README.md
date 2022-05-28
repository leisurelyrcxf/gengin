A web framework using generic and gin.

Gengin allows you to focus on your business logic. No need to write any code regard to http handler or router any more.

Gengin can also generate restful inteface descriptions automatically.

You can use just a few lines code to create a web application, sth like below:

```
    v1 := s.Group("v1")

    s.ServiceDescription = gengin.NewServices("usr", v1, "User", s.service.Auth, nil)
	
    gengin.RegisterService(s.ServiceDescription, "SignIn", "", "POST", "login", s.service.SignIn)
    gengin.RegisterAuthenticatedService(s.ServiceDescription, "Profile", "", "GET", "get user profile", s.service.Profile)
```

Refer to example to find more detail.
