package utils

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func InitServer(ctx context.Context, rdb *redis.Client) *gin.Engine {
	r := gin.Default()

	// 获取容器
	r.POST("/code/get", func(c *gin.Context) {
		username, ue := c.GetPostForm("username")
		school, se := c.GetPostForm("school")
		nameReg, _ := regexp.MatchString("^[a-zA-Z0-9]+$", username)
		schoolReg, _ := regexp.MatchString("^[a-zA-Z0-9]+$", school)
		if !(ue && se) || !nameReg || !schoolReg {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "params error",
			})
		} else {
			token := username + school
			port, err := GetKey("shadow:"+token, rdb, ctx)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"msg": port,
				})
			} else {
				// 新建端口
				port, err := GetPortFromPool(rdb, ctx)
				tmp := strconv.Itoa(port)
				fmt.Println(tmp)
				if err != nil || port <= 0 {
					c.JSON(http.StatusInternalServerError, gin.H{
						"msg": err.Error(),
					})
				} else {
					// err := AddContainerByPortNumber(port, token, ctx, rdb)
					err := AddContainer(port, token, ctx, rdb)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"msg": err.Error(),
						})
					} else {
						AddEx(token, strconv.Itoa(port), rdb, ctx)
						AddShadow("shadow:"+token, strconv.Itoa(port), rdb, ctx)
						c.JSON(http.StatusOK, gin.H{
							"msg": port,
						})
					}
				}
			}
		}
	})

	r.POST("/code/delete", func(c *gin.Context) {
		username, ue := c.GetPostForm("username")
		school, se := c.GetPostForm("school")
		if !(ue && se) {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": "params error",
			})
		} else {
			token := username + school
			val, err := GetKey("shadow:"+token, rdb, ctx)
			if err == nil {
				fmt.Println("removing key: " + token)
				rdb.Del(ctx, "shadow:"+token, token)
				fmt.Println("removing code: " + token)
				DelContainerByName(token, ctx, rdb)
				fmt.Println("releasing port: " + val)
				DelPort(val, rdb, ctx)
				c.String(http.StatusOK, "资源被释放")
			}
		}
	})

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "OK",
		})
	})

	return r
}
