package imperator

import (
	"fmt"
	"regexp"
	"runtime"
	"time"
)

func (i *Imperator) LoadTime(start time.Time) {
	elapsed := time.Since(start)
	caller, _, _, _ := runtime.Caller(1)
	funcObj := runtime.FuncForPC(caller)
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")
	i.InfoLog.Println(fmt.Sprintf("Load time: %s took %s", name, elapsed))
}
