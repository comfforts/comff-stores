package loader

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/comfforts/comff-stores/pkg/errors"
)

// ReadFileArray reads an array of json data from existing file, one by one,
// and returns individual result at defined rate through returned channel
func ReadFileArray(ctx context.Context, cancel func(), fileName string) (<-chan map[string]interface{}, error) {
	filePath := filepath.Join("data", fileName)

	// check if file exists
	err := ifFileExists(filePath)
	if err != nil {
		fmt.Printf("error checking file: %s existence, %v\n", filePath, err)
		cancel()
		return nil, errors.WrapError(err, "error checking %s existence", filePath)
	}

	// Open file and deferred close it
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("error opening file %s, %v\n", filePath, err)
		cancel()
		return nil, errors.WrapError(err, "error reading file: %s", filePath)
	}

	resultStream := make(chan map[string]interface{}, 2)
	go func(ct context.Context, can func(), fp string, fi *os.File, rs chan map[string]interface{}) {
		defer func(rst chan map[string]interface{}) {
			fmt.Println("Closing result stream")
			close(rst)
		}(rs)

		defer func(fpa string, fil *os.File, canc func()) {
			fmt.Println("Closing file")
			if err = fil.Close(); err != nil {
				fmt.Printf("error closing file %s, %v\n", filePath, err)
				canc()
			}
		}(fp, fi, can)

		r := bufio.NewReader(fi)
		dec := json.NewDecoder(r)

		// read open bracket
		t, err := dec.Token()
		if err != nil {
			fmt.Printf("error reading starting token %v from file %s, %v\n", t, filePath, err)
			cancel()
		}

		// while the array contains values
		for dec.More() {
			var result map[string]interface{}
			err := dec.Decode(&result)
			if err != nil {
				fmt.Printf("error decoding result json, %v\n", err)
			}
			// log.Printf("Retrieved %#v\n", result)
			select {
			case <-ct.Done():
				return
			case rs <- result:
			}
		}

		// read closing bracket
		t, err = dec.Token()
		if err != nil {
			fmt.Printf("error reading closing token %v from file %s, %v\n", t, filePath, err)
			can()
		}
	}(ctx, cancel, fileName, f, resultStream)

	return resultStream, nil
}

// checks if file exists
func ifFileExists(filePath string) error {
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Println(err)
	// }
	// log.Println("ifFileExists() - current path: ", path)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("file doesn't exist %s, %v\n", filePath, err)
			return errors.WrapError(err, "File: %s doesn't exist", filePath)
		} else {
			fmt.Printf("unable to access file %s, %v\n", filePath, err)
			return errors.WrapError(err, "Error accessing file: %s", filePath)
		}
	}
	return nil
}
