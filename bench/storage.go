package bench

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ReadRuns(path string) ([]Run, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()

	return DecodeRuns(f)
}

func DecodeRuns(r io.Reader) ([]Run, error) {
	var runs []Run
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&runs); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}
	return runs, nil
}

func WriteRuns(path string, runs []Run) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer f.Close()

	return EncodeRuns(f, runs)
}

func EncodeRuns(w io.Writer, runs []Run) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(runs); err != nil {
		return fmt.Errorf("error encoding JSON: %w", err)
	}
	return nil
}

func WriteRunToFile(path string, run *Run, appendMode bool) error {
	var runs []Run

	if appendMode {

		existingRuns, err := ReadRuns(path)
		if err == nil {
			runs = existingRuns
		}
	}

	runs = append(runs, *run)

	return WriteRuns(path, runs)
}

func CreateRun(suites []Suite, version string, date int64, tags []string) Run {
	return Run{
		Version: version,
		Date:    date,
		Tags:    tags,
		Suites:  suites,
	}
}
