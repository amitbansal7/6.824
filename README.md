# MIT 6.824 Labs
## Test output

### Lab 1:
```
*** Starting wc test.
2020/05/12 17:06:03 rpc.Register: reply type of method "AddTasks" is not a pointer: "[]string"
2020/05/12 17:06:03 rpc.Register: method "Checker" has 1 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: method "CleanWorkerFiles" has 2 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: method "CreateNewWork" has 1 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: method "CreateTasks" has 2 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: method "Done" has 1 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: reply type of method "Init" is not a pointer: "int"
2020/05/12 17:06:03 rpc.Register: method "RemoveActiveWork" has 2 input parameters; needs exactly three
2020/05/12 17:06:03 rpc.Register: method "UpdateFiles" has 2 input parameters; needs exactly three
--- wc test: PASS
*** Starting indexer test.
2020/05/12 17:06:12 rpc.Register: reply type of method "AddTasks" is not a pointer: "[]string"
2020/05/12 17:06:12 rpc.Register: method "Checker" has 1 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: method "CleanWorkerFiles" has 2 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: method "CreateNewWork" has 1 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: method "CreateTasks" has 2 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: method "Done" has 1 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: reply type of method "Init" is not a pointer: "int"
2020/05/12 17:06:12 rpc.Register: method "RemoveActiveWork" has 2 input parameters; needs exactly three
2020/05/12 17:06:12 rpc.Register: method "UpdateFiles" has 2 input parameters; needs exactly three
--- indexer test: PASS
*** Starting map parallelism test.
2020/05/12 17:06:15 rpc.Register: reply type of method "AddTasks" is not a pointer: "[]string"
2020/05/12 17:06:15 rpc.Register: method "Checker" has 1 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: method "CleanWorkerFiles" has 2 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: method "CreateNewWork" has 1 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: method "CreateTasks" has 2 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: method "Done" has 1 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: reply type of method "Init" is not a pointer: "int"
2020/05/12 17:06:15 rpc.Register: method "RemoveActiveWork" has 2 input parameters; needs exactly three
2020/05/12 17:06:15 rpc.Register: method "UpdateFiles" has 2 input parameters; needs exactly three
--- map parallelism test: PASS
*** Starting reduce parallelism test.
2020/05/12 17:06:22 rpc.Register: reply type of method "AddTasks" is not a pointer: "[]string"
2020/05/12 17:06:22 rpc.Register: method "Checker" has 1 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: method "CleanWorkerFiles" has 2 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: method "CreateNewWork" has 1 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: method "CreateTasks" has 2 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: method "Done" has 1 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: reply type of method "Init" is not a pointer: "int"
2020/05/12 17:06:22 rpc.Register: method "RemoveActiveWork" has 2 input parameters; needs exactly three
2020/05/12 17:06:22 rpc.Register: method "UpdateFiles" has 2 input parameters; needs exactly three
--- reduce parallelism test: PASS
*** Starting crash test.
2020/05/12 17:06:30 rpc.Register: reply type of method "AddTasks" is not a pointer: "[]string"
2020/05/12 17:06:30 rpc.Register: method "Checker" has 1 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: method "CleanWorkerFiles" has 2 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: method "CreateNewWork" has 1 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: method "CreateTasks" has 2 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: method "Done" has 1 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: reply type of method "Init" is not a pointer: "int"
2020/05/12 17:06:30 rpc.Register: method "RemoveActiveWork" has 2 input parameters; needs exactly three
2020/05/12 17:06:30 rpc.Register: method "UpdateFiles" has 2 input parameters; needs exactly three
2020/05/12 17:07:10 dialing:dial unix /var/tmp/824-mr-0: connect: connection refused
2020/05/12 17:07:10 dialing:dial unix /var/tmp/824-mr-0: connect: connection refused
--- crash test: PASS
*** PASSED ALL TESTS
```


### Lab 2:

```
λ go test -race -run 2
Test (2A): initial election ...
  ... Passed --   3.0  3   52   14112    0
Test (2A): election after network failure ...
  ... Passed --   5.7  3  150   30200    0
Test (2B): basic agreement ...
  ... Passed --   0.7  3   16    4310    3
Test (2B): RPC byte count ...
  ... Passed --   1.6  3   50  114104   11
Test (2B): agreement despite follower disconnection ...
  ... Passed --   4.1  3   90   23506    7
Test (2B): no agreement if too many followers disconnect ...
  ... Passed --   3.5  5  248   47552    3
Test (2B): concurrent Start()s ...
  ... Passed --   0.5  3   16    4751    6
Test (2B): rejoin of partitioned leader ...
  ... Passed --   6.9  3  186   43905    4
Test (2B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  20.0  5 2672 1968137  102
Test (2B): RPC counts aren't too high ...
  ... Passed --   2.1  3   62   19097   12
Test (2C): basic persistence ...
  ... Passed --   4.9  3   92   22570    6
Test (2C): more persistence ...
  ... Passed --  20.4  5 1280  246186   16
Test (2C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   2.2  3   36    8789    4
Test (2C): Figure 8 ...
  ... Passed --  31.0  5  560   98310   14
Test (2C): unreliable agreement ...
  ... Passed --   2.1  5 1056  373678  246
Test (2C): Figure 8 (unreliable) ...
  ... Passed --  41.4  5 9742 22892640   37
Test (2C): churn ...
  ... Passed --  16.5  5 4056 1933022  761
Test (2C): unreliable churn ...
  ... Passed --  16.5  5 2776 4165010  518
PASS
ok    6.824/src/raft  183.924s
```