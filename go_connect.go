package main

///////////////////////////////////////////////////
// Depndencies
import (
	"fmt"
  "io"
  "io/ioutil"
	"net/http"
  "os"
  "log"
  "strings"
  "encoding/xml"
  "reflect"
  "sort"
  "strconv"
  "./libs"
  "github.com/spf13/viper"
  "syscall"
  "time"
  "runtime"
  "sync"
)  

///////////////////////////////////////////////////
// Strunct data for xml file
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

///////////////////////////////////////////////////
// Log error and exit
func check(e error) {
  if e != nil {
    log.Fatal(e)
  }
}

func logError(e error) {
  if e != nil {
    log.Println(e.Error())
  }
}

///////////////////////////////////////////////////
// Check DB file age
func checkDBFileAge() bool{

  var st syscall.Stat_t
  err := syscall.Stat(viper.GetString("db.path"), &st)
  check(err)
  
  status := (time.Now().Unix() - st.Mtimespec.Sec) > viper.GetInt64("db.max_age")
  return status

}

///////////////////////////////////////////////////
// Download DB
func downloadData(wg *sync.WaitGroup) {
  
  defer wg.Done()

  url := viper.GetString("db.url")
  fileName := viper.GetString("db.path")
  tmp_fileName := fileName + ".tmp"
	

  _, err := os.Stat(tmp_fileName)
  if err != nil {
    log.Println("Downloading", url, "to", tmp_fileName)

    var netClient = &http.Client{
      Timeout: time.Second * viper.GetDuration("db.url_timeout"),
    }
    
    response, err := netClient.Get(url)
    if err == nil {
      
      defer response.Body.Close()
      output, err := os.Create(tmp_fileName)
      
      if err == nil {

        defer output.Close()
        n, err := io.Copy(output, response.Body)
        check(err)
      
        log.Println(n, "bytes downloaded.")
        
        err = os.Rename(tmp_fileName, fileName)
        check(err)

      } else {
        log.Println("here")
        log.Println(err.Error())  
      }
    } else {
      log.Println(err.Error())
    }

  } else {
    log.Println("Skip, already downloading!")
  }
  
}

///////////////////////////////////////////////////
// Check if string is numeric
func IsNumeric(s string) bool {
  _, err := strconv.ParseFloat(s, 64)
  return err == nil
}

///////////////////////////////////////////////////
// Main block
func main() {
  runtime.GOMAXPROCS(1)
  var wg sync.WaitGroup
// Config file name and path
  viper.SetConfigName("config")
  viper.AddConfigPath(".")
  
// Find and read the config file
  err := viper.ReadInConfig() 
  check(err)

// Get arguments
  args := os.Args
  
// Read xml file
  data, err := ioutil.ReadFile(viper.GetString("db.path"))
  if err != nil {
    log.Println("Cant read file: " + viper.GetString("db.path"))
    wg.Add(1)
    downloadData(&wg)
  }

// Force DB download or print help when wrong parameters  
  if len(args) < 3 {
    if len(args) == 2 && args[1] == "download" {
      wg.Add(1)
      downloadData(&wg)
    } else {
      fmt.Println("Help ")
    }
    os.Exit(0)
  }

// Force download if DB reach max_file age 
  if(checkDBFileAge()){
    wg.Add(1)
    go downloadData(&wg)
  }

  data_unmarsh := Project{}
  err = xml.Unmarshal([]byte(data), &data_unmarsh)
  check(err)

// Covert arg[1] to Title (first char uper case) to match Node struct
  args[1] = strings.Title(args[1])

// Get and sort xml values by query arguments
  data_array := make(map[string]string)
  keys := []string{}
  
  for  _, value := range data_unmarsh.Nodes {
// reflect used to use method dinamicaly
    s := reflect.ValueOf(value)
    f := s.FieldByName(args[1]).Interface().(string)

// populate map and keys if string found
    if strings.Contains(f, args[2]){
      data_array[value.Desc] = value.Ip
      keys = append(keys, value.Desc)  
    }
  }
// Sort keys 
  sort.Strings(keys)
   
// Initializa flux variables
  count := 1
  cssh_string := ""
  
// Run flux based on option
  for _, val := range keys {
// No index in arguments - print list
    if len(args) < 4 {
      fmt.Println(count, " - ", val," :: ", data_array[val])
      
    } else { 
// Numeric argument 
      if IsNumeric(args[3]) { 
        index, err := strconv.Atoi(args[3])
        check(err)
// If match, connect and exit
        if index == count {
          connect.SshConn(data_array[val])
          os.Exit(0)
        }
// Literal - connect all list 
      } else if args[3] == "all" {
        connect.SshConn(data_array[val])
      
// Literal - create cssh string
      } else if args[3] == "cssh" {
        cssh_string += connect.Username() + "@" + data_array[val] + " "
      }
    }
    count++
  }
// Run cssh is exist and list is not empty
  if len(cssh_string) > 0 {
    if (viper.GetBool("cssh.enable")){ 
      fmt.Println("cssh " + cssh_string)
    } else {
      fmt.Println("cssh disabled!")
    }
    
  }
  wg.Wait()
}
