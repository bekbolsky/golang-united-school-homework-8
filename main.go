package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	ErrFileNameFlagMissing  = errors.New("-fileName flag has to be specified")
	ErrOperationFlagMissing = errors.New("-operation flag has to be specified")
	ErrItemFlagMissing      = errors.New("-item flag has to be specified")
	ErrIDFlagMissing        = errors.New("-id flag has to be specified")
)

// User struct represents user in the json file.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Arguments represents command line arguments.
// Keys are: operation, item, fileName, id.
type Arguments map[string]string

func parseArgs() Arguments {
	idFlag := flag.String("id", "", "id of the user")
	operationFlag := flag.String("operation", "", "add, list, findById, remove")
	itemFlag := flag.String("item", "", "item to add to the file")
	fileNameFlag := flag.String("fileName", "", "json file name")

	flag.Parse()
	return Arguments{
		"id":        *idFlag,
		"operation": *operationFlag,
		"item":      *itemFlag,
		"fileName":  *fileNameFlag,
	}
}

// addItem writes item to the file as JSON array.
// If file is empty, then user should be added to the file,
// otherwise user should be added to the end of the file.
// If user with specified id already exists in file,
// then error has to be returned.
func addItem(file *os.File, writer io.Writer, item string) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		_, err = writer.Write([]byte(item))
		if err != nil {
			return err
		}
		return nil
	}
	var itemArray []User
	err = json.Unmarshal(data, &itemArray)
	if err != nil {
		return err
	}
	var user User
	err = json.Unmarshal([]byte(item), &user)
	if err != nil {
		return err
	}
	for _, userItem := range itemArray {
		if userItem.ID == user.ID {
			_, err = writer.Write([]byte("Item with id " + user.ID + " already exists"))
			if err != nil {
				return err
			}
			return nil
		}
	}
	itemArray = append(itemArray, user)
	data, err = json.Marshal(itemArray)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// ListItems retrieves list from the file and write it to the io.Writer stream.
// Uses writer to print the result!
func listItems(file *os.File, writer io.Writer) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		_, err = writer.Write([]byte(""))
		if err != nil {
			return err
		}
		return nil
	}
	var item []User
	err = json.Unmarshal(data, &item)
	if err != nil {
		return err
	}
	data, err = json.Marshal(item)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// FindUserById finds user by id.
// If user with specified id does not exist in file,
// then empty string has to be written to the writer interface.
// If user exists, then json object should be written in writer interface.
// If file is empty, then nothing has to be written to the writer interface.
func findUserById(file *os.File, writer io.Writer, id string) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		_, err = writer.Write([]byte(""))
		if err != nil {
			return err
		}
		return nil
	}
	var item []User
	err = json.Unmarshal(data, &item)
	if err != nil {
		return err
	}
	for _, user := range item {
		if user.ID == id {
			data, err = json.Marshal(user)
			if err != nil {
				return err
			}
			_, err = writer.Write(data)
			if err != nil {
				return err
			}
			return nil
		}
	}
	_, err = writer.Write([]byte(""))
	if err != nil {
		return err
	}
	return nil
}

// removeUser removes user from the JSON array by id.
// If user with id 2 does not exist in file,
// it should print message to the io.Writer «Item with id 2 not found»
// Otherwise, user should be removed from the file.
func removeUser(file *os.File, writer io.Writer, id string) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		_, err = writer.Write([]byte(""))
		if err != nil {
			return err
		}
		return nil
	}

	var item []User
	err = json.Unmarshal(data, &item)
	if err != nil {
		return err
	}
	for i, user := range item {
		if user.ID == id {
			item = append(item[:i], item[i+1:]...)
			data, err = json.Marshal(item)
			if err != nil {
				return err
			}
			_, err = writer.Write(data)
			if err != nil {
				return err
			}
			return nil
		}
	}
	_, err = writer.Write([]byte("Item with id " + id + " not found"))
	if err != nil {
		return err
	}
	return nil
}

// Users list should be stored in the JSON file.
// When you start your application and tries to perform some operations,
// existing file should be used or new one should be created if it does not exist.
func Perform(args Arguments, writer io.Writer) error {
	fileName := args["fileName"]
	if fileName == "" {
		return ErrFileNameFlagMissing
	}

	operation := args["operation"]
	if operation == "" {
		return ErrOperationFlagMissing
	}

	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	switch operation {
	case "add":
		item := args["item"]
		if item == "" {
			return ErrItemFlagMissing
		}
		err = addItem(file, writer, item)
		if err != nil {
			return err
		}

	case "list":
		err = listItems(file, writer)
		if err != nil {
			return err
		}

	case "findById":
		id := args["id"]
		if id == "" {
			return ErrIDFlagMissing
		}
		err = findUserById(file, writer, id)
		if err != nil {
			return err
		}
	case "remove":
		id := args["id"]
		if id == "" {
			return ErrIDFlagMissing
		}
		err = removeUser(file, writer, id)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}
	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

// ./main -operation="add" -item="{\"id\":\"1\",\"email\":\"test@test.com\",\"age\":34}" -fileName="users.json"
