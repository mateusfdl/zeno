package bench

import (
	"fmt"
	"slices"
	"sort"
)

func MergeRuns(runs ...[]Run) []Run {
	var result []Run
	for _, r := range runs {
		result = append(result, r...)
	}
	return result
}

func MergeRunsFromFiles(paths ...string) ([]Run, error) {
	var allRuns []Run

	for _, path := range paths {
		runs, err := ReadRuns(path)
		if err != nil {
			return nil, err
		}
		allRuns = append(allRuns, runs...)
	}

	return allRuns, nil
}

func SortByDate(runs []Run) {
	sort.Slice(runs, func(i, j int) bool { return runs[i].Date < runs[j].Date })
}

func SortByDateDescending(runs []Run) {
	sort.Slice(runs, func(i, j int) bool { return runs[i].Date > runs[j].Date })
}

func DeduplicateRuns(runs []Run) []Run {
	seen := make(map[string]bool)
	var result []Run

	for _, run := range runs {
		key := fmt.Sprintf("%s:%d", run.Version, run.Date)
		if !seen[key] {
			seen[key] = true
			result = append(result, run)
		}
	}

	return result
}

func FilterByTag(runs []Run, tag string) []Run {
	var result []Run

	for _, run := range runs {
		if slices.Contains(run.Tags, tag) {
			result = append(result, run)
			break
		}
	}

	return result
}

func FilterByTags(runs []Run, tags []string) []Run {
	var result []Run

	for _, run := range runs {
		if hasAllTags(run, tags) {
			result = append(result, run)
		}
	}

	return result
}

func hasAllTags(run Run, tags []string) bool {
	tagMap := make(map[string]bool)
	for _, t := range run.Tags {
		tagMap[t] = true
	}

	for _, tag := range tags {
		if !tagMap[tag] {
			return false
		}
	}

	return true
}
