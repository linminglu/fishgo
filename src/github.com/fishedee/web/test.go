package web

import (
	"fmt"
	. "github.com/fishedee/language"
	"math/rand"
	"os"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type TestInterface interface {
	ControllerInterface
}

type Test struct {
	Controller
}

func (this *Test) getTraceLineNumber(traceNumber int) string {
	_, filename, line, ok := runtime.Caller(traceNumber + 1)
	if !ok {
		return "???.go:???"
	} else {
		return path.Base(filename) + ":" + strconv.Itoa(line)
	}
}

func (this *Test) Concurrent(number int, concurrency int, handler func()) {
	if number <= 0 {
		panic("benchmark numer is invalid")
	}
	if concurrency <= 0 {
		panic("benchmark concurrency is invalid")
	}
	singleConcurrency := number / concurrency
	if singleConcurrency <= 0 ||
		number%concurrency != 0 {
		panic("benchmark numer/concurrency is invalid")
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			for i := 0; i < singleConcurrency; i++ {
				handler()
			}
		}()
	}
	wg.Wait()
}

func (this *Test) Benchmark(number int, concurrency int, handler func(), testCase ...interface{}) {
	beginTime := time.Now().UnixNano()
	this.Concurrent(number, concurrency, handler)
	endTime := time.Now().UnixNano()

	totalTime := endTime - beginTime
	singleTime := totalTime / int64(number)
	singleReq := float64(number) / (float64(totalTime) / 1e9)
	if len(testCase) == 0 {
		testCase = []interface{}{""}
	}
	traceInfo := this.getTraceLineNumber(1)
	fmt.Printf(
		"%v: %v number %v concurrency %v / req, %.2freq / s\n",
		traceInfo,
		number,
		concurrency,
		time.Duration(singleTime).String(),
		singleReq,
	)
}

func (this *Test) AssertEqual(left interface{}, right interface{}, testCase ...interface{}) {
	errorString, isEqual := DeepEqual(left, right)
	if isEqual {
		return
	}
	traceInfo := this.getTraceLineNumber(1)
	if len(testCase) == 0 {
		this.Ctx.Testing.Errorf("%v: assertEqual Fail! %v", traceInfo, errorString)
	} else {
		this.Ctx.Testing.Errorf("%v:%v: assertEqual Fail! %v", traceInfo, testCase[0], errorString)
	}
}

func (this *Test) AssertError(left Exception, rightCode int, rightMessage string, testCase ...interface{}) {
	errorString := ""
	if left.GetCode() != rightCode {
		errorString = fmt.Sprintf("assertError Code Fail! %v != %+v ", left.GetCode(), rightCode)
	}
	if left.GetMessage() != rightMessage {
		errorString = fmt.Sprintf("assertError Message Fail! %v != %v ", left.GetMessage(), rightMessage)
	}
	if errorString == "" {
		return
	}
	traceInfo := this.getTraceLineNumber(1)
	if len(testCase) == 0 {
		this.Ctx.Testing.Errorf("%v: %v", traceInfo, errorString)
	} else {
		this.Ctx.Testing.Errorf("%v:%v: %v", traceInfo, testCase[0], errorString)
	}

}

func (this *Test) RandomInt() int {
	return rand.Int()
}

func (this *Test) RandomString(length int) string {
	result := []rune{}
	for i := 0; i != length; i++ {
		var single rune
		randInt := rand.Int() % (10 + 26 + 26)
		if randInt < 10 {
			single = rune('0' + randInt)
		} else if randInt < 10+26 {
			single = rune('A' + (randInt - 10))
		} else {
			single = rune('a' + (randInt - 10 - 26))
		}
		result = append(result, single)
	}
	return string(result)
}

func (this *Test) RequestReset() {
	this.InitEmpty(this.appController, this.Ctx.Testing)
}

var testMap map[string][]TestInterface
var testMapInit bool

func init() {
	testMap = map[string][]TestInterface{}
	testMapInit = false
}

func runSingleTest(t *testing.T, test TestInterface) {
	//初始化test
	test.InitEmpty(test, t)

	isBenchTest := false
	for _, singleArgv := range os.Args {
		if strings.Index(singleArgv, "bench") != -1 {
			isBenchTest = true
		}
	}

	//遍历test，执行测试
	testType := reflect.TypeOf(test)
	testValue := reflect.ValueOf(test)
	testMethodNum := testType.NumMethod()
	for i := 0; i != testMethodNum; i++ {
		singleValueMethodType := testType.Method(i)
		if isBenchTest == false {
			if strings.HasPrefix(singleValueMethodType.Name, "Test") == false {
				continue
			}
		} else {
			if strings.HasPrefix(singleValueMethodType.Name, "Benchmark") == false ||
				singleValueMethodType.Name == "Benchmark" {
				continue
			}
		}
		//执行测试
		singleValueMethodType.Func.Call([]reflect.Value{testValue})
	}
}

func RunTest(t *testing.T, data interface{}) {
	//获取package
	pkgPath := reflect.TypeOf(data).Elem().PkgPath()

	//初始化runtime
	if testMapInit == false {
		runtime.GOMAXPROCS(runtime.NumCPU() * 4)
		rand.Seed(time.Now().Unix())
		testMapInit = true
	}

	//遍历测试
	for _, singleTest := range testMap[pkgPath] {
		runSingleTest(t, singleTest)
	}
}

func InitTest(test TestInterface) {
	pkgPath := reflect.TypeOf(test).Elem().PkgPath()
	testMap[pkgPath] = append(testMap[pkgPath], test)
}
