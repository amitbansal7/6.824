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
  ... Passed --   3.1  3   58   15418    0
Test (2A): election after network failure ...
  ... Passed --   5.6  3  142   28172    0
Test (2B): basic agreement ...
  ... Passed --   0.8  3   16    4342    3
Test (2B): RPC byte count ...
  ... Passed --   1.6  3   48  113794   11
Test (2B): agreement despite follower disconnection ...
  ... Passed --   4.2  3   92   23658    7
Test (2B): no agreement if too many followers disconnect ...
  ... Passed --   3.7  5  232   44904    4
Test (2B): concurrent Start()s ...
  ... Passed --   0.6  3   20    6079    6
Test (2B): rejoin of partitioned leader ...
  ... Passed --   4.8  3  144   33355    4
Test (2B): leader backs up quickly over incorrect follower logs ...
  ... Passed --  20.3  5 2600 1969861  102
Test (2B): RPC counts aren't too high ...
  ... Passed --   2.0  3   58   20425   12
Test (2C): basic persistence ...
  ... Passed --   5.2  3  108   25670    6
Test (2C): more persistence ...
  ... Passed --  32.1  5 1996  374964   16
Test (2C): partitioned leader and one follower crash, leader restarts ...
  ... Passed --   4.2  3   46   10377    4
Test (2C): Figure 8 ...
  ... Passed --  34.6  5  648  114697   20
Test (2C): unreliable agreement ...
  ... Passed --   4.6  5 1132  394588  246
Test (2C): Figure 8 (unreliable) ...
  ... Passed --  45.1  5 9299 18245727  283
Test (2C): churn ...
  ... Passed --  16.3  5 3796 2512636  648
Test (2C): unreliable churn ...
  ... Passed --  16.6  5 3406 1693937  651
PASS
ok  	/6.824/src/raft	206.032s
```
