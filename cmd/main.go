package main

import (
	"flag"
	"fmt"
	"log"

	"OAPIClientGenerator"
)

const help = `Usage: ./oapiClientGenerator -project=PROJECT_NAME -file=FILE.yaml -class=CLASS_NAME`

func main() {
	oapiFilePath := flag.String("file", "", "File containing OpenAPI definition")
	className := flag.String("class", "", "Name of the generated class")
	projectName := flag.String("project", "", "Name of the project")
	flag.Parse()
	if oapiFilePath == nil || *oapiFilePath == "" {
		log.Fatalln(help)
	}
	if className == nil || *className == "" {
		log.Fatalln(help)
	}
	if projectName == nil || *projectName == "" {
		log.Fatalln(help)
	}

	var oapi OAPIClientGenerator.OAPI
	err := oapi.Parse(*oapiFilePath)
	if err != nil {
		fmt.Println(err)
	}

	exportPath := "./export"
	err = OAPIClientGenerator.GenerateHeader(*projectName, *className, exportPath, oapi)
	if err != nil {
		fmt.Println(err)
	}
	err = OAPIClientGenerator.GenerateClass(*className, exportPath, oapi)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Move %s to your %s/Source/%s/Public/\n", *className+".h", *projectName, *projectName)
	fmt.Printf("Move %s to your %s/Source/%s/Private/\n\n", *className+".cpp", *projectName, *projectName)
	fmt.Println("!!! IMPORTANT !!!")
	fmt.Printf("Add \"Http\", \"Json\", \"JsonUtilities\" to PublicDependencyModuleNames in your %s/Source/%s/HTTPClientTest.Build.cs file with:\n\n", *projectName, *projectName)
}
