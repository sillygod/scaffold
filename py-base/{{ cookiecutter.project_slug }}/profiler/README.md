# Profile

By default, pyroscope python client only supports cpu metrics.
This app intends to integrate the austin profiler into pyroscope which only accpets pprof file format.

## How it works

Compile the pprof's proto file to pb file which can be imported by python. Write the glue codes to transform the memory profile produced by the austin.

By design, we currently only implement a push mode client target to pyroscope server.


### spawn container with the following args

```sh
docker run xx --cap-add SYS_PTRACE xx
```
