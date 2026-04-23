package database

var (
	jsonRepoInstance  = NewJSONRepo(defaultJSONHDataPath)
	redisRepoInstance = NewRedisRepository(defaultRedisHDataPath)
)

func DefaultJSONRepo() Repository {
	return jsonRepoInstance
}
func DefaultRedisRepo() RedisRepository {
	return redisRepoInstance
}
