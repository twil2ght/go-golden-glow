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
