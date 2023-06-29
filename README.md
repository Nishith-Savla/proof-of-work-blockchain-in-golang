A demo Proof of Work blockchain prototype built in Golang

To run -
```bash
go run main.go
```

Open `http://localhost:8080` in a browser to view the blocks created.

To add blocks, you send a POST request to `localhost:8080` using CURL.

```bash
curl -X POST -H "Content-Type: application/json" -d '{"data": 75}' http://localhost:8080
```

The blockchain will start calculating the hash and validating each hash generated.
Once it will find a hash that matches the given criteria, the block would be generated and added to the blockchain.
