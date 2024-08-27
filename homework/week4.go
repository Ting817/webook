package homework

// 作业：用本地缓存来替换 Redis
// 现在因为你不想连 Redis，所以你打算提供一个基于本地缓存实现的 cache.CodeCache。
//
// 你需要做几件事：
//
// 定义一个 CodeCache 接口，将现在的 CodeCache 改名为 CodeRedisCache。
// 提供一个基于本地缓存的 CodeCache 实现。你可以自主决定用什么本地缓存，在这个过程注意体会技术选型要考虑的点。
// 保证单机并发安全，也就是你可以假定这个实现只用在开发环境，或者单机环境下。
