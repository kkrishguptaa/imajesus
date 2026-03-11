package main

import (
	"os"
	"path/filepath"
	"strings"
)

type ImageJob struct {
	Input  string
	Output string
}

var supportedExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true,
}

func webpName(path string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ".webp"
}

func LoopDirectory(dir string) ([]ImageJob, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var jobs []ImageJob
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !supportedExts[ext] {
			continue
		}
		jobs = append(jobs, ImageJob{
			Input:  filepath.Join(dir, e.Name()),
			Output: filepath.Join(dir, webpName(e.Name())),
		})
	}
	return jobs, nil
}
