package storage

var (
	jsonRepoInstance  = NewJSONRepo(defaultJSONHDataPath, defaultJSONDataPath)
	redisRepoInstance = NewRedisRepository(defaultRedisHDataPath)
)

func DefaultJSONRepo() Repository {
	return jsonRepoInstance
}
func DefaultRedisRepo() RedisRepository {
	return redisRepoInstance
}
