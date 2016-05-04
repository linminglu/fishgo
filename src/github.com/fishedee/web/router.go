package web

import (
	"bytes"
	"github.com/fishedee/language"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type methodInfo struct {
	viewName       string
	controllerType reflect.Type
	methodType     reflect.Method
}

type handlerType struct {
	routerControllerMethod map[string]methodInfo
}

func (this *handlerType) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	beginTime := time.Now().UnixNano()
	this.handleBeegoRequest(request, response)
	endTime := time.Now().UnixNano()
	globalBasic.Log.Debug("%s %s : %s", request.Method, request.URL.String(), time.Duration(endTime-beginTime).String())
}

func (this *handlerType) firstLowerName(name string) string {
	return strings.ToLower(name[0:1]) + name[1:]
}

func (this *handlerType) firstUpperName(name string) string {
	return strings.ToUpper(name[0:1]) + name[1:]
}

func (this *handlerType) isPublic(name string) bool {
	fisrtStr := name[0:1]
	if fisrtStr >= "A" && fisrtStr <= "Z" {
		return true
	} else {
		return false
	}
}

func (this *handlerType) addRoute(namespace string, target interface{}) {
	if this.routerControllerMethod == nil {
		this.routerControllerMethod = map[string]methodInfo{}
	}
	controllerType := reflect.TypeOf(target)
	numMethod := controllerType.NumMethod()
	for i := 0; i != numMethod; i++ {
		singleMethod := controllerType.Method(i)
		singleMethodName := singleMethod.Name
		if this.isPublic(singleMethodName) == false {
			continue
		}
		methodNameInfo := language.Explode(singleMethodName, "_")
		if len(methodNameInfo) < 2 {
			continue
		}
		namespace := strings.Trim(namespace, "/")
		methodName[0] = strings.Trim(methodName[0], "/")
		url := namespace + "/" + methodName[0]
		routerControllerMethod[url] = &methodInfo{
			viewName:       firstLowerName(methodNameInfo[1]),
			controllerType: controllerType,
			methodType:     singleMethod,
		}
	}
}

func (this *handlerType) handleRequest(request *http.Request, response http.ResponseWriter) {
	//查找路由
	url := ctx.Input.URL()
	url = strings.Trim(url, "/")
	method, isExist := this.routerControllerMethod[url]
	if isExist == false {
		ctx.Abort("404", "File Not Found")
		return
	}

	//执行路由
	controller := reflect.New(method.controllerType)
	this.runRequest(controller, method, request, response)
}

func (this *handlerType) runRequest(controller reflect.Value, method methodInfo, request *http.Request, response http.ResponseWriter) {
	urlMethod := request.Method()
	target := controller.Interface().(ControllerInterface)
	defer language.CatchCrash(func(exception language.Exception) {
		target.GetBasic().Log.Critical("Buiness Crash Code:[%d] Message:[%s]\nStackTrace:[%s]", exception.GetCode(), exception.GetMessage(), exception.GetStackTrace())
	})
	target.Init(controller, request, response, nil)
	var controllerResult interface{}
	if urlMethod == "GET" || urlMethod == "POST" ||
		urlMethod == "DELETE" || urlMethod == "PUT" {
		result := this.runRequestBusiness(target, method.methodType, []reflect.Value{controller})
		if len(result) != 1 {
			panic("url controller should has return value " + url)
		}
		controllerResult = result[0].Interface()
	} else {
		controllerResult = nil
	}
	target.AutoRender(controllerResult, method.viewName)
}

func (this *handlerType) runRequestBusiness(target ControllerInterface, method reflect.Value, arguments []reflect.Value) (result []reflect.Value) {
	defer language.Catch(func(exception language.Exception) {
		target.GetBasic().Log.Error("Buiness Error Code:[%d] Message:[%s]\nStackTrace:[%s]", exception.GetCode(), exception.GetMessage(), exception.GetStackTrace())
		result = []reflect.Value{reflect.ValueOf(exception)}
	})
	result = method.Call(arguments)
	if len(result) == 0 {
		result = []reflect.Value{reflect.Zero(reflect.TypeOf(ControllerInterface))}
	}
	return
}

var handler handlerType

func InitRoute(namespace string, target interface{}) {
	handler.addRoute(namespace, target)
}

func Run() error {
	httpPort := globalBasic.Config.GetInt("httpport")
	if httpPort == 0 {
		httpPort = 8080
	}
	globalBasic.Log.Debug("Server is Running :%v", httpPort)
	err := http.ListenAndServe(":"+strings.Atoi(httpPort), &handler)
	if err != nil {
		globalBasic.Log.Error("Listen fail! " + err.Error())
		return err
	}
	return nil
}