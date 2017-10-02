package main

///////////////////////////////////////////////////
// Depndencies
import (
	"fmt"
  "io"
  "io/ioutil"
  "net/http"
  "net"
  "net/url"
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
  "flag"
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
  Nodes struct {
    Api string
    RequestTimeout time.Duration
    ConnectionTimeout time.Duration
    NodesFile string
    MaxAge int64
    AutoUpdate bool
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
// Check Nodes file age
func checkNodesFileAge() bool{

  fileInfo, err := os.Lstat(cfg.Nodes.NodesFile)
  check(err)
  
  status := (time.Now().Unix() - fileInfo.ModTime().Unix()) > cfg.Nodes.MaxAge
  return status

}

///////////////////////////////////////////////////
// Download Nodes file
func downloadData() (error) {
  
  dbapi := cfg.Nodes.Api
  fileName := cfg.Nodes.NodesFile

  u, err := url.Parse(dbapi)
  hostTest := u.Host

// Teste host and port
  _, err = net.DialTimeout("tcp", hostTest, 1 * time.Second )
  if err != nil {

      return err

  } else {

    log.Println("Downloading", dbapi, "to", fileName)
    netTransport := &http.Transport{
      Dial: (&net.Dialer{
        Timeout: time.Second * cfg.Nodes.ConnectionTimeout,
      }).Dial,
    }

    netClient := &http.Client{
      Transport: netTransport,
      Timeout: time.Second * cfg.Nodes.RequestTimeout,
    }
    
    response, err := netClient.Get(dbapi)
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
  if err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }

// Unmarshal config to global cfg var  
  err = viper.Unmarshal(&cfg)
  check(err)
  
// Logfile
  logfile, err := os.OpenFile(cfg.General.LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  check(err)

  defer logfile.Close() 
  log.SetOutput(logfile)


// Get arguments
  keyOption := flag.String("i", "", "private key file")
  userOption := flag.String("u", "", "username")
  
  flag.Parse()
  args := flag.Args()
  
// Username and key
  sshkey := *keyOption
  username := *userOption
  if username == "" {
    username = connect.Username()
  }
  

// Read xml file
  data, err := ioutil.ReadFile(cfg.Nodes.NodesFile)
  if err != nil {
    fmt.Println("Cant read file: " + cfg.Nodes.NodesFile)
    err := downloadData()
    if err != nil {
      fmt.Println("Cant download nodes file, check "+cfg.General.LogFile+" for details!")    
      check(err)
    }
  }

// Force Nodes file download or print help when wrong parameters  
  if len(args) < 2 {
    if len(args) == 1 && args[0] == "download" {
      err := downloadData(); check(err)
    } else {
      help(filepath.Base(executable))
    }
    os.Exit(0)
  }

// Force download if Nodes file reach max_file age 
  if(checkNodesFileAge() && cfg.Nodes.AutoUpdate){
    err := downloadData()
    if err != nil {
      fmt.Println("Cant update nodes file, check "+cfg.General.LogFile+" for details!")
    }

  }

// Unmarshal xml, download a new one if corrupt  
  data_unmarsh := Project{}
  err = xml.Unmarshal([]byte(data), &data_unmarsh)
  if err != nil {
    fmt.Println("Data file "+cfg.Nodes.NodesFile+" corrupt. Downloading a new one")
    err := downloadData(); check(err)
  }

// Covert arg[1] to Title (first char uper case) to match Node struct
  args[0] = strings.Title(args[0])
  

// Get and sort xml values by query arguments
  arrayIPs := make(map[string]string)
  arrayRoles := make(map[string]string)
  keys := []string{}

  
  
  for  _, value := range data_unmarsh.Nodes {
// reflect used to use method dinamicaly
    s := reflect.ValueOf(value)
    f := s.FieldByName(args[0]).Interface().(string)

// populate map and keys if string found
    if strings.Contains(f, args[1]){
      arrayIPs[value.Desc] = value.Ip
      arrayRoles[value.Desc] = value.Roles
      keys = append(keys, value.Desc)  
    }
  }
  
  if (len(keys) == 0){
    fmt.Println("No results found!")
    os.Exit(0)
  }
// Sort keys 
  sort.Strings(keys)
   
// Initializa flux variables
  count := 1
  cssh_string := ""
  
// Run flux based on option
  for _, val := range keys {

// No index in arguments - print list
    if len(args) < 3 {
      fmt.Println(count, " - ", val," :: ", arrayIPs[val], " :: ",arrayRoles[val])
      
    } else { 

// Numeric argument 
      if IsNumeric(args[2]) { 
        index, err := strconv.Atoi(args[2])
        check(err)

// If match, connect and exit
        if index == count {
          connect.SshConn(arrayIPs[val], username, sshkey)
          os.Exit(0)
        }

// Literal - connect all list 
      } else if args[2] == "all" {
        connect.SshConn(arrayIPs[val], username, sshkey)
      
// Literal - create cssh string
      } else if args[2] == "cssh" {
        cssh_string += username + "@" + arrayIPs[val] + " "
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
