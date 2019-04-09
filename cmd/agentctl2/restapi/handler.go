package restapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/ligato/cn-infra/db/keyval/etcd"
	"github.com/ligato/vpp-agent/cmd/agentctl2/utils"
)

func GetLog(endpoints []string, path string) string {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed to read config - "+err.Error()))
	}

	//TODO: Not nice solution
	addr := etcdConfig.Config.Endpoints[0]

	addrPath := "http://" + addr + path
	fmt.Printf("%s\n", addrPath)
	resp, err := http.Get(addrPath)
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get http request - "+err.Error()))
	}

	msg, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed get message body - "+err.Error()))
	}

	return string(msg)
}

func SetLog(endpoints []string, path string) {
	if len(endpoints) > 0 {
		ep := strings.Join(endpoints, ",")
		os.Setenv("ETCD_ENDPOINTS", ep)
	}

	cfg := &etcd.Config{}
	etcdConfig, err := etcd.ConfigToClient(cfg)

	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed to read config - "+err.Error()))
	}

	//TODO: Not nice solution
	addr := etcdConfig.Config.Endpoints[0]
	addrPath := "http://" + addr + path

	fmt.Printf("%s\n", addrPath)
	client := http.Client{}
	request, err := http.NewRequest(http.MethodPut, addrPath, strings.NewReader(""))
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed create put message body - "+err.Error()))
	}

	response, err := client.Do(request)
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed send put message body - "+err.Error()))
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if nil != err {
		utils.ExitWithError(utils.ExitError,
			errors.New("Failed receiver answer to put message body - "+err.Error()))
	}
	fmt.Printf("%s\n", contents)
}
