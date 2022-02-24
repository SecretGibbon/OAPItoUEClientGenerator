# OAPItoUEClientGenerator

Simple (not perfect, nor complete) generator that generates Blueprint ready UE classes from OpenAPI definitions.

## Known issues:
- Currently supports only integer and string types

## How to run
- go run cmd/main.go -project=[YOUR_UE_PROJECT_NAME] -file=[OAPI_DEFINITION_FILE] -class=[RESULTING_UE_CLASS_NAME]
- for example go run cmd/main.go -project=ue_test -file=petstore.yaml -class=petstore_api
