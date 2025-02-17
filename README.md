# Server
Build docker image
```
docker build -f ./Dockerfile.server -t pow-server .
```
Run docker container
```
docker run --env-file ./server/.env --network host pow-server
```

# Client
Build docker image
```
docker build -f ./Dockerfile.client -t pow-client .
```
Run docker container (configure arguments over `./client/.env`)
```
docker run --network host pow-client
```
