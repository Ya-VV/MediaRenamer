package pkg

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"
)

var l log.Logger

// log.SetOutput(os.Stdout)

func TestWalking(t *testing.T) {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	testdata := path.Join(mydir, "testdata")
	t.Error(testdata)

}
