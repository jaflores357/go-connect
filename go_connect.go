package main

import (
	"fmt"
  "io"
  "io/ioutil"
	"net/http"
	"os"
  "strings"
  "encoding/xml"
  "reflect"
  "sort"
  "strconv"
  "./libs"
)  



type Project struct {
	Nodes    []Node    `xml:"node"`
}

type Node struct {
  Name        string `xml:"name,attr"`
  Desc        string `xml:"description,attr"`
  OsArch      string `xml:"osArch,attr"`
  OsFamily    string `xml:"osFamily,attr"`
  OsName      string `xml:"osName,attr"`
  OsVersion   string `xml:"osVersion,attr"`
  Roles       string `xml:"roles,attr"`
  Env         string `xml:"environment,attr"`
  Ip          string `xml:"hostname,attr"`
}

func check(e error) {
  if e != nil {
      panic(e)
  }
}

func downloadFromUrl(url string) {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}

func IsNumeric(s string) bool {
  _, err := strconv.ParseFloat(s, 64)
  return err == nil
}


func main() {

  // Get arguments
  args := os.Args
  
  // Download xml data
  //downloadFromUrl("http://chef.zenvia360.com:4567/allnodes")

  // Read xml file
  data, err := ioutil.ReadFile("allnodes")
  check(err)

  v := Project{}
  err = xml.Unmarshal([]byte(data), &v)
  check(err)

  // Get and sort xml values by query arguments
  m := make(map[string]string)
  keys := []string{}
  
  for  _, value := range v.Nodes {
    // reflect used to use method dinamicaly
    s := reflect.ValueOf(value)
    f := s.FieldByName(args[1]).Interface().(string)

    // populate map and keys if string found
    if strings.Contains(f, args[2]){
      m[value.Desc] = value.Ip
      keys = append(keys, value.Desc)  
    }
  }
  // Sorte de keys 
  sort.Strings(keys)
  
  
  // Run flux based on option
  count := 1
  cssh_string := ""
  
  for _, val := range keys {
    if len(args) < 4 {
      fmt.Println(count, " - ", val," :: ", m[val])
      
    } else { 
      
      if IsNumeric(args[3]) { 
        index, err := strconv.Atoi(args[3])
        check(err)
        if index == count {
          
          connect.SshConn(m[val])
          //fmt.Println(count, " - ", val," :: ", m[val])
          os.Exit(0)
        }
      } else if args[3] == "all" {
        fmt.Println(count, " - ", val," :: ", m[val])
      } else if args[3] == "cssh" {
        cssh_string += " " + val + "@" + m[val]
      }
    }
    count++
  }
  if len(cssh_string) > 0 {
    fmt.Println("cssh " + cssh_string)
  }

}
