# Redis Cluster Failover

This is a test to reproduce an issue with Redis Cluster during a failover, with and without authentication required for the nodes.

## No auth required

First, when no auth is required, it works as expected. Running against our docker-compose Redis cluster:

```
# fleet instance, async processing enabled to add activity on all redis nodes
$ ./build/fleet serve --dev --osquery_detail_update_interval 1m --redis_address 127.0.0.1:7001 --redis_cluster_follow_redirections --osquery_enable_async_host_processing true

# osquery-perf running to simulate 100 hosts
$ go run ./cmd/osquery-perf/agent.go --host_count 100 --enroll_secret ...

# trigger a failover by killing a primary node
$ redis-cli -p 7004 DEBUG SEGFAULT

# fleet logs show CLUSTERDOWN error for a few seconds, then settles and resumes without further error
level=info ts=2022-09-26T18:02:54.460325422Z hostID=400 <-- before failover

-- failover period
level=error ts=2022-09-26T18:03:35.451097803Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.452832339Z component=http method=POST uri=/api/osquery/distributed/read took=4.375326ms ip_addr=127.0.0.1:46104 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.454514398Z component=http method=POST uri=/api/osquery/distributed/write took=713.81µs ip_addr=127.0.0.1:46104 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.549627067Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.550354931Z component=http method=POST uri=/api/osquery/distributed/read took=2.205733ms ip_addr=127.0.0.1:39556 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.552235195Z component=http method=POST uri=/api/osquery/distributed/write took=1.410764ms ip_addr=127.0.0.1:39556 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.83446799Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.836012466Z component=http method=POST uri=/api/osquery/distributed/read took=4.566934ms ip_addr=127.0.0.1:46152 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.838115329Z component=http method=POST uri=/api/osquery/distributed/write took=1.158836ms ip_addr=127.0.0.1:46152 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.933404698Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.93493938Z component=http method=POST uri=/api/osquery/distributed/read took=4.271876ms ip_addr=127.0.0.1:39588 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:35.937775879Z component=http method=POST uri=/api/osquery/distributed/write took=1.830599ms ip_addr=127.0.0.1:39588 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:40.75624143Z component=http method=POST uri=/api/osquery/distributed/read took=5.005747596s ip_addr=127.0.0.1:39574 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:45.348621673Z component=http method=POST uri=/api/osquery/distributed/read took=10.006297236s ip_addr=127.0.0.1:39548 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:45.655529199Z component=http method=POST uri=/api/osquery/distributed/read took=10.006409945s ip_addr=127.0.0.1:39562 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:03:46.04921201Z component=http method=POST uri=/api/osquery/distributed/read took=10.006717323s ip_addr=127.0.0.1:39594 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
-- end of failover

-- I ran a live query with fleetctl query to generate some more logs
level=error ts=2022-09-26T18:05:15.017211059Z component=http user=martin.n.angers+test@gmail.com method=POST uri=/api/latest/fleet/queries/run_by_names took=8.435452ms sql="select * from osquery_info;" query_id=null numHosts=0 err="no hosts targeted"
2022/09/26 14:05:53 http: response.WriteHeader on hijacked connection from github.com/prometheus/client_golang/prometheus/promhttp.(*responseWriterDelegator).WriteHeader (delegator.go:65)
2022/09/26 14:05:53 http: response.Write on hijacked connection from github.com/prometheus/client_golang/prometheus/promhttp.(*responseWriterDelegator).Write (delegator.go:74)
```

Restarting the failed Redis node (with `docker-compose start redis-cluster-4` in this case) properly brings back the node as a replica in the cluster, no effect on Fleet (it keeps running properly).

## Auth required

Next step was to modify the redis configuration files in `tools/redis-tests/redis-clusert-{1-6}.conf` to add the authentication (`requirepass`) and the master authentication so that replicas can replicate from the primary (`masterauth`):

```
requirepass foobared
masterauth foobared
```

And then run Fleet with an additional `--redis_password foobared` flag. After that, trigger a failover in the same way, sending a `DEBUG SEGFAULT` to a primary node (though now we have to first run `AUTH foobared` to that node).

The result was pretty much the same as without auth, after a few seconds, the cluster fixed itself and the Fleet instance started working again:

```
level=info ts=2022-09-26T18:21:50.690658539Z hostID=500 <-- before failover, normal host enrolling

-- start of failover
level=error ts=2022-09-26T18:23:00.382513561Z component=http method=POST uri=/api/osquery/distributed/read took=33.799369ms ip_addr=127.0.0.1:55800 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:00.3825188Z component=http method=POST uri=/api/osquery/distributed/read took=259.175362ms ip_addr=127.0.0.1:55790 x_for_ip_addr= err="record host last seen: run redis script: read tcp 172.20.0.1:46170->172.20.0.36:7006: read: connection reset by peer"
level=error ts=2022-09-26T18:23:00.382619387Z component=http method=POST uri=/api/osquery/distributed/read took=173.655948ms ip_addr=127.0.0.1:55796 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:00.42233781Z component=http method=POST uri=/api/osquery/distributed/read took=2.865245ms ip_addr=127.0.0.1:60160 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:05.625107621Z component=http method=POST uri=/api/osquery/distributed/read took=5.003847448s ip_addr=127.0.0.1:60170 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:05.798377389Z component=http method=POST uri=/api/osquery/distributed/read took=5.004935878s ip_addr=127.0.0.1:55822 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:05.858632437Z component=http method=POST uri=/api/osquery/distributed/read took=5.005014436s ip_addr=127.0.0.1:55826 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:06.257066605Z component=http method=POST uri=/api/osquery/distributed/read took=5.005970831s ip_addr=127.0.0.1:55374 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:06.464851505Z component=http method=POST uri=/api/osquery/distributed/read took=5.005727129s ip_addr=127.0.0.1:55384 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:06.768684202Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:06.770128857Z component=http method=POST uri=/api/osquery/distributed/read took=5.006549261s ip_addr=127.0.0.1:55414 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:06.868358282Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:06.870042614Z component=http method=POST uri=/api/osquery/distributed/read took=5.005558895s ip_addr=127.0.0.1:55420 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:07.563636775Z op=QueriesForHost err="load active queries: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:07.565387662Z component=http method=POST uri=/api/osquery/distributed/read took=5.004787888s ip_addr=127.0.0.1:55862 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:07.896501703Z component=http method=POST uri=/api/osquery/distributed/read took=5.00398931s ip_addr=127.0.0.1:55528 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.101024814Z component=http method=POST uri=/api/osquery/distributed/read took=5.005116839s ip_addr=127.0.0.1:55884 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.182346904Z component=http method=POST uri=/api/osquery/distributed/read took=5.004373509s ip_addr=127.0.0.1:55560 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.277128757Z component=http method=POST uri=/api/osquery/distributed/read took=5.005209659s ip_addr=127.0.0.1:55898 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.373758337Z component=http method=POST uri=/api/osquery/distributed/read took=5.00620095s ip_addr=127.0.0.1:55910 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.573463255Z component=http method=POST uri=/api/osquery/distributed/read took=5.004005945s ip_addr=127.0.0.1:55922 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.723423738Z component=http method=POST uri=/api/osquery/distributed/read took=5.003958997s ip_addr=127.0.0.1:55610 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:08.995743727Z component=http method=POST uri=/api/osquery/distributed/read took=5.002679579s ip_addr=127.0.0.1:55946 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.095104336Z component=http method=POST uri=/api/osquery/distributed/read took=5.002678753s ip_addr=127.0.0.1:55952 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.175360889Z component=http method=POST uri=/api/osquery/distributed/read took=5.00331228s ip_addr=127.0.0.1:55966 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.276860742Z component=http method=POST uri=/api/osquery/distributed/read took=5.005991515s ip_addr=127.0.0.1:55668 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.426172493Z component=http method=POST uri=/api/osquery/distributed/read took=5.005848842s ip_addr=127.0.0.1:55982 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.593271809Z component=http method=POST uri=/api/osquery/distributed/read took=5.005009795s ip_addr=127.0.0.1:55996 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.690527992Z component=http method=POST uri=/api/osquery/distributed/read took=5.005697944s ip_addr=127.0.0.1:55998 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.776926602Z component=http method=POST uri=/api/osquery/distributed/read took=5.004110531s ip_addr=127.0.0.1:56008 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:09.992287061Z component=http method=POST uri=/api/osquery/distributed/read took=5.005668154s ip_addr=127.0.0.1:55724 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.189549968Z component=http method=POST uri=/api/osquery/distributed/read took=5.005808486s ip_addr=127.0.0.1:55756 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.308335733Z component=http method=POST uri=/api/osquery/distributed/read took=5.005285823s ip_addr=127.0.0.1:56020 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.403292792Z component=http method=POST uri=/api/osquery/distributed/read took=5.003257262s ip_addr=127.0.0.1:55776 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.499064545Z component=http method=POST uri=/api/osquery/distributed/read took=5.005523714s ip_addr=127.0.0.1:56036 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.537920105Z component=http method=POST uri=/api/osquery/distributed/read took=10.004954815s ip_addr=127.0.0.1:60164 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.615776105Z component=http method=POST uri=/api/osquery/distributed/read took=5.003160344s ip_addr=127.0.0.1:56052 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.721262376Z component=http method=POST uri=/api/osquery/distributed/read took=10.003908768s ip_addr=127.0.0.1:60180 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.800089854Z component=http method=POST uri=/api/osquery/distributed/read took=5.003233988s ip_addr=127.0.0.1:56056 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:10.95788744Z component=http method=POST uri=/api/osquery/distributed/read took=10.006325591s ip_addr=127.0.0.1:55834 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.36:7006"
level=error ts=2022-09-26T18:23:11.091070481Z component=http method=POST uri=/api/osquery/distributed/read took=10.003418088s ip_addr=127.0.0.1:55352 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.15716064Z component=http method=POST uri=/api/osquery/distributed/read took=10.003928977s ip_addr=127.0.0.1:55838 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.359097191Z component=http method=POST uri=/api/osquery/distributed/read took=10.00659888s ip_addr=127.0.0.1:55382 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.40650269Z component=http method=POST uri=/api/osquery/distributed/read took=5.004457043s ip_addr=127.0.0.1:55864 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.492870402Z component=http method=POST uri=/api/osquery/distributed/read took=5.004910556s ip_addr=127.0.0.1:56088 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.561155769Z component=http method=POST uri=/api/osquery/distributed/read took=10.006087538s ip_addr=127.0.0.1:55394 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.603545438Z component=http method=POST uri=/api/osquery/distributed/read took=5.006262498s ip_addr=127.0.0.1:56100 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.684743665Z component=http method=POST uri=/api/osquery/distributed/read took=10.006785595s ip_addr=127.0.0.1:55410 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.736198489Z component=http method=POST uri=/api/osquery/distributed/read took=5.002627003s ip_addr=127.0.0.1:55882 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.893244788Z component=http method=POST uri=/api/osquery/distributed/read took=5.003020209s ip_addr=127.0.0.1:56118 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:11.983199259Z component=http method=POST uri=/api/osquery/distributed/read took=10.007226432s ip_addr=127.0.0.1:55844 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:11.99138117Z component=http method=POST uri=/api/osquery/distributed/read took=5.002380208s ip_addr=127.0.0.1:56124 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.087719545Z component=http method=POST uri=/api/osquery/distributed/read took=10.006543044s ip_addr=127.0.0.1:55852 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.141428605Z component=http method=POST uri=/api/osquery/distributed/read took=5.00305894s ip_addr=127.0.0.1:55908 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.181990104Z component=http method=POST uri=/api/osquery/distributed/read took=10.005394446s ip_addr=127.0.0.1:55446 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.282419501Z component=http method=POST uri=/api/osquery/distributed/read took=10.006509984s ip_addr=127.0.0.1:55858 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.315530054Z component=http method=POST uri=/api/osquery/distributed/read took=5.005997368s ip_addr=127.0.0.1:60004 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.368931161Z component=http method=POST uri=/api/osquery/distributed/read took=10.007376911s ip_addr=127.0.0.1:55468 x_for_ip_addr= err="record host last seen: run redis script: CLUSTERDOWN The cluster is down"
level=error ts=2022-09-26T18:23:12.406031243Z component=http method=POST uri=/api/osquery/distributed/read took=5.00319869s ip_addr=127.0.0.1:60010 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.479359808Z component=http method=POST uri=/api/osquery/distributed/read took=10.004982235s ip_addr=127.0.0.1:55484 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.621559991Z component=http method=POST uri=/api/osquery/distributed/read took=5.004851159s ip_addr=127.0.0.1:55570 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.719997108Z component=http method=POST uri=/api/osquery/distributed/read took=10.005404836s ip_addr=127.0.0.1:55516 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
level=error ts=2022-09-26T18:23:12.765987815Z component=http method=POST uri=/api/osquery/distributed/read took=10.005468241s ip_addr=127.0.0.1:55872 x_for_ip_addr= err="record host last seen: run redis script: MOVED 7812 172.20.0.32:7002"
-- failover ended


-- started running a live query to force some Redis action
2022/09/26 14:23:50 http: response.WriteHeader on hijacked connection from github.com/prometheus/client_golang/prometheus/promhttp.(*responseWriterDelegator).WriteHeader (delegator.go:65)
2022/09/26 14:23:50 http: response.Write on hijacked connection from github.com/prometheus/client_golang/prometheus/promhttp.(*responseWriterDelegator).Write (delegator.go:74)
```
