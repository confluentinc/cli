Output flinkshell.txt
Output flinkshell.mp4
Set Height 5000

Type "./test/bin/confluent flink shell --compute-pool lfcp-123456"
Enter
Sleep 1s

Type "use catalog default;"
Enter
Sleep 1s

Type "use db1;"
Enter
Sleep 1s

Type "set;"
Enter
Sleep 1s

Type "set 'cli.test-key'='value';"
Enter
Sleep 1s

Type "set;"
Enter
Sleep 1s

Type "reset;"
Enter
Sleep 1s

Type "exit"
Enter
Sleep 1s
