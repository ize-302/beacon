#!/bin/bash

API_SERVER="${API_SERVER:-http://localhost:8080/openapi.json}" # If variable not set or null, use default.

OUTPUT_FOLDER="./src/client"

rm -rf $OUTPUT_FOLDER

openapi-generator generate -i $API_SERVER \
	-g typescript-axios \
	-o $OUTPUT_FOLDER \
	--skip-validate-spec
