package dataloader

import (
	"goldenglow/m"
)

func OnLoadingError(file string) {
	logger.Error("loading json data err: ",
		"module", "data loader",
		"file", file,
	)
}
func OnStart(dir string, fileAmount int) {
	logger.Info("starting data loader", "dir", dir, "fileAmount", fileAmount)
}
func OnFinishedInfoSum(tagMap m.Map[int]) {
	logger.Info(" data loading is Done",
		"module", "data loader",
		"plugin amount", len(tagMap),
	)
	for tag, count := range tagMap {
		logger.Info("info:",
			"module", "data loader",
			"plugin tag", tag,
			"file count", count,
		)
	}
}
