package storage

var (
	jsonRepoInstance  = NewJSONRepo(defaultJSONHDataPath, defaultJSONDataPath)
	redisRepoInstance = NewRedisRepository(defaultRedisHDataPath, defaultRedisDataPath)
)

func DefaultJSONRepo() Repository {
	return jsonRepoInstance
}
func DefaultRedisRepo() RedisRepository {
	return redisRepoInstance
}
