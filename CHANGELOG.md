0.14.0
------
- Update to Alpine 3.4
- Cap request size in Datadog backend (#27)
- Set timeouts on Dials and tcp sends (#23)
- Reuse HTTP connections in Datadog backend

0.13.4
------
- Async rate limited cloud provider lookups (#22, #3)
- Internals refactoring

0.13.3
------
- Add configurable CPU profiler endpoint

0.13.2
------
- Increase default Datadog timeouts to reduce number of errors in the logs

0.13.1
------
- Log intermediate errors in Datadog backend
- Consistently set timeouts for AWS SDK service clients
- Update all dependencies
- Fix Datadog backend retry error #18
- Various internal improvements

0.13.0
------
- Fix goroutine start bug in dispatcher - versions 0.12.6, 0.12.7 do not work properly
- [Datadog events](http://docs.datadoghq.com/guides/overview/#events) support

0.12.7
------
- Remove deprecated -f flag passed to Docker tag command
- Rename num_stats back to numStats to be compatible with original statsd

0.12.6
------
- null backend to do benchmarking
- Internals refactoring

0.12.5
------
- Implement negative lookup cache (#8)
- Read configuration from environment, flags and config

0.12.4
------
- Do not multiply number of metric workers

0.12.3
------
- Do not replace dash and underscore in metric names and tags

0.12.2
------
- Fix handling of NaN specific case
- Minor refactorings for linter

0.12.1
------
- Use pointer to metric instead of passing by value
- Optimise (5-10x) performance of line parser

0.12.0
------
- Revert dropping message and block instead
- Use different values for number of worker to read from socket, process messages and metrics
- Minor fixes and performance improvements
- Calculate per second counters since last flush time instead of interval

0.11.2
------
- Process internal stats as standard metrics
- Add benchmarks
- Improve performance for buffering data out of the socket
- Drop messages and metrics instead of blocking when overloaded

0.11.1
------
- Normalize tags consistently but don't force lower case for metric names

0.11.0
------
- Add support for cloud plugins to retrieve host information e.g. instance id, aws tags, etc.

0.10.4
------
- Add load testing tool

0.10.3
------
- Performance improvements for receiver and aggregator

0.10.2
------
- Use goroutines to read net.PacketConn instead of buffered channel

0.10.1
------
- Graphite: replace dots by underscores in metric name (tags)
- Limit concurrency by using buffered channels for incoming messages
- Discard empty tags
- Datadog: add retries on post errors

0.10.0
------
- Add support for default tags

0.9.1
-----
- statsd backend: ensure not going over the udp datagram max size on metrics send
- Datadog: normalise tags to always be of form "key:value"

0.9.0
-----
- Reset counters, gauges, timers, etc. after an expiration delay
- Datadog: add interval to metric payload
- Datadog: use source ip address as hostname

0.8.5
-----
- Remove extraneous dot in metric names for stdout and graphite backends when no tag
- Add more and improve internal statsd stats

0.8.4
-----
- Datadog: set dogstatsd version and user-agent headers
- Datadog: use `rate` type for per second metric

0.8.3
-----
- Fix issue with tags overwriting metrics without tags

0.8.2
-----
- Fix undesired overwriting of metrics

0.8.1
-----
- Datadog: rollback change to metric names for preserving backward compatibility

0.8.0
-----
- Add more timers aggregations e.g. mean, standard deviation, etc.

0.7.1
-----
- Improve `graphite` backend reliability by re-opening connection on each metrics send
- Fix value of metrics displayed in web UI
- Improve web console UI look'n'feel

0.7.0
-----
- Add statsd backend

0.6.0
-----
- Add support for global metrics namespace
- Datadog: remove api url from config

0.5.0
-----
- Add support for "Set" metric type

0.4.2
-----
- Datadog: send number of metrics received
- Use appropriate log level for errors

0.4.1
-----
- Remove logging the url on each flush in datadog backend
- Use alpine base image for docker instead of scratch to avoid ca certs root error

0.4.0
-----
- Add datadog backend

0.3.1
-----
- Fix reset of metrics

0.3.0
-----
- Implement tags handling: use tags in metric names

0.2.0
-----
- Implement support for pluggable backends
- Add basic stdout backend
- Configure backends via toml, yaml or json configuration files
- Add support for tags and sample rate

0.1.0
-----
- Initial release
