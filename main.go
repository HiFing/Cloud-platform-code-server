package main

import (
	. "code/utils"
	"context"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/go-redis/redis/v8"
)

var ctx context.Context
var rdb *redis.Client

func main() {
	ctx = context.Background()

	rdb = redis.NewClient(&redis.Options{
		Addr:     "http://localhost:6379",
		Password: "",
		DB:       6,
	})

	//todo 要在redis里面执行CONFIG set notify-keyspace-events EKx
	pubsub := rdb.PSubscribe(ctx, "__keyevent@6__:expired")

	go releaseContainer(pubsub)

	clientConfig := constant.ClientConfig{
		//NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468", //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./log",
		CacheDir:            "./cache",
		LogLevel:            "debug",
		Username:            "nacos",
		Password:            "nacos",
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	// Another way of create naming client for service discovery (recommend)
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		fmt.Println(err)
	}

	// Another way of create config client for dynamic configuration (recommend)
	//configClient, _ := clients.NewConfigClient(
	//	vo.NacosClientParam{
	//		ClientConfig:  &clientConfig,
	//		ServerConfigs: serverConfigs,
	//	},
	//)

	_, err = namingClient.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          "localhost",
		Port:        8888,
		ServiceName: "ace-container",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		//Metadata:    map[string]string{"idc": "shanghai"},
		//ClusterName: "cluster-a", // default value is DEFAULT
		//GroupName:   "group-a",   // default value is DEFAULT_GROUP
	})
	if err != nil {
		fmt.Println(err)
	}

	//dataId := "gin.json"
	//group := "gin"
	//content, err := configClient.GetConfig(vo.ConfigParam{
	//	DataId: dataId,
	//	Group:  group})

	r := InitServer(ctx, rdb)

	r.Run(":8888")

	defer pubsub.Close()
}

// 释放过期资源(key\shadow\code\Port记录)
func releaseContainer(pubsub *redis.PubSub) {
	for {
		msg := <-pubsub.Channel()
		token := msg.Payload
		val, err := GetKey("shadow:"+token, rdb, ctx)
		if err == nil {
			fmt.Println("removing key: " + token)
			rdb.Del(ctx, "shadow:"+token, token)
			fmt.Println("removing code: " + token)
			DelContainerByName(token, ctx, rdb)
			fmt.Println("releasing port: " + val)
			DelPort(val, rdb, ctx)
		}
	}
}
