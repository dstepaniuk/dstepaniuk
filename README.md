# Go test project

Repository contains both server and client in respective folders.

## Server
* Navigate into server folder from the project root:
```shell script
cd ./server
```

* Build server:
```shell script
go build
```

* Run server tests:
```shell script
go test -v ./...
```

* Run server (replace placeholders with actual AWS credentials/configuration values): 
```shell script
AWS_ACCESS_KEY_ID=<...> AWS_SECRET_ACCESS_KEY=<...> AWS_REGION=<...> BUCKET_NAME=<...> ./server
```

* Performance

Memory usage: 180 MB

QPS: 12000

## Client
* Navigate into client folder from the project root:
```shell script
cd ./client
```

* Build client:
```shell script
go build
```

* Run client (runs 1 million iterations): 
```shell script
./client
```


## Time report

1. Basic server and client implementation 2h
2. Workers implementation 8h
3. Basic test and this report creation 1h

## Known issues:
1. missing gzip compression
2. improve solution because current solution blocks requests processing when object parts are delivered to s3. AWS Kinesis Firehose usage looks as more suitable solution to deliver data streams to S3 though it will not allow to meet all requirements regarding objects naming. 
3. refactor solution to interfaces to make the code more testable using mocks 
4. add more tests