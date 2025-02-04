# Sample Requests

## Indices
* Notice - indices here are more like aliases than indices in elasticsearch

### Create Indices 
`index` - is the alias of the index to be created, index will be created with datetime in the elastic.
```shell

curl --location 'http://localhost/indices/' \
--data '{
    "index": "sample-index",
    "fields": {
        "id": {
            "type": "integer"
        },
        "name": {
            "type": "text",
            "autocomplete": true,
            "search": true
        },
        "price": {
            "type": "float"
        },
        "is_available": {
            "type": "boolean"
        },
        "description": {
            "type": "text",
            "analyzer": "english"
        },
        "metadata": {
            "type": "nested",
            "properties": {
                "tags": {
                    "type": "keyword"
                },
                "author": {
                    "type": "text"
                }
            }
        }
    }
}'
```

### List Indices
format will be `"index-name": "alias"`
```shell

curl --location 'http://localhost/indices/'
```

### Get Index(alias) Details/Mappings
```shell

curl --location 'http://localhost:8080/indices/sample-index/info'
```


### Check if index/alias Exists
```shell

curl --location 'http://localhost:8080/indices/sample-index/exists'
```


### Delete Index
```shell

curl --location --request DELETE 'http://localhost:8080/indices/sample-index'
```


### Update Index
```shell

curl --location --request PUT 'localhost:8080/indices/sample-index' \
--header 'Content-Type: application/json' \
--data '{
    "fields": {
        "id": {
            "type": "integer"
        },
        "name": {
            "type": "text",
            "autocomplete": true,
            "search": true
        },
        "price": {
            "type": "float"
        },
        "is_available": {
            "type": "boolean"
        },
        "description": {
            "type": "text",
            "analyzer": "english"
        },
        "metadata": {
            "type": "nested",
            "properties": {
                "tags": {
                    "type": "keyword"
                },
                "author": {
                    "type": "text"
                }
            }
        }
    }
}'
```

# Documents
* note - `sample-index` is more like alias than actual index in elastic search
## Add Batch Documents
```shell

curl --location 'http://localhost:8080/documents/sample-index/add' \
--header 'Content-Type: application/json' \
--data '{
    "data": [
        {
            "id": 1,
            "name": "Product A",
            "price": 19.99,
            "is_available": true,
            "description": "A high-quality product",
            "metadata": {
                "tags": ["electronics", "gadgets"],
                "author": "John Doe"
            }
        },
        {
            "id": 2,
            "name": "Product B",
            "price": 29.99,
            "is_available": false,
            "description": "Another great product",
            "metadata": {
                "tags": ["home", "kitchen"],
                "author": "Jane Smith"
            }
        }
    ]
}'
```

### Search Documents
```shell

curl --location 'http://localhost:8080/documents/sample-index/search' \
--header 'Content-Type: application/json' \
--data '{
    "query": "Product A",
    "filters": {
        "is_available": true
    },
    "pagination": {
        "from": 0,
        "size": 10
    },
    "min_score": 1.0,
    "search_fields": ["name", "description"]
}'
```

### List Documents
```shell

curl --location 'http://localhost:8080/documents/sample-index'
```

### Update Documents
```shell

curl --location --request PUT 'http://localhost:8080/documents/sample-index/HqSJt5QBZZRaRGGCePfg' \
--header 'Content-Type: application/json' \
--data '{
    "data": {
        "name": "Updated Product A",
        "price": 25.99
    }
}'
```

### Delete Document
```shell

curl --location --request DELETE 'http://localhost:8080/documents/sample-index/rMbtDZQBKigOlnkKmMyA'
```

### Get Document By ID
```shell

curl --location 'http://localhost:8080/documents/sample-index/HqSJt5QBZZRaRGGCePfg'
```


## Suggest / Auto-Complete
```shell

curl --location 'http://localhost:8080/documents/sample-index/suggest' \
--header 'Content-Type: application/json' \
--data '{
    "field":"name",
    "input": "Product A"
}'
```

### Export Documents
* if bulk is set to true, it will return documents in bulk format (ndjson)
```shell

curl --location --request POST 'localhost:8080/documents/sample-index/export?bulk=true'
```

### Import Documents
with file if file format is `ndjson use` `bulk=true`
```shell

curl --location 'localhost:8080/documents/sample-index/import' \
--form 'index="sample-bulk-json-ndjson"' \
--form 'file=@"/C:/Users/HP/Documents/Projects/go-es/bulk.ndjson"' \
--form 'bulk="true"'
```

OR pure `json`
```shell

curl --location 'localhost:8080/documents/sample-index/import' \
--form 'index="sample-bulk-json-ndjson"' \
--form 'json="{\"index\":{\"_index\":\"sample-pure-json-file\"}}
{\"description\":\"A high-quality product\",\"id\":1,\"is_available\":true,\"metadata\":{\"author\":\"John Doe\",\"tags\":[\"electronics\",\"gadgets\"]},\"name\":\"Updated Product A\",\"price\":25.99}
{\"index\":{\"_index\":\"sample-pure-json-file\"}}
{\"description\":\"Another great product\",\"id\":2,\"is_available\":false,\"metadata\":{\"author\":\"Jane Smith\",\"tags\":[\"home\",\"kitchen\"]},\"name\":\"Product B\",\"price\":29.99}
"' \
--form 'file=@"/C:/Users/HP/Documents/Projects/go-es/bulk.ndjson"'
```