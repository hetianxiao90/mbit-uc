package nacos

import (
	"bytes"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddr string
	Namespace  string
	DataId     string
	Group      string
}

func NewNacosConfig(serverAddr string, namespace, dataId, group string) *Config {
	return &Config{
		ServerAddr: serverAddr,
		Namespace:  namespace,
		DataId:     dataId,
		Group:      group,
	}
}

func (n *Config) GetConfig() (string, error) {

	// 拼接nacos配置
	var serverConfigs []constant.ServerConfig
	values := strings.Split(n.ServerAddr, ",")
	for _, v := range values {
		vs := strings.Split(v, ":")
		if len(vs) != 2 {
			continue
		}
		port, _ := strconv.ParseInt(vs[1], 10, 64)
		serverConfigs = append(serverConfigs, *constant.NewServerConfig(vs[0], uint64(port)))
	}

	clientConfig := *constant.NewClientConfig(
		constant.WithNamespaceId(n.Namespace),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel("warn"),
	)

	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)

	if err != nil {
		return "", err
	}

	// 获取配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: n.DataId,
		Group:  n.Group,
	})
	if err != nil {
		return "", err
	}
	go func() {
		err = client.ListenConfig(vo.ConfigParam{
			DataId: n.DataId,
			Group:  n.Group,
			OnChange: func(namespace, group, dataId, data string) {
				dataByte := []byte(data)
				if err = viper.MergeConfig(bytes.NewBuffer(dataByte)); err != nil {
					fmt.Printf("viper MergeConfig err: %v", err)
				}
			},
		})
	}()
	return content, nil
}
