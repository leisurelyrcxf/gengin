package gengin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	SessionCtxKey = "session"

	bearerLength = len("Bearer ")
)

var (
	serviceLogRegExp = regexp.MustCompile("[a-z]+")

	checkLoc = func(name string) {
		if !serviceLogRegExp.MatchString(name) {
			panic("invalid name")
		}
	}

	ErrCantGetSessionFromContext = ServiceError{ErrMsg: "can't get session from context", HttpCode: http.StatusInternalServerError}
	ErrTokenNotFound             = ServiceError{ErrMsg: "no access token found, header format should be 'Authorization: Bearer '", HttpCode: http.StatusUnauthorized}
)

func GetCommentOfStruct(obj interface{}) string {
	return ""
}

func MustMarshalJson(obj any) string {
	desc, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return string(desc)
}

type Error interface {
	error
	HTTPCode() int
}

type ServiceError struct {
	ErrMsg   string
	HttpCode int
}

func (e ServiceError) Error() string {
	return e.ErrMsg
}

func (e ServiceError) HTTPCode() int {
	return e.HttpCode
}

type Request interface {
	Sanitize() Request
	Validate() error
}

type Services[SESSION any] struct {
	Name string
	Desc string

	Parent *gin.RouterGroup
	Group  *gin.RouterGroup

	Services   []*Service[SESSION]
	serviceMap map[string]*Service[SESSION]

	authFunc     func(ctx context.Context, token string) (SESSION, error)
	errConvertor func(error) Error
}

// NewServices create services with a default auth function.
func NewServices[SESSION any](name string, parent *gin.RouterGroup, desc string,
	authFunc func(ctx context.Context, token string) (SESSION, error), errHandler func(error) Error) *Services[SESSION] {
	if errHandler == nil {
		errHandler = func(err error) Error {
			if err, ok := err.(Error); ok {
				return err
			}
			return ServiceError{
				ErrMsg:   err.Error(),
				HttpCode: http.StatusInternalServerError,
			}
		}
	}
	return &Services[SESSION]{
		Name: name,
		Desc: desc,

		Parent: parent,
		Group:  parent.Group(name),

		Services:   nil,
		serviceMap: make(map[string]*Service[SESSION]),

		authFunc:     authFunc,
		errConvertor: errHandler,
	}
}

func (ss *Services[SESSION]) MustLookupService(name string) *Service[SESSION] {
	service, ok := ss.serviceMap[name]
	if !ok {
		panic(fmt.Sprintf("service '%s' not found", name))
	}
	return service
}

func (ss *Services[SESSION]) LookupService(name string) (*Service[SESSION], bool) {
	s, ok := ss.serviceMap[name]
	return s, ok
}

func (ss *Services[SESSION]) GetDescription() string {
	return ss.GetDescriptionEx("", "")
}

func (ss *Services[SESSION]) GetDescriptionEx(tab string, language string) string {
	if tab == "" {
		tab = "        "
	}
	if language == "" {
		language = "en"
	}
	sb := strings.Builder{}
	switch language {
	case "en":
		sb.WriteString(fmt.Sprintf("%s service:\n", ss.Desc))
	case "zh":
		sb.WriteString(fmt.Sprintf("%s??????:\n", ss.Desc))
	default:
		panic(fmt.Sprintf("not supported language '%s'", language))
	}
	for _, srv := range ss.Services {
		sb.WriteString(fmt.Sprintf("%s%s\n", tab, srv.GetDescription(language)))
	}
	return sb.String()
}

type Service[SESSION any] struct {
	Name   string
	loc    string
	Method string
	Desc   string

	Parent  *Services[SESSION]
	auth    gin.HandlerFunc
	handler gin.HandlerFunc

	exampleReqVal  interface{}
	exampleRespVal interface{}
}

func (it *Service[SESSION]) SetExampleReqValue(val interface{}) {
	it.exampleReqVal = val
}

func (it *Service[SESSION]) SetExampleRespValue(val interface{}) {
	it.exampleRespVal = val
}

func (it *Service[SESSION]) GetLoc() string {
	return it.Parent.Group.BasePath() + "/" + it.loc
}

func (it *Service[SESSION]) GenReqFormat() string {
	if it.exampleReqVal == nil {
		panic("it.exampleReqVal == nil")
	}
	return MustMarshalJson(it.exampleReqVal)
}

func (it *Service[SESSION]) GenRespFormat() (respFormat string) {
	if it.exampleRespVal == nil {
		panic("it.exampleRespVal == nil")
	}

	respType := reflect.ValueOf(it.exampleRespVal).Type()
	respKind := respType.Kind()
	if respKind != reflect.Slice {
		return MustMarshalJson(it.exampleRespVal)
	}

	respFormat = "["
	respEleType := respType.Elem()
	respEleVal := reflect.Zero(respEleType).Interface()
	respFormat += MustMarshalJson(respEleVal)
	respFormat += "]"
	return respFormat
}

func (it *Service[SESSION]) String() string {
	return it.GetDescription("en")
}

func (it *Service[SESSION]) GetDescription(language string) string {
	switch language {
	case "en":
		authDesc := "Yes"
		if it.auth == nil {
			authDesc = "No"
		}
		return fmt.Sprintf(
			"ServiceName: \"%s\", NeedAuth: %s, ReqURL: %s,"+
				" ReqFormat: %s%s,"+
				" RespFormat: %s%s",
			it.Desc, authDesc, it.GetLoc(),
			it.GenReqFormat(), GetCommentOfStruct(it.exampleReqVal),
			it.GenRespFormat(), GetCommentOfStruct(it.exampleRespVal))
	case "zh":
		authDesc := "???"
		if it.auth == nil {
			authDesc = "???"
		}
		return fmt.Sprintf(
			"?????????: \"%s\", ????????????: %s, ????????????: %s,"+
				" ????????????: %s%s,"+
				" ????????????: %s%s",
			it.Desc, authDesc, it.GetLoc(),
			it.GenReqFormat(), GetCommentOfStruct(it.exampleReqVal),
			it.GenRespFormat(), GetCommentOfStruct(it.exampleRespVal))
	default:
		panic(fmt.Sprintf("not supported language '%s'", language))
	}
}

// RegisterService register a service
// ss Services
// name service name
// loc service location (the url)
// method http method
// desc description
// serviceFunc your own service implementation
func RegisterService[REQ Request, RESP any, SESSION any](
	ss *Services[SESSION],
	name string, loc string, method string, desc string,
	serviceFunc func(ctx context.Context, req REQ) (RESP, error)) *Service[SESSION] {
	if _, ok := ss.LookupService(name); ok {
		panic(fmt.Sprintf("Service '%s' already exists", name))
	}

	var (
		req  REQ
		resp RESP
	)
	srv := &Service[SESSION]{
		Name:   name,
		loc:    genLoc(name, loc),
		Method: method,
		Desc:   desc,

		Parent: ss,
		handler: func(c *gin.Context) {
			process(c, ss.errConvertor, serviceFunc)
		},

		exampleReqVal:  req,
		exampleRespVal: resp,
	}
	srv.register()
	ss.serviceMap[name] = srv
	ss.Services = append(ss.Services, srv)
	return srv
}

// RegisterAuthenticatedService register a service which needs authentication
// ss Services
// name service name
// loc service location (the url)
// method http method
// desc description
// serviceFunc your own service implementation
func RegisterAuthenticatedService[REQ Request, RESP any, SESSION any](
	ss *Services[SESSION],
	name string, loc string, method string, desc string,
	serviceFunc func(ctx context.Context, req REQ, session SESSION) (RESP, error)) *Service[SESSION] {
	if _, ok := ss.LookupService(name); ok {
		panic(fmt.Sprintf("Service '%s' already exists", name))
	}

	var (
		req  REQ
		resp RESP
	)
	srv := &Service[SESSION]{
		Name:   name,
		loc:    genLoc(name, loc),
		Method: method,
		Desc:   desc,

		Parent: ss,
		auth: func(c *gin.Context) {
			auth(c, ss)
		},
		handler: func(c *gin.Context) {
			processWithSession(c, ss.errConvertor, serviceFunc)
		},

		exampleReqVal:  req,
		exampleRespVal: resp,
	}
	srv.register()
	ss.serviceMap[name] = srv
	ss.Services = append(ss.Services, srv)
	return srv
}

func (it *Service[SESSION]) register() {
	switch it.Method {
	case "POST":
		if it.auth != nil {
			it.Parent.Group.POST("/"+it.loc, it.auth, it.handler)
		} else {
			it.Parent.Group.POST("/"+it.loc, it.handler)
		}
	case "GET":
		if it.auth != nil {
			it.Parent.Group.GET("/"+it.loc, it.auth, it.handler)
		} else {
			it.Parent.Group.GET("/"+it.loc, it.handler)
		}
	default:
		panic(fmt.Sprintf("unknown method '%s'", it.Method))
	}
}

func process[REQ Request, RESP any](
	ctx *gin.Context,
	errConvertor func(error) Error,
	handler func(ctx context.Context, req REQ) (result RESP, _ error)) {
	_, _, _ = func(ctx *gin.Context,
		handler func(ctx context.Context, req REQ) (result RESP, _ error)) (resp RESP, httpCode int, err error) {
		const (
			httpCodeUnknown = 0
		)

		defer func() {
			if err != nil {
				if err := errConvertor(err); httpCode != httpCodeUnknown {
					ctx.AbortWithStatusJSON(httpCode, err)
				} else {
					ctx.AbortWithStatusJSON(err.HTTPCode(), err)
				}
				return
			}

			ctx.JSON(http.StatusOK, resp)
		}()

		var (
			req REQ
		)
		if err := ctx.ShouldBind(&req); err != nil {
			return resp, http.StatusBadRequest, err
		}

		req = req.Sanitize().(REQ)
		if err := req.Validate(); err != nil {
			return resp, http.StatusBadRequest, err
		}

		if resp, err = handler(ctx, req); err != nil {
			return resp, httpCodeUnknown, err
		}
		return resp, http.StatusOK, nil
	}(ctx, handler)
}

func auth[SESSION any](ctx *gin.Context, ss *Services[SESSION]) {
	_, _, _ = func(ctx *gin.Context) (session SESSION, httpCode int, err error) {
		const (
			httpCodeUnknown = 0
		)

		defer func() {
			if err != nil {
				if err := ss.errConvertor(err); httpCode != httpCodeUnknown {
					ctx.AbortWithStatusJSON(httpCode, err)
				} else {
					ctx.AbortWithStatusJSON(err.HTTPCode(), err)
				}
				return
			}

			ctx.Set(SessionCtxKey, session)
			ctx.Next()
		}()

		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) <= bearerLength {
			var session SESSION
			return session, http.StatusBadRequest, ErrTokenNotFound
		}
		token := strings.TrimSpace(authHeader[bearerLength:])
		if session, err = ss.authFunc(ctx, token); err != nil {
			var session SESSION
			return session, http.StatusUnauthorized, err
		}
		return session, http.StatusOK, nil
	}(ctx)
}

func processWithSession[REQ Request, RESP any, SESSION any](
	ctx *gin.Context,
	errConvertor func(error) Error,
	handler func(context.Context, REQ, SESSION) (result RESP, _ error)) {
	_, _, _ = func(ctx *gin.Context,
		handler func(ctx context.Context, req REQ, session SESSION) (RESP, error)) (resp RESP, httpCode int, err error) {
		const (
			httpCodeUnknown = 0
		)
		defer func() {
			if err != nil {
				if err := errConvertor(err); httpCode != httpCodeUnknown {
					ctx.AbortWithStatusJSON(httpCode, err)
				} else {
					ctx.AbortWithStatusJSON(err.HTTPCode(), err)
				}
				return
			}

			ctx.JSON(http.StatusOK, resp)
		}()

		var (
			req REQ
		)
		if err := ctx.ShouldBind(&req); err != nil {
			return resp, http.StatusBadRequest, err
		}

		req = req.Sanitize().(REQ)
		if err := req.Validate(); err != nil {
			return resp, http.StatusBadRequest, err
		}

		obj, ok := ctx.Get(SessionCtxKey)
		if !ok {
			return resp, http.StatusInternalServerError, ErrCantGetSessionFromContext
		}
		if resp, err = handler(ctx, req, obj.(SESSION)); err != nil {
			var resp RESP
			return resp, httpCodeUnknown, err
		}
		return resp, http.StatusOK, nil
	}(ctx, handler)
}

func genLoc(name string, loc string) string {
	if loc == "" {
		loc = strings.ToLower(name)
		checkLoc(loc)
	}
	return loc
}
