package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/barasher/go-exiftool"
)

func main() {
	timeNow := time.Now()
	logFile, err := os.OpenFile(timeNow.Format("20060102150405")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(mw, "INFO: ", log.Flags()&^(log.Ldate|log.Ltime))
	defer logFile.Close()

	logger.Println("yaRenamer started!")

	checkEt(logger)
	workDir := getConfig(logger)
	dirFiles, forExifTool := walkingOnFilesystem(workDir, logger)
	if len(dirFiles)+len(forExifTool) == 0 {
		logger.Println("Nothin to do!\nBye :)")
		os.Exit(0)
	}
	if len(dirFiles) > 0 {
		mustCompile1 := regexp.MustCompile(`^[A-Z]{3}_(\d{8})_(\d{6})`)
		mustCompile2 := regexp.MustCompile(`^.*(\d{4})[_:-]?(\d{2})[_:-]?(\d{2})[_:-](\d{6})`)
		mustCompile3 := regexp.MustCompile(`^.*(\d{4})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})[_:-](\d{2})`)
		for key, val := range dirFiles {
			logger.SetPrefix(filepath.Base(key) + " ")
			switch {
			case val.doByName:
				nameSlice := mustCompile1.FindStringSubmatch(filepath.Base(key))
				newName := nameSlice[1] + "_" + nameSlice[2]
				renamer(key, newName, logger)
				logger.Println("New name is a: " + newName + "of file:" + key)
			case val.doByName2:
				nameSlice := mustCompile2.FindStringSubmatch(filepath.Base(key))
				newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4]
				renamer(key, newName, logger)
				logger.Println("New name is a: " + newName + "of file:" + key)
			case val.doByName3:
				nameSlice := mustCompile3.FindStringSubmatch(filepath.Base(key))
				newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4] + nameSlice[5] + nameSlice[6]
				renamer(key, newName, logger)
				logger.Println("New name is a: " + newName + "of file:" + key)
			default:
				logger.Println("Look like something wrong in main::for::switch block ", key)
			}
		}
	}
	if len(forExifTool) > 0 && exiftoolExist {
		et, err := exiftool.NewExiftool()
		if err != nil {
			panic(fmt.Errorf("Error when intializing: %v", err))
		}
		defer et.Close()
		for _, item := range forExifTool {
			logger.SetPrefix(filepath.Base(item) + " ")
			exifData, err := getExif(et, item, logger)
			if err != nil { //если не получилось вынуть exif
				logger.Println("func:fsTimeStamp; when exif data corrupted")
				fInfo, err := os.Stat(item)
				check(err)
				fTimestamp := fInfo.ModTime()
				newName := fTimestamp.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
				renamer(item, newName, logger)
				logger.Println("main:fsTimeStamp:rename; newName: " + newName)
			} else {
				newName := exifData.Format(stdLongYear + stdZeroMonth + stdZeroDay + "_" + stdHour + stdZeroMinute + stdZeroSecond)
				renamer(item, newName, logger)
				logger.Println("main:exifToolRename; newName: " + newName)
			}
		}
	}
}
