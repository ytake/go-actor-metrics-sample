# Actor + OpenTelemetry + NewRelic

Proto Actorを使ったアクターシステムを監視するためのサンプルです。

```mermaid
graph TB
    subgraph "Consul Cluster"
        CA1[Consul Agent 1]
        CA2[Consul Agent 2]
        CA3[Consul Agent 3]
        CS1[Consul Server 1]
        CS2[Consul Server 2]
        CSB[Consul Server Bootstrap]
        CA1 --- CS1
        CA2 --- CS2
        CA3 --- CSB
        CS1 --- CS2
        CS2 --- CSB
        CSB --- CS1
    end

    subgraph "Proto Actor Cluster"
        Client[Client Server]
        FizzGrain[Fizz Grain]
        BuzzGrain[Buzz Grain]
        Client <-.->|gRPC| FizzGrain
        Client <-.->|gRPC| BuzzGrain
    end

    CS1 --- Client
    CS2 --- Client
    CSB --- Client
```

```bash
$ export OTEL_EXPORTER_OTLP_ENDPOINT=your_endpoint
$ export NR_API_KEY=your_api_key
```

## run

下記のコマンドで起動します  
fizzアクター/buzzアクターはそれぞれ別のターミナルで起動してください  
クラスタになっており virtual actor / grainとして起動します

```bash
# run client
$ go run main.go
# run fizz buzz actor
$ go run fizz/main.go
$ go run buzz/main.go
```

クラスタの管理にはzookeeperを利用しています。  

## proto file generate

[proto](./shared) ディレクトリにあるprotoファイルをコンパイルする場合に利用します.  
サンプルには含まれているため不要ですが、protoファイルを変更した場合には以下のコマンドでコンパイルしてください。  

```bash
$ go install github.com/asynkron/protoactor-go/protobuf/protoc-gen-gograinv2@dev 
```

