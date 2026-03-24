package storage

var (
	jsonRepoInstance = NewJSONRepo(defaultJSONHDataPath, defaultJSONDataPath)
)

func DefaultJSONRepo() Repository {
	return jsonRepoInstance
}
