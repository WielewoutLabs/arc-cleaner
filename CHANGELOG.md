# Changelog

## 0.1.0 (2024-09-14)


### Features

* add app entrypoint with --config option ([2243bfc](https://github.com/Wielewout/arc-cleaner/commit/2243bfc0bf9bafcccb237529856ebf6b90b2c4bc))
* add context with cancel on os signals ([882cbb3](https://github.com/Wielewout/arc-cleaner/commit/882cbb3d668f8f9606f1e80e2d80d9529d259c23))
* add dryrun to config ([c50cc83](https://github.com/Wielewout/arc-cleaner/commit/c50cc83c23a26fccf6ce9271eec0c6c0aacf2689))
* add helm chart ([f11f122](https://github.com/Wielewout/arc-cleaner/commit/f11f1229a2ab6d46165ddf12006e1dfc70c7caf0))
* add k8s client with github arc scheme ([e18dfc7](https://github.com/Wielewout/arc-cleaner/commit/e18dfc71eca562c05ed569f251bef8610e8f1c4d))
* add liveness and readiness http endpoints ([f11f122](https://github.com/Wielewout/arc-cleaner/commit/f11f1229a2ab6d46165ddf12006e1dfc70c7caf0))
* add log level to config ([ff49114](https://github.com/Wielewout/arc-cleaner/commit/ff49114b92d2f26be373d277768529085fb3fa49))
* add namespace to config ([8aefaf4](https://github.com/Wielewout/arc-cleaner/commit/8aefaf4e85fbdacfe9bcf1707c27c9419ce9a582))
* add version and commit in startup log ([a95bcd1](https://github.com/Wielewout/arc-cleaner/commit/a95bcd1861be781b42a8138bc6c4b269a5ab9126))
* compress binary ([33781ff](https://github.com/Wielewout/arc-cleaner/commit/33781ffa099e0062505890775b54917f2396bbf4))
* delete workflow pod if runner pod pending ([964efd0](https://github.com/Wielewout/arc-cleaner/commit/964efd031d9b9b98fd48fab5d224a4a477501b8f))
* delete workflow pod without ephemeral runner ([01eeefb](https://github.com/Wielewout/arc-cleaner/commit/01eeefb8a07e6392caeb565d1ecef356d0b4a46f))
* make clean up period configurable ([21f7944](https://github.com/Wielewout/arc-cleaner/commit/21f79440ed4f735b05d44e831ceb1338213a46ec))
* periodically clean up ephemeral runners ([583a469](https://github.com/Wielewout/arc-cleaner/commit/583a469bc647f186befa4eaa4d9af1fcca95f001))


### Bug Fixes

* fix logging with context ([61adaa4](https://github.com/Wielewout/arc-cleaner/commit/61adaa4b3cbb30fd29552cdc4df112a6831135d6))
* stop app when kubernetes client creation fails ([c0f2bb6](https://github.com/Wielewout/arc-cleaner/commit/c0f2bb63bef03116baaa07193a67cd77eb1e60e3))
* work on ephemeral runner iso set ([8b39e45](https://github.com/Wielewout/arc-cleaner/commit/8b39e458f7c3c1f7094e88e480ca36e9a9924bc0))


### Miscellaneous Chores

* release please as 0.1.0 ([81733ec](https://github.com/Wielewout/arc-cleaner/commit/81733ec367278bd971b1965f2352bdc41a268174))
