package jdprice

import (
	"fmt"
	"regexp"

	"github.com/beego/beego/v2/adapter/httplib"
	"github.com/buger/jsonparser"
	"github.com/cdle/sillyGirl/core"
	"github.com/gin-gonic/gin"
)

func init() {
	otto := core.Bucket("otto")
	core.Server.GET("/jdprice/:sku", func(c *gin.Context) {
		sku := core.Int(c.Param("sku"))
		if sku == 0 {
			c.String(404, "巴嘎")
			return
		}
		req := httplib.Get("https://api.jingpinku.com/get_rebate_link/api?" +
			"appid=" + otto.Get("jingpinku_appid") +
			"&appkey=" + otto.Get("jingpinku_appkey") +
			"&union_id=" + otto.Get("jd_union_id") +
			"&content=" + fmt.Sprintf("https://item.jd.com/%d.html", sku))
		data, err := req.Bytes()
		if err != nil {
			c.String(404, err.Error())
			return
		}
		code, _ := jsonparser.GetInt(data, "code")
		if code != 0 {
			msg, _ := jsonparser.GetString(data, "msg")
			c.String(404, msg)
			return
		}
		official, _ := jsonparser.GetString(data, "official")
		if official == "" {
			c.String(404, "暂无商品信息。")
			return
		}
		image, _ := jsonparser.GetString(data, "images", "[0]")
		var price string = ""
		if res := regexp.MustCompile(`京东价：(.*)\n`).FindStringSubmatch(official); len(res) > 0 {
			price = res[1]
		}
		c.JSON(200, map[string]interface{}{
			"official": official,
			"price":    price,
			"image":    image,
		})
	})
}
