# Go-ES
A Full-Text Search Golang Microservice integrated with Elastic Search

## Overview
Go-ES is a microservice written in Go that provides a full-text search functionality
using Elastic Search. It allows users to perform various operations such as indexing,
searching, and retrieving documents. This is a for more simpler usage than elastic for whom
do not want to dive deep into elastic search echo system. For advance usage, use elastic search
sdks or rest api.

## Notice
The indices used for documents are more like aliases than indices in elasticsearch. In the backend we create an index with
datetime in the index name (the one users sent) and create an alias with exact index name as users requests.

## Features
* Integration with Elastic Search for full-text search capabilities
* Support for indexing and searching documents
* Retrieval of documents based on search queries
* Error handling and logging mechanisms

## Requirements
* Elastic Search 8.17.1 (other version hasn't been tested against yet)
* Elastic Search API Key

## Installation
### From Release
To install Go-ES, following these steps:
1. Go to the release page and download your os and architecture
2. Copy and modify the [config.yaml](config.yaml) file
3. Run the downloaded executable via `./go-es --config=config.yaml`


### From Source
To install Go-ES, follow these steps:
1. Clone the repository: `https://github.com/MeowSaiGithub/go-es.git`
2. Navigate to the project directory: `cd go-es`
3. Build the project: `go build -o go-es`
4. Modify the Configuration file [config.yaml](config.yaml) to suit your needs
5. Run the project: `./go-es --config config.yaml`

### From Docker
To install Elastic Search using Docker - https://www.elastic.co/guide/en/elasticsearch/reference/current/docker.html


To install Go-ES using Docker, follow these steps:
1. Clone the repository: `https://github.com/MeowSaiGithub/go-es.git`
2. Navigate to the project directory: `cd go-es`
3. Modify the Configuration file [config.yaml](config.yaml) to suit your needs
4. Build the Docker image: `docker build -t go-es .`
5. Run the Docker container: `docker run --network elastic -p 8080:8080 go-es`

## Configuration
The project uses a configuration file `config.yaml` to store settings for Elastic Search and other dependencies.
You can modify this file to suit your needs. You can see the example in the config for references.

## API Endpoints
The project provides several API endpoints for interacting with the full-text search functionality. Please see [sample-request.md](sample-request.md) for references.

## License
Go-ES is licensed under the Apache License, Version 2.0. See the LICENSE file for more information.

## Contributing
Contributions are welcome! If you'd like to contribute to Go-ES, please fork the repository and submit a pull request with your changes.