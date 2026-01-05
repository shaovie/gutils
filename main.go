package main

import (
	"fmt"
	"time"

	"github.com/shaovie/gutils/gutils"
	"github.com/shaovie/gutils/ihttp"
	"github.com/shaovie/gutils/ilog"
	"github.com/shaovie/gutils/isqlite"
)

func main() {
	code, _, _ := ihttp.Get("https://g.cn", 1*time.Second, nil)
	fmt.Printf("code = %d\n", code)
	ilog.Init("")
	isqlite.New("", "")
	fmt.Printf("random string = %s\n", gutils.RandomStr(12))
}
