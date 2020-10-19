package pkg

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"
)

var l = log.New(os.Stdout, "TESTING: ", log.Lshortfile)

func TestWalking(t *testing.T) {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	testdata := path.Join(mydir, "testdata")
	getDateInName, getDateInExif := walkingOnFilesystem(testdata, l)
	fmt.Println(getDateInName, getDateInExif)
	wantDateInName := 1
	wantDateInExif := 2
	if len(getDateInExif) != wantDateInExif || len(getDateInName) != wantDateInName {
		t.Error("some wrong")
	}
}
