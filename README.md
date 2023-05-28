# QOTD-service
Qoute of the day service built with Go


#### We want to build a quote of the day service
- - -
### Dependencies
1. Go
2. gRPC
3. protoc
4. buf


### Project Diretory
```
$ tree
.
|-- LICENSE
|-- client   
|   |-- bin  
|   |   `-- qotd.go 
|   `-- client.go   
|-- go.mod
|-- go.sum
|-- internal        
|   `-- server      
|       `-- server.go   
|-- makefile
|-- proto
|   `-- qotd
|       |-- buf.gen.yaml
|       |-- buf.yaml       
|       |-- qotd.pb.go     
|       |-- qotd.proto     
|       `-- qotd_grpc.pb.go
|-- qotd.go  
|-- qotd.png 
`-- qotd_rest
    `-- main.go

7 directories, 16 files
```

### Build & Run

```
# start server
go run qotd.go --addr="127.0.0.1:3000
# if `addr` is left blank, it will default to `127.0.0.1:80`
```

```
# start client
go run client/bin/qotd.go --addr="127.0.0.1:3000 --author="Dave Chapelle"
# if author is left blank,it will choose a random author
# if `addr` is left blank, it will default to `127.0.0.1:80`
```
![QOTD](https://github.com/marcusbello/QOTD-service/raw/main/qotd.png)
