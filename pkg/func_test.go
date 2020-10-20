package pkg

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"
)

var l = log.New(os.Stdout, "TESTING: ", log.Lshortfile)

func TestCheckWorkDir(t *testing.T) {
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	workDir = path.Join(mydir, "testdata")
	checkWorkDir(l)
}

func TestWalking(t *testing.T) {
	exiftoolExist = true
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	testdata := path.Join(mydir, "testdata")
	getDateInName, getDateInExif := walkingOnFilesystem(testdata, l)
	fmt.Println(getDateInName, getDateInExif)
	wantDateInName := 1
	wantDateInExif := 2
	if len(getDateInExif) != wantDateInExif {
		t.Error("Get ", len(getDateInExif), "in DateInExif slice, but want: ", wantDateInExif)
	} else if len(getDateInName) != wantDateInName {
		t.Error("Get ", len(getDateInName), "in DateInExif slice, but want: ", wantDateInName)
	}
}
