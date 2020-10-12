package pkg

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/barasher/go-exiftool"
)

//LetsGo basis action
func LetsGo() {
	logFile, err := os.OpenFile("yarenamer-"+timeNow.Format("20060102150405")+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(mw, "INFO: ", log.Flags()&^(log.Ldate|log.Ltime))
	defer logFile.Close()

	logger.Println("yaRenamer started!")

	checkEt(logger)
	workDir = checkWorkDir(logger)
	dirFiles, forExifTool := walkingOnFilesystem(workDir, logger)
	totalFiles := len(dirFiles) + len(forExifTool)
	if totalFiles == 0 {
		logger.Println("Nothin to do!\nBye :)")
		os.Exit(0)
	}

	if len(dirFiles) > 0 {
		mustCompile := regexp.MustCompile(`.*(\d{4})[\._:-]?(\d{2})[\._:-]?(\d{2})[\._:-]?\s?(\d{2})[\._:-]?(\d{2})[\._:-]?(\d{2}).*`)
		for _, item := range dirFiles {
			if !fileExists(item) {
				continue
			}
			logger.SetPrefix(filepath.Base(item) + " ")
			nameSlice := mustCompile.FindStringSubmatch(filepath.Base(item))
			if areYearActual(nameSlice[1], logger); err != nil {
				logger.Println("Moved to exifRenamer func; when DateInName data corrupted")
				forExifTool = append(forExifTool, item)
				fmt.Println(err)
				continue
			}
			newName := nameSlice[1] + nameSlice[2] + nameSlice[3] + "_" + nameSlice[4] + nameSlice[5] + nameSlice[6]
			renamer(item, newName, logger)
			logger.Println("New name is a: " + newName + "of file: " + item)
		}
	}
	if len(forExifTool) > 0 && exiftoolExist {
		et, err := exiftool.NewExiftool()
		if err != nil {
			panic(fmt.Errorf("Error when intializing: %v", err))
		}
		defer et.Close()
		for _, item := range forExifTool {
			if !fileExists(item) {
				continue
			}
			logger.SetPrefix(filepath.Base(item) + " ")
			exifDateParsed, err := getExif(et, item, logger)
			if err != nil { //if getExif data is failed
				logger.Println(err)
				logger.Println("func:fsTimeStamp; when exif data corrupted.")
				useFSTimeStamp(item, logger)
			} else {
				renamer(item, exifDateParsed, logger)
				logger.Println("exifToolRename; newName: " + exifDateParsed)
			}
		}
	} else {
		if !exiftoolExist {
			logger.Println("SKIPPED: ", len(forExifTool), " files in ExifTool processing. Because exiftool is not installed.")
			totalFiles = totalFiles - len(forExifTool)
		}
	}
	logger.SetPrefix("INFO: ")
	logger.Println("Total files processed: ", totalFiles)
	if checkDublesFlag {
		logger.Println("Total removed dublicates: ", removedCount)
	}
}
