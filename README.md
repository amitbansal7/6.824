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
λ go test -run 2
Test (2A): initial election ...
  ... Passed --   3.1  3   54   14664    0
Test (2A): election after network failure ...
  ... Passed --   6.1  3  158   31116    0
Test (2B): basic agreement ...
  ... Passed --   0.7  3   16    4342    3
Test (2B): RPC byte count ...
  ... Passed --   1.4  3   48  113794   11
Test (2B): agreement despite follower disconnection ...
  ... Passed --   4.3  3   92   23702    7
Test (2B): no agreement if too many followers disconnect ...
  ... Passed --   3.6  5  244   47054    3
Test (2B): concurrent Start()s ...
  ... Passed --   0.5  3   18    5397    6
Test (2B): rejoin of partitioned leader ...
  ... Passed --   7.6  3  202   47910    4
Test (2B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  17.2  5 2256 1808318  102
Test (2B): RPC counts aren't too high ...
  ... Passed --   2.0  3   58   18036   12
Test (2C): basic persistence ...
  ... Passed --   4.3  3   84   21466    6
Test (2C): more persistence ...
  ... Passed --  24.9  5 1572  298943   16
Test (2C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   2.6  3   40    9663    4
Test (2C): Figure 8 ...
  ... Passed --  24.9  5  476   90866   19
Test (2C): unreliable agreement ...
  ... Passed --   1.5  5 1036  364042  246
Test (2C): Figure 8 (unreliable) ...
  ... Passed --  31.6  5 10764 41995661  411
Test (2C): churn ...
  ... Passed --  16.1  5 14986 70869004 2684
Test (2C): unreliable churn ...
  ... Passed --  16.1  5 5112 5402667 1102
PASS
ok    6.824/src/raft  168.701s
```