package helloworld

import (
    "fmt"
    "log"

    "github.com/gin-gonic/gin"

    "github.com/leisurelyrcxf/gengin"
    "github.com/leisurelyrcxf/gengin/example/helloworld/service"
    "github.com/leisurelyrcxf/gengin/example/helloworld/types"
)

type Server struct {
    *gin.Engine

    service service.Service

    ServiceDescription *gengin.Services[*types.Session]

    port int
}

func NewServer(port int) *Server {
    s := &Server{
        Engine: gin.New(),
        port:   port,
    }
    return s
}

func (s *Server) RegisterServices() error {
    s.service = service.NewServiceImpl()
    v1 := s.Group("v1")

    s.ServiceDescription = gengin.NewServices("usr", v1, "User", s.service.Auth, nil)
    gengin.RegisterService(s.ServiceDescription, "SignIn", "POST", "login", s.service.SignIn)
    gengin.RegisterAuthenticatedService(s.ServiceDescription, "Profile", "GET", "get user profile", s.service.Profile)
    return nil
}

func (s *Server) Serve() (err error) {
    if s.service == nil {
        if err := s.RegisterServices(); err != nil {
            return err
        }
    }
    log.Printf("HTTP Server started")
    return s.Run(fmt.Sprintf("localhost:%d", s.port))
}
