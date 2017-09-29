package main

///////////////////////////////////////////////////
// Depndencies
import (
	"fmt"
  "io"
  "io/ioutil"
  "net/http"
  "net"
  "os"
  "log"
  "strings"
  "encoding/xml"
  "reflect"
  "sort"
  "strconv"
  "./libs"
  "github.com/spf13/viper"
  "time"
  "path/filepath"
  "github.com/kardianos/osext"
)  

///////////////////////////////////////////////////
// Struct data for xml file
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
// Struct config file
type Config struct {
  General struct {
    LogFile string
  }
  DB struct {
    Api string
    RequestTimeout time.Duration
    ConnectionTimeout time.Duration
    Path string
    MaxAge int64
  }
  CSSH struct {
    Enable bool
    Path string
  }
}

///////////////////////////////////////////////////
// Help function
func help(executable string){

  helpString := `
  Search:
  -------

  `+executable+` <TYPE> <STRING>

  TYPE: [name|desc|osArch|osFamily|osName|osVersion|roles|env|ip]

  Ex.:
  `+executable+` name prd-sms-smpp
  `+executable+` desc prd-sms-smpp
  `+executable+` osArch x86_64
  `+executable+` osFamily unix
  `+executable+` osName centos
  `+executable+` osVersion 6.5
  `+executable+` roles td-agent
  `+executable+` env sms-production
  `+executable+` ip 10.0.2.41

  Connect:
  --------

  `+executable+` <TYPE> <STRING> < #Index | all | cssh >

  TYPE: [name|desc|osArch|osFamily|osName|osVersion|roles|env|ip]

  ID: Connect ssh specific host with ID listed in the Search
  all: Connect ssl ALL hosts listed in the Search
  cssh: If installed and have binary point in connect.conf file

  Ex.:

  `+executable+` name prd-pay-cm 1
  `+executable+` name prd-pay-cm all
  `+executable+` name prd-pay-cm cssh
  
  `
  fmt.Println(helpString)


}



///////////////////////////////////////////////////
// Log error and exit
func check(e error) {
  if e != nil {
    log.Fatal(e)
  }
}

///////////////////////////////////////////////////
// Check DB file age
func checkDBFileAge() bool{

  fileInfo, err := os.Lstat(cfg.DB.Path)
  check(err)
  
  status := (time.Now().Unix() - fileInfo.ModTime().Unix()) > cfg.DB.MaxAge
  return status

}

///////////////////////////////////////////////////
// Download DB
func downloadData() (error) {
  
  url := cfg.DB.Api
  fileName := cfg.DB.Path

  log.Println("Downloading", url, "to", fileName)

  _, err := net.DialTimeout("tcp", "chef.zenvia360.com:4567", 1 * time.Second )
  if err != nil {

      return err

  } else {

    netTransport := &http.Transport{
      DialContext: (&net.Dialer{
        Timeout: time.Second * cfg.DB.ConnectionTimeout,
      }).DialContext,
    }

    netClient := &http.Client{
      Transport: netTransport,
      Timeout: time.Second * cfg.DB.RequestTimeout,
    }
    
    response, err := netClient.Get(url)
    if err == nil {
      
      defer response.Body.Close()
      output, err := os.Create(fileName)
      check(err)
      
      defer output.Close()
      
      n, err := io.Copy(output, response.Body)
      check(err)
      
      log.Println(n, "bytes downloaded.")
      return nil

    } else {
      log.Println(err.Error())
      return err
    }
  }
}

///////////////////////////////////////////////////
// Check if string is numeric
func IsNumeric(s string) bool {
  _, err := strconv.ParseFloat(s, 64)
  return err == nil
}

///////////////////////////////////////////////////
// Global var
var cfg Config

///////////////////////////////////////////////////
// Main block
func main() {

// Application directory

  executable, err := osext.Executable(); check(err)
  linkpath, err := os.Readlink(executable);
  if err != nil{
    linkpath = executable
  }
  
  appFolder, err := filepath.Abs(filepath.Dir(linkpath)); check(err)
  
// Config file name and path
  viper.SetConfigName("config")
  viper.AddConfigPath(appFolder)

// Find and read the config file
  err = viper.ReadInConfig() 
  check(err)

// Unmarshal config to global cfg var  
  err = viper.Unmarshal(&cfg)
  check(err)
  
// Logfile
  logfile, err := os.OpenFile(cfg.General.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  check(err)

  defer logfile.Close() 
  log.SetOutput(logfile)


// Get arguments
  args := os.Args
  
// Read xml file
  data, err := ioutil.ReadFile(cfg.DB.Path)
  if err != nil {
    log.Println("Cant read file: " + cfg.DB.Path)
    err := downloadData()
    if err != nil {
      fmt.Println("Cant download nodes file, check "+cfg.General.LogFile+" for details!")    
      check(err)
    }
  }

// Force DB download or print help when wrong parameters  
  if len(args) < 3 {
    if len(args) == 2 && args[1] == "download" {
      err := downloadData(); check(err)
    } else {
      help(filepath.Base(executable))
    }
    os.Exit(0)
  }

// Force download if DB reach max_file age 
  if(checkDBFileAge()){
    err := downloadData(); check(err)
  }

// Unmarshal xml, download a new one if corrupt  
  data_unmarsh := Project{}
  err = xml.Unmarshal([]byte(data), &data_unmarsh)
  if err != nil {
    fmt.Println("Data file "+cfg.DB.Path+" corrupt. Downloading a new one")
    err := downloadData(); check(err)
  }

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
    if (cfg.CSSH.Enable){ 
      fmt.Println("cssh " + cssh_string)
    } else {
      fmt.Println("cssh disabled!")
    }
    
  }
}
