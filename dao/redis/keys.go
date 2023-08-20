package redis

// 存放redis key

// redis key 尽量使用命名空间的方式(例如:分割)，方便查询和区分
const (
	KeyUserScoreZSet      = "scgpteval:user:score" // zset; 存储每个用户的刷题量
	KeyUserQuizSetPrefix  = "scgpteval:user:quiz:" // set; 存储某个用户刷过的题，后接用户id
	KeyUserInfoHashPrefix = "scgpteval:user:info:" // hash; 存储某个用户的信息，后接用户id
	KeyQuizBaseSet        = "scgpteval:quiz:base"  // set; 记录指定社区下的帖子，后面的参数为社区name
)
