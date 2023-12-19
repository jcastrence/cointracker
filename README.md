# Cointracker
---
## Running Application
From root directory:
```go run main.go```
Application will run on localhost by default

---
## Endpoints

### /createaccount
```
curl --location 'http://127.0.0.1:8000/createaccount' \
--header 'Content-Type: application/json' \
--data '{
    "username": "hello",
    "password": "world"
}'
```

### /addaddresses
```
curl --location 'http://127.0.0.1:8000/addaddresses' \
--header 'Content-Type: application/json' \
--data '{
    "username": "hello",
    "password": "world",
    "addresses": [
        "3E8ociqZa9mZUSwGdSmAEMAoAxBK3FNDcd",
        "12xQ9k5ousS8MqNsMBqHKtjAtCuKezm2Ju",
        "bc1qm34lsc65zpw79lxes69zkqmk6ee3ewf0j77s3h"
    ]
}'
```

### /removeaddresses
```
curl --location --request DELETE 'http://127.0.0.1:8000/removeaddresses' \
--header 'Content-Type: application/json' \
--data '{
    "username": "hello",
    "password": "world",
    "addresses": [
        "bc1qm34lsc65zpw79lxes69zkqmk6ee3ewf0j77s3h"
    ]
}'
```

### /getaccountinfo
```
curl --location --request GET 'http://127.0.0.1:8000/getaccountinfo' \
--header 'Content-Type: application/json' \
--data '{
    "username": "hello",
    "password": "world"
}'
```

### /updateaccount
```
curl --location --request PUT 'http://127.0.0.1:8000/updateaccount' \
--header 'Content-Type: application/json' \
--data '{
    "username": "hello",
    "password": "world"
}'
```