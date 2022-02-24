package OAPIClientGenerator

import (
	"fmt"
	"os"
	"strings"
)

const headerImports = `#pragma once

#include "CoreMinimal.h"
#include "GameFramework/Actor.h"
#include "Runtime/Online/HTTP/Public/Http.h"
#include "Runtime/JsonUtilities/Public/JsonObjectConverter.h"
#include "%s.generated.h"
`

const headerBase = `
UCLASS()
class %s %s : public AActor
{
	GENERATED_BODY()
	
public:
	FHttpModule* Http;

	// Sets default values for this actor's properties
	%s();

`
const headerEnd = `
};
`
const classInclude = `
#include "%s"
`
const classBase = `
%s::%s()
{
	Http = &FHttpModule::Get();
}
`
const requestFuncHeader = `
void %s::%s(%s)
{
`
const requestFuncBody = `	TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = Http->CreateRequest();
	Request->OnProcessRequestComplete().BindUObject(this, &%s::%s);
	//This is the url on which to process the request
	Request->SetURL("%s");
	Request->SetVerb("%s");
	Request->SetHeader(TEXT("User-Agent"), "X-UnrealEngine-Agent");
	Request->SetHeader("Content-Type", TEXT("application/json"));
`
const requestParameter = `
	TSharedPtr<FJsonObject> %sJsonObject = FJsonObjectConverter::UStructToJsonObject<%s>(%s);
	FString %sContentString;
	TSharedRef< TJsonWriter<> > %sWriter = TJsonWriterFactory<>::Create(&%sContentString);
	FJsonSerializer::Serialize(%sJsonObject.ToSharedRef(), %sWriter);
	Request->SetContentAsString(%sContentString);
`
const requestFuncEnd = `
	Request->ProcessRequest();
}`
const responseFuncBody = `
void %s::%s(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
`
const responseFuncBodyEnd = `
}
`

// int32 recievedInt = JsonObject->GetIntegerField("customInt");
const responseFuncArguments = `			result.%s = JsonObject->Get%sField("%s");
`
const responseCheck = `	if (Response->GetResponseCode() == %s) {
`
const responseCheckElse = `	else {
`
const responseCheckEnd = `	}
`
const resultCreation = `			%s result;
`
const resultArrayCreation = `			TArray<%s> result;
`
const responseSingleObject = `		TSharedPtr<FJsonObject> JsonObject;
		TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(Response->GetContentAsString());
		if (FJsonSerializer::Deserialize(Reader, JsonObject))
		{
`
const responseSingleObjectEnd = `		}
`
const responseArray = `		TArray<TSharedPtr<FJsonValue>> JsonArray;
		TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(Response->GetContentAsString());
		if (FJsonSerializer::Deserialize(Reader, JsonArray)) {
			for (int i = 0; i < JsonArray.Num(); i++) {
`
const responseArrayLoopEnd = `			}
`
const responseArrayEnd = `		}
`
const responseFuncCreateItem = `			%s Item;
`
const responseFuncArrayItems = `			Item.%s = JsonArray[i]->AsObject()->Get%sField("%s");
`
const responseFuncCreateItemEnd = `			result.Add(Item);
`

func withUEClassPrefix(name string) string {
	return "A" + name
}

func GenerateHeader(projectName, className, exportPath string, oapi OAPI) error {
	headerContent := fmt.Sprintf(headerImports, className)

	// generate definitions
	for definitionName, definition := range oapi.Definitions {
		cppStructure := definition.generateCppStructure(definitionName)
		if cppStructure != "" {
			headerContent += cppStructure
		}
	}
	headerContent += fmt.Sprintf(headerBase, strings.ToUpper(projectName)+"_API", withUEClassPrefix(className), withUEClassPrefix(className))

	// generate functions
	for path, methods := range oapi.Paths {
		for method, endpoint := range methods {
			funcName, pathArgs := getFuncName(path, method)
			parameters := getParameters(endpoint.Parameters)
			headerContent += "\n\tUFUNCTION(BlueprintCallable, Category = OAPI)"
			headerContent += "\n\tvoid " + funcName + "(" + strings.Join(append(pathArgs, parameters...), ",") + ");"                                   //todo add params
			headerContent += "\n\tvoid " + getResponseFuncName(funcName) + "(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);" //todo return values?
			for responseCode, response := range endpoint.Responses {
				definitionName := refToDefName(response.Schema.Ref)
				definition := oapi.Definitions[definitionName]
				if responseCode == "default" {
					responseCode = "Error"
				}
				headerContent += "\n\tUFUNCTION(BlueprintImplementableEvent, Category = OAPI)"
				switch {
				case definitionName == "":
					headerContent += "\n\tvoid " + getResponseFuncName(funcName) + responseCode + "();"
				case definition.Type == "array":
					headerContent += "\n\tvoid " + getResponseFuncName(funcName) + responseCode + "(const TArray<" + withUEStructPrefix(refToDefName(definition.Items.Ref)) + "> &Result);"
				case definition.Type == "object":
					headerContent += "\n\tvoid " + getResponseFuncName(funcName) + responseCode + "(" + withUEStructPrefix(definitionName) + " Result);"
				}
			}
		}
	}
	headerContent += `
	UFUNCTION(BlueprintImplementableEvent, Category = OAPI)
	void OnOapiError(const FString &text);`

	headerContent += headerEnd
	return os.WriteFile(exportPath+className+".h", []byte(headerContent), 0644)
}

func getParameters(parameters []OAPIParameter) []string {
	result := []string{}
	for _, parameter := range parameters {
		if parameter.Ref != "" {
			defName := refToDefName(parameter.Ref)
			name := parameter.Name
			if name == "" {
				name = strings.ToLower(defName)
			}
			result = append(result, withUEStructPrefix(defName)+" "+name)
		}
	}
	return result
}

func getFuncName(path, method string) (string, []string) {
	pathArgs := []string{}
	pathCrumbs := strings.Split(path, "/")
	funcNameParts := []string{}
	for i := 0; i < len(pathCrumbs); i++ {
		if pathCrumbs[i] == "" {
			continue
		}
		if strings.HasPrefix(pathCrumbs[i], "{") && strings.HasSuffix(pathCrumbs[i], "}") {
			pathArg := strings.TrimSuffix(strings.TrimPrefix(pathCrumbs[i], "{"), "}")
			pathArgs = append(pathArgs, "FString "+pathArg)
			funcNameParts = append(funcNameParts, "By"+strings.Title(pathArg))
			continue
		}
		funcNameParts = append(funcNameParts, strings.Title(pathCrumbs[i]))
	}
	return strings.Title(method) + strings.Join(funcNameParts, ""), pathArgs
}

func getResponseFuncName(funcName string) string {
	return "On" + funcName + "Response"
}

func GenerateClass(className, exportPath string, oapi OAPI) error {
	classContent := fmt.Sprintf(classInclude, className+".h")

	// generate constructor
	classContent += fmt.Sprintf(classBase, withUEClassPrefix(className), withUEClassPrefix(className))

	// generate functions
	for path, methods := range oapi.Paths {
		for method, endpoint := range methods {
			funcName, pathArgs := getFuncName(path, method)
			parameters := getParameters(endpoint.Parameters)
			responseFuncName := getResponseFuncName(funcName)
			url := getUrlWithParameters(oapi.Host, oapi.BasePath, path)
			classContent += fmt.Sprintf(requestFuncHeader, withUEClassPrefix(className), funcName, strings.Join(append(pathArgs, parameters...), ","))
			classContent += fmt.Sprintf(requestFuncBody, withUEClassPrefix(className), responseFuncName, url, method)
			for _, parameter := range parameters {
				split := strings.Split(parameter, " ")
				parameterName := split[1]
				parameterType := split[0]
				classContent += fmt.Sprintf(requestParameter, parameterName, parameterType, parameterName, parameterName, parameterName, parameterName, parameterName, parameterName, parameterName)
			}
			classContent += fmt.Sprint(requestFuncEnd)
			classContent += fmt.Sprintf(responseFuncBody, withUEClassPrefix(className), responseFuncName)
			var defaultResponse *OAPIResponse
			for responseCode, response := range endpoint.Responses {
				if responseCode == "default" {
					defaultResponse = &response
					continue
				}
				classContent += fmt.Sprintf(responseCheck, responseCode)
				classContent += getResponsePart(responseCode, responseFuncName, response, oapi)
				classContent += fmt.Sprint(responseCheckEnd)
			}
			if defaultResponse != nil {
				classContent += fmt.Sprint(responseCheckElse)
				classContent += getResponsePart("default", responseFuncName, *defaultResponse, oapi)
				classContent += fmt.Sprint(responseCheckEnd)
				classContent += fmt.Sprintf(`	OnOapiError("` + funcName + ` error");`)
			}
			classContent += responseFuncBodyEnd
		}
	}
	return os.WriteFile(exportPath+className+".cpp", []byte(classContent), 0644)
}

func getUrlWithParameters(host, basePath, path string) string {
	url := host + basePath + path
	split := strings.Split(url, "/")
	for i := 0; i < len(split); i++ {
		if strings.HasPrefix(split[i], "{") && strings.HasSuffix(split[i], "}") {
			split[i] = `"+` + strings.TrimSuffix(strings.TrimPrefix(split[i], "{"), "}") + `+"`
		}
	}
	return strings.Join(split, "/")
}

func refToDefName(ref string) string {
	return strings.TrimPrefix(ref, "#/definitions/")
}

func getResponsePart(responseCode, responseFuncName string, response OAPIResponse, oapi OAPI) string {
	result := ""
	if responseCode == "default" {
		responseCode = "Error"
	}
	definitionName := refToDefName(response.Schema.Ref)
	if definitionName != "" {
		if definition, ok := oapi.Definitions[definitionName]; ok {
			if definition.Type == "object" {
				result += fmt.Sprintf(resultCreation, withUEStructPrefix(definitionName))
				result += fmt.Sprint(responseSingleObject)
				for propertyName, property := range definition.Properties {
					result += fmt.Sprintf(responseFuncArguments, strings.Title(propertyName), getJsonType(getCppType(property.Type, property.Format)), propertyName) //todo arguments
				}
				result += `		` + responseFuncName + responseCode + `(result);`
				result += `		` + "return;"
				result += fmt.Sprint(responseSingleObjectEnd)
			} else if definition.Type == "array" {
				itemName := refToDefName(definition.Items.Ref)
				result += fmt.Sprintf(resultArrayCreation, withUEStructPrefix(itemName))
				result += fmt.Sprint(responseArray)
				var item = oapi.Definitions[itemName]
				result += fmt.Sprintf(responseFuncCreateItem, withUEStructPrefix(itemName))
				for propertyName, property := range item.Properties {
					result += fmt.Sprintf(responseFuncArrayItems, strings.Title(propertyName), getJsonType(getCppType(property.Type, property.Format)), propertyName) //todo arguments
				}
				result += fmt.Sprint(responseFuncCreateItemEnd)
				result += fmt.Sprint(responseArrayLoopEnd)
				result += `		` + responseFuncName + responseCode + `(result);`
				result += `		` + "return;"
				result += fmt.Sprint(responseArrayEnd)
			}
		} else {
			//error
		}
	} else {
		result += `		` + responseFuncName + responseCode + `();
`
	}
	return result
}

func getJsonType(cppType string) string {
	switch cppType {
	case "int32":
		fallthrough
	case "int":
		return "Integer"
	case "FString":
		return "String"
	}
	return ""
}
