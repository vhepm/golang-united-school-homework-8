package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Users []User

func ReadFile() (Users, error) {
	content, err := ioutil.ReadFile("./users.json")
	var payload []User

	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	return payload, err
}

func parseArgs() Arguments {
	id := flag.String("id", "", "id of the user")
	item := flag.String("item", "", "item to add")
	operation := flag.String(
		"operation",
		"",
		"operation to perform on users",
	)
	fileName := flag.String(
		"fileName",
		"users.json",
		"file storage",
	)

	flag.Parse()

	fmt.Println("id:", *id)
	fmt.Println("item:", *item)
	fmt.Println("operation:", *operation)
	fmt.Println("file name:", *fileName)

	args := make(Arguments)

	args["id"] = *id
	args["item"] = *item
	args["operation"] = *operation
	args["fileName"] = *fileName

	return args
}

func Add(users Users, user string, file string) {
	data := User{}
	err := json.Unmarshal([]byte(user), &data)

	users = append(users, data)
	payload, _ := json.Marshal(users)

	err = ioutil.WriteFile(file, payload, os.ModeAppend)
	if err != nil {
		return
	}

}

func List(users Users) Users {
	return users
}

func GetById(users Users, id string) (User, bool) {
	var ok bool
	var user_ User

	for _, user := range users {
		if user.Id == id {
			ok = true
			user_ = user
		}
	}

	if ok != true {
		return User{}, ok
	}

	return user_, ok
}

func FindById(users Users, id string, writer io.Writer) {
	enc := json.NewEncoder(writer)
	user, ok := GetById(users, id)

	if ok == true {
		if err := enc.Encode(user); err != nil {
			panic(err)
		}
	} else {
		_, err := writer.Write([]byte(""))
		if err != nil {
			return
		}
	}
}

func Remove_(u User, users Users) Users {
	for idx, v := range users {
		if v == u {
			return append(users[0:idx], users[idx+1:]...)
		}
	}
	return users
}

func Remove(users Users, id string, writer io.Writer, file string) {
	user, ok := GetById(users, id)
	if ok == true {
		users = Remove_(user, users)
		payload, _ := json.Marshal(users)
		err := ioutil.WriteFile(file, payload, os.ModeAppend)
		if err != nil {
			_, err := writer.Write([]byte("Unable to delete"))
			if err != nil {
				return
			}
		}
	} else {
		_, err := writer.Write([]byte("Item with id 2 not found"))
		if err != nil {
			return
		}
	}
}

func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func Perform(args Arguments, writer io.Writer) error {
	if len(strings.TrimSpace(args["fileName"])) == 0 {
		return errors.New("-fileName flag has to be specified")
	}
	if len(strings.TrimSpace(args["operation"])) == 0 {
		return errors.New("-operation flag has to be specified")
	}
	if args["operation"] == "add" && len(strings.TrimSpace(args["item"])) == 0 {
		return errors.New("-item flag has to be specified")
	}
	if (args["operation"] == "remove" || args["operation"] == "findById") && len(strings.TrimSpace(args["id"])) == 0 {
		return errors.New("-id flag has to be specified")
	}

	usersFile := args["fileName"]
	exists, _ := Exists(usersFile)
	if !exists {
		return errors.New("file doesn't exist")
	}

	users, _ := ReadFile()

	switch args["operation"] {

	case "add":
		Add(users, args["item"], usersFile)
	case "list":
		List(users)
	case "findById":
		FindById(users, args["id"], writer)
	case "remove":
		Remove(users, args["id"], writer, usersFile)
	default:
		return errors.New("Operation " + args["operation"] + " not allowed!")
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
