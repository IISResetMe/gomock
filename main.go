package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/rs/xid"
)

func generateEmbeddedResource(tmpl *ActionTemplate) (*jen.Statement, error) {
	return nil, nil
}

func generateWebRequest(tmpl *ActionTemplate) (*jen.Statement, error) {
	var actionModuleName = "web"
	var operationNames = []string{"head", "get", "post"}

	if actionModuleName != tmpl.Module {
		return nil, errors.New("Can only generate file operations")
	}

	if !contains(operationNames, tmpl.Operation) {
		return nil, errors.New(fmt.Sprintf("Can only generate one of the following file operations: %v", operationNames))
	}

	return nil, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (op *FileOperation) populateFileFields(ac *ActionTemplate) error {
	if err := op.populateActionBlock(ac); err != nil {
		return err
	}

	if filePath, ok := ac.Parameters["path"]; !ok {
		return errors.New("Missing mandatory parameter 'path' in FileStatOperation")
	} else {
		op.Path = filePath
	}
	return nil
}

func (op *FileIOOperation) populateFileIOFields(ac *ActionTemplate) error {
	if err := op.populateFileFields(ac); err != nil {
		return err
	}

	if offsetString, ok := ac.Parameters["delay"]; ok {
		if offset, err := strconv.Atoi(offsetString); err != nil {
			return err
		} else if offset < 0 {
			return errors.New("Expected non-negative integer for 'offset'")
		} else {
			op.Offset = offset
		}
	} else {
		op.Offset = 0
	}

	if lengthString, ok := ac.Parameters["repeat"]; ok {
		if length, err := strconv.Atoi(lengthString); err != nil {
			return err
		} else if length < 0 {
			return errors.New("Expected non-negative integer for 'length'")
		} else {
			op.Length = length
		}
	} else {
		op.Length = 0
	}

	return nil
}

func (op *fileWriteOperation) populateFileWriteFields(ac *ActionTemplate) error {
	if err := op.populateFileIOFields(ac); err != nil {
		return err
	}

	if content, ok := ac.Parameters["content"]; !ok {
		return errors.New("Expected string value for 'content'")
	} else {
		op.Content = content
	}

	return nil
}

func generateFileOperation(tmpl *ActionTemplate) (*jen.Statement, error) {
	var actionModuleName = "file"
	var operationNames = []string{"stat", "read", "write"}

	if actionModuleName != tmpl.Module {
		return nil, errors.New("Can only generate file operations")
	}

	if !contains(operationNames, tmpl.Operation) {
		return nil, errors.New(fmt.Sprintf("Can only generate one of the following file operations: %v", operationNames))
	}

	stmt := new(jen.Statement)

	switch tmpl.Operation {
	case "stat":
		op := new(FileStatOperation)
		op.populateFileFields(tmpl)
		if code, err := generateFileStatOperation(op); err != nil {
			return nil, err
		} else {
			stmt.Add(code)
		}

	case "read":
		op := new(FileReadOperation)
		op.populateFileIOFields(tmpl)
		if code, err := generateFileReadOperation(op); err != nil {
			return nil, err
		} else {
			stmt.Add(code)
		}

	case "write":
		op := new(fileWriteOperation)
		op.populateFileWriteFields(tmpl)
		if code, err := generateFileWriteOperation(op); err != nil {
			return nil, err
		} else {
			stmt.Add(code)
		}
	}

	return stmt, nil
}

func generateFileStatOperation(op *FileStatOperation) (*jen.Statement, error) {
	stmt := new(jen.Statement)
	opName := "fileStat_" + xid.New().String()
	stmt.Add(jen.Comment("Operation: " + opName).Line())
	stmt.Add(jen.Id("_").Op(":=").Qual("os", "Stat").Call(
		jen.Lit(op.Path),
	)).Line()

	return stmt, nil
}

func generateFileReadOperation(op *FileReadOperation) (*jen.Statement, error) {
	return nil, nil
}

func generateFileWriteOperation(op *fileWriteOperation) (*jen.Statement, error) {
	return nil, nil
}

func (ab *ActionBlock) populateActionBlock(tmpl *ActionTemplate) error {
	if delayString, ok := tmpl.Parameters["delay"]; ok {
		if delay, err := strconv.Atoi(delayString); err != nil {
			return err
		} else if delay < 0 {
			return errors.New("Expected non-negative integer for 'delay'")
		} else {
			ab.Delay = delay
		}
	} else {
		ab.Delay = 0
	}

	if repeatString, ok := tmpl.Parameters["repeat"]; ok {
		if repeat, err := strconv.Atoi(repeatString); err != nil {
			return err
		} else if repeat < 0 {
			return errors.New("Expected non-negative integer for 'repeat'")
		} else {
			ab.Repeat = repeat
		}
	} else {
		ab.Repeat = 0
	}

	return nil
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		log.Fatal("Please provide a module name")
	}
	name := strings.Replace(args[0], " ", "", -1)

	if len(args) < 2 {
		log.Fatal("Please provide a json file path")
	}

	inpath := args[1]
	jsonFile, err := os.Open(inpath)
	if errors.Is(err, os.ErrNotExist) {
		log.Fatal("Please provide a path to an existing json file as the second argument")
	} else if err != nil {
		log.Fatalf("Error encountered while accessing file %q: %q", inpath, err.Error())
	}

	f := *jen.NewFile(name)

	fmt.Println("Decoding JSON config")
	decoder := json.NewDecoder(jsonFile)

	var newProgram ProgramTemplate
	decoder.Decode(&newProgram)

	fmt.Println("Generating source code")
	m := f.Func().Id("main").Params()
	stmt := new(jen.Statement)
	for i := range newProgram.Blocks {
		block := newProgram.Blocks[i]

		switch block.Module {
		case "file":
			fmt.Println("Generating file operation")
			if code, err := generateFileOperation(&block); err != nil {
				log.Fatalf("Encountered error: %q", err)
			} else {
				stmt.Add(code)
			}

		case "web":
			fmt.Println("Generating web operation")
			if code, err := generateWebRequest(&block); err != nil {
				log.Fatalf("Encountered error: %q", err)
			} else {
				stmt.Add(code)
			}

		case "embed":
			fmt.Println("Generating embed operation")
			if code, err := generateEmbeddedResource(&block); err != nil {
				log.Fatalf("Encountered error: %q", err)
			} else {
				stmt.Add(code)
			}

		default:
		}
	}

	m.Block(stmt)

	fmt.Println("Rendering source code")
	outBuffer := &bytes.Buffer{}
	err = f.Render(outBuffer)
	if err != nil {
		log.Fatalf("Encountered error: %q", err)
	}

	fmt.Println("Here comes ya program: ")
	fmt.Println(string(outBuffer.Bytes()[:]))
	f.Save("C:\\users\\mail\\test\\test.go")
}

type ProgramTemplate struct {
	Name     string
	Metadata ProgramMetadata
	Blocks   []ActionTemplate
}

type ProgramMetadata struct {
	Author string
	Notes  string
}

type ActionTemplate struct {
	Module     string
	Operation  string
	Parameters map[string]string
}

type ActionBlock struct {
	Delay  int
	Repeat int
}

type EmbedDemand struct {
	Content string
	ActionBlock
}

type WebRequest struct {
	Uri        string
	RetryCount int
	ActionBlock
}

type FileOperation struct {
	Path string
	ActionBlock
}

type FileStatOperation struct {
	FileOperation
}

type FileIOOperation struct {
	Offset int
	Length int
	FileOperation
}

type FileReadOperation struct {
	FileIOOperation
}

type fileWriteOperation struct {
	Content string
	FileIOOperation
}
