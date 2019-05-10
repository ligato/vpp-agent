package restapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/vpp-agent/cmd/agentctl/utils"
)

func GetMsg(endpoints []string, path string) string {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed to read config - "+err.Error()))
	}

	//TODO: Not nice solution
	addr := etcdConfig.Config.Endpoints[0]

	addrPath := "http://" + addr + path
	resp, err := http.Get(addrPath)
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get http request - "+err.Error()))
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get message body - "+err.Error()))
	}

	return string(msg)
}

func SetMsg(endpoints []string, path string) {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed to read config - "+err.Error()))
	}

	//TODO: Not nice solution
	addr := etcdConfig.Config.Endpoints[0]
	addrPath := "http://" + addr + path

	client := http.Client{}
	request, err := http.NewRequest(http.MethodPut, addrPath, strings.NewReader(""))
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed create put message body - "+err.Error()))
	}

	response, err := client.Do(request)
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed send put message body - "+err.Error()))
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed receiver answer to put message body - "+err.Error()))
	}
	fmt.Printf("%s\n", contents)
}

func PostMsg(endpoints []string, path string, jsonData string) string {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed to read config - "+err.Error()))
	}

	//TODO: Not nice solution
	addr := etcdConfig.Config.Endpoints[0]

	addrPath := "http://" + addr + path
	resp, err := http.Post(addrPath, "application/json", strings.NewReader(jsonData))
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get http request - "+err.Error()))
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get message body - "+err.Error()))
	}

	return string(msg)
}
