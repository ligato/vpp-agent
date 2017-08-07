package main

import (
	"go/build"
	"log"
	"os"
	"os/exec"

	"github.com/taylorchu/generic"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 3 {
		log.Fatalln("generic [SRCPATH] [DEST] [TypeXXX->OtherType]...")
	}

	if os.Args[1] == "" {
		log.Fatalln("SRCPATH cannot be empty")
	}

	if os.Args[2] == "" {
		log.Fatalln("DEST cannot be empty")
	}

	ctx, err := generic.NewContext(os.Args[1], os.Args[2], os.Args[3:]...)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err := build.Import(os.Args[1], ".", build.FindOnly); err != nil {
		cmd := exec.Command("go", "get", "-u", os.Args[1])
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Fatalln(err)
		}
	}

	err = generic.RewritePackage(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
