package jdprice

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/adapter/httplib"
	"github.com/buger/jsonparser"
	"github.com/cdle/sillyGirl/core"
	"github.com/gin-gonic/gin"
)

func init() {
	otto := core.Bucket("otto")
	core.OttoFuncs["jdprice"] = func(str string) string {
		sku := core.Int(str)
		if sku == 0 {
			return `{}`
		}
		req := httplib.Get("https://api.jingpinku.com/get_rebate_link/api?" +
			"appid=" + otto.Get("jingpinku_appid") +
			"&appkey=" + otto.Get("jingpinku_appkey") +
			"&union_id=" + otto.Get("jd_union_id") +
			"&content=" + fmt.Sprintf("https://item.jd.com/%d.html", sku))
		data, err := req.Bytes()
		if err != nil {
			return `{}`
		}
		short, _ := jsonparser.GetString(data, "content")
		code, _ := jsonparser.GetInt(data, "code")
		if code != 0 {
			// msg, _ := jsonparser.GetString(data, "msg")
			return `{}`
		}
		official, _ := jsonparser.GetString(data, "official")
		if official == "" {
			return `{}`
		}
		lines := strings.Split(official, "\n")
		official = ""
		title := ""
		for i, line := range lines {
			if i == 0 {
				title = strings.Trim(regexp.MustCompile("【.*?】").ReplaceAllString(line, ""), " ")
			}
			if !strings.Contains(line, "佣金") {
				official += line + "\n"
			}
		}
		official = strings.Trim(official, "\n")
		image, _ := jsonparser.GetString(data, "images", "[0]")
		var price string = ""
		var final string = ""
		if res := regexp.MustCompile(`京东价：(.*)\n`).FindStringSubmatch(official); len(res) > 0 {
			price = res[1]
		}
		if res := regexp.MustCompile(`促销价：(.*)\n`).FindStringSubmatch(official); len(res) > 0 {
			final = res[1]
		}
		if math.Abs(core.Float64(price)-core.Float64(final)) < 0.1 {
			final = price
		} else {
			req := httplib.Get("https://api.jingpinku.com/get_powerful_coup_link/api?" +
				"appid=" + otto.Get("jingpinku_appid") +
				"&appkey=" + otto.Get("jingpinku_appkey") +
				"&union_id=" + otto.Get("jd_union_id") +
				"&content=" + fmt.Sprintf("https://item.jd.com/%d.html", sku))
			data, _ := req.Bytes()
			quan, _ := jsonparser.GetString(data, "content")
			if strings.Contains(quan, "https://u.jd.com") {
				short = quan
			}
		}
		data, _ = json.Marshal(map[string]interface{}{
			"title":    title,
			"short":    short,
			"official": official,
			"price":    price,
			"final":    final,
			"image":    image,
		})
		return string(data)
	}
	core.Server.GET("/jdprice/:sku", func(c *gin.Context) {
		sku := c.Param("sku")

		c.String(200, core.OttoFuncs["jdprice"](sku))
	})
}
