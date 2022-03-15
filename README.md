# immudb_connect_benchmark

This go binary generates multiple connection towards immudb and then, for each of them, a sequence of simple gRPC commands.

It can be used to check immudb performance in a scenario where multiple connection are expected, or to measure proxy performance in such situation.

Included are traefik yaml configuration files.


