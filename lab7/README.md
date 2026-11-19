# Lab 7 - DynamoDB

## Usage with RPC Setup

cd into the lab7 directory

```sh
cd lab7
```

To start all the server nodes simulataneously as individual processes, use the following, with <N> being the total number of server nodes.

```sh
python start_election.py <N>
```

Then, start the client process in a separate terminal by running

```sh
go run client/client.go
```

Once connected, you can interface with the distributed database using the following commands:

```sh
put <key> <value>
get <key>
quit
```

### Troubleshooting

Note that if the nodes do not seem to connect to one another, try adjusting the BASE_PORT global const in BOTH `start_election.py` and `client.go` to a different allowable port. This was found to be the cause of some issues on one our of local machines.

### References

(dynamo/dynamoDB Distributed Key-Value store in Go)[https://deepengineering.net/p/building-a-distributed-key-value]
