package main

import (
	"errors"
	"fmt"
	"gitlab.com/codmill/customer-projects/guardian/lobby-attendant/helpers"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

type ListingHandler struct {
	RootPath   string
	LevelLimit int
}

func getListingArgs(requestUrl *url.URL) (string, int, error) {
	parentPath := requestUrl.Query().Get("parent")
	depthString := requestUrl.Query().Get("depth")

	var depthNum int64
	if depthString != "" {
		var convErr error
		depthNum, convErr = strconv.ParseInt(depthString, 10, 16)
		if convErr != nil {
			return "", 0, errors.New(fmt.Sprintf("ERROR: could not convert depth argument '%s' to an integer: %s", depthString, convErr))
		}
	}
	return parentPath, int(depthNum), nil
}

func recurseDirectories(fromDir string, depth int, level int) ([]FileEntry, error) {
	log.Printf("debug: at directory %s level %d depth %d", fromDir, level, depth)

	files, readErr := ioutil.ReadDir(fromDir)
	if readErr != nil {
		log.Printf("Could not read '%s': %s", fromDir, readErr)
		return nil, readErr
	}

	log.Printf("level %d got %d more entries", level, len(files))

	ongoingResults := make([]FileEntry, len(files))
	for i, file := range files {
		ongoingResults[i] = NewFileEntry(fromDir, level, file)
	}

	for _, f := range files {
		if f.IsDir() && !strings.HasPrefix(f.Name(), ".") && level < depth {
			newFiles, nextLevelErr := recurseDirectories(path.Join(fromDir, f.Name()), depth, level+1)
			if nextLevelErr != nil {
				if os.IsPermission(nextLevelErr) {
					log.Printf("WARNING ListingHandler permission denied for %s", path.Join(fromDir, f.Name()))
				} else if os.IsNotExist(nextLevelErr) {
					log.Printf("WARNING ListingHandler path %s removed while iterating", path.Join(fromDir, f.Name()))
				} else {
					return nil, nextLevelErr
				}
			} else {
				ongoingResults = append(ongoingResults, newFiles...)
			}
		}
	}
	log.Printf("level %d returning %d entries", level, len(files))
	return ongoingResults, nil
}

func (h ListingHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if !helpers.AssertHttpMethod(request, w, "GET") {
		return //http error is already sent
	}

	requestUrl, urlErr := url.ParseRequestURI(request.RequestURI)
	if urlErr != nil {
		helpers.WriteJsonContent(helpers.GenericErrorResponse{
			Status: "internal_error",
			Detail: "could not parse request uri, this should not happen",
		}, w, 500)
		return
	}

	parentPath, depth, argErr := getListingArgs(requestUrl)
	if argErr != nil {
		helpers.WriteJsonContent(helpers.GenericErrorResponse{
			Status: "invalid_arguments",
			Detail: "could not parse arguments, see log for details",
		}, w, 400)
		return
	}

	if depth > h.LevelLimit {
		helpers.WriteJsonContent(helpers.GenericErrorResponse{
			Status: "invalid_arguments",
			Detail: "requested recursion is too deep",
		}, w, 400)
		return
	}
	targetDir := path.Join(h.RootPath, parentPath)
	if _, statErr := os.Stat(targetDir); os.IsNotExist(statErr) {
		helpers.WriteJsonContent(helpers.GenericErrorResponse{
			Status: "not_found",
			Detail: "not found",
		}, w, 404)
		return
	}

	fullResults, resultErr := recurseDirectories(targetDir, depth, 0)
	if resultErr != nil {
		log.Printf("ERROR ListingHandler could not list '%s' for '%s': %s", targetDir, parentPath, resultErr)
		helpers.WriteJsonContent(helpers.GenericErrorResponse{
			Status: "internal_error",
			Detail: "could not list",
		}, w, 500)
		return
	}

	helpers.WriteJsonContent(fullResults, w, 200)
}
