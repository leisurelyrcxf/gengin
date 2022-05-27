HTTP Server:  
&nbsp;&nbsp;&nbsp;&nbsp;`./main`  
&nbsp;&nbsp;&nbsp;&nbsp;`curl -X POST http://localhost:8080/v1/usr/signin -H 'Content-Type: application/json' -d '{"Email": "john@gmail.com", "Password": "123456"}'`  
&nbsp;&nbsp;&nbsp;&nbsp;`curl -X GET http://localhost:8080/v1/usr/profile -H 'Authorization: bearer 1'`

Print Interface:  
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;`./desc`  
&nbsp;&nbsp;You will get:
<pre> 
User service:  
    ServiceName: "login", NeedAuth: No, ReqURL: /v1/usr/signin, ReqFormat: {"Email":"","Password":""}, RespFormat: {"_t":""}   
    ServiceName: "get user profile", NeedAuth: Yes, ReqURL: /v1/usr/profile, ReqFormat: {}, RespFormat: {"ID":0,"Phone":"","Email":""}
</pre> 
