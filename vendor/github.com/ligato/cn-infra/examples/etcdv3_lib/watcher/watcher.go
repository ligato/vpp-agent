package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/ligato/cn-infra/config"
	"github.com/ligato/cn-infra/datasync"
	"github.com/ligato/cn-infra/db/keyval"
	"github.com/ligato/cn-infra/db/keyval/etcdv3"
	"github.com/ligato/cn-infra/db/keyval/kvproto"
	"github.com/ligato/cn-infra/examples/etcdv3_lib/model/phonebook"
	"github.com/ligato/cn-infra/logging/logroot"
)

func processArgs() (*etcdv3.ClientConfig, error) {
	fileConfig := &etcdv3.Config{}
	if len(os.Args) > 2 {
		if os.Args[1] == "--cfg" {

			err := config.ParseConfigFromYamlFile(os.Args[2], fileConfig)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("incorrect arguments")
		}
	}

	return etcdv3.ConfigToClientv3(fileConfig)
}

func printUsage() {
	fmt.Printf("\n\n%s: [--cfg CONFIG_FILE] <delete NAME | put NAME COMPANY PHONE>\n\n", os.Args[0])
}

func printContact(c *phonebook.Contact) {
	fmt.Printf("\t%s\n\t\t%s\n\t\t%s\n", c.Name, c.Company, c.Phonenumber)
}

func main() {

	cfg, err := processArgs()
	if err != nil {
		printUsage()
		fmt.Println(err)
		os.Exit(1)
	}

	//create connection to etcd
	broker, err := etcdv3.NewEtcdConnectionWithBytes(*cfg, logroot.StandardLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//initialize proto decorator
	protoBroker := kvproto.NewProtoWrapper(broker)

	respChan := make(chan keyval.ProtoWatchResp, 0)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	err = protoBroker.Watch(keyval.ToChanProto(respChan), phonebook.EtcdPath())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Watching the key: ", phonebook.EtcdPath())

watcherLoop:
	for {
		select {
		case resp := <-respChan:
			switch resp.GetChangeType() {
			case datasync.Put:
				contact := &phonebook.Contact{}
				fmt.Println("Creating ", resp.GetKey())
				resp.GetValue(contact)
				printContact(contact)
			case datasync.Delete:
				fmt.Println("Removing ", resp.GetKey())
			}
			fmt.Println("============================================")
		case <-sigChan:
			break watcherLoop
		}
	}
	fmt.Println("Stop requested ...")
	protoBroker.Close()
}
