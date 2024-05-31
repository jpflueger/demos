SERVER ?= http://127.0.0.1:3000
CURL_ARGS ?= -i -H 'content-type: application/json'

.PHONY: test-list
test-list:
	curl $(CURL_ARGS) -X GET $(SERVER)

.PHONY: test-create
test-create:
	curl $(CURL_ARGS) -X POST -d '{"name":"Minimalist","data":{"color":"seafoam green","capacity":"128 GB"}}' $(SERVER)

.PHONY: test-update
test-update:
	curl $(CURL_ARGS) -X PUT -d '{"id":"$(ID)","name":"Minimalist","data":{"color":"oxford blue","capacity":"256 GB","Stuff":true}}' $(SERVER)/$(ID)
