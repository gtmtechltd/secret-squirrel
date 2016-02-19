package main

import (
         "fmt"
         "os"
         "syscall"
         "strings"
         "container/list"
)

func main() {

         executable := "/bin/dockerstart"
         credentialsEnvKeyPrefix := "CREDENTIALS_"

         if _, err := os.Stat(executable); os.IsNotExist(err) {
                 fmt.Println("secret_squirrel: " + executable + " does not exist")
                 os.Exit(1)
         }

         keys := list.New()

         for _, e := range os.Environ() {
                 pair := strings.Split(e, "=")
                 key := pair[0]
                 if strings.HasPrefix(pair[0], credentialsEnvKeyPrefix) {
                         key := key[len(credentialsEnvKeyPrefix):len(key)]
                         keys.PushBack(key)
                 }
         }

         for keyi := keys.Front(); keyi != nil; keyi = keyi.Next() {
                 keyString := keyi.Value
                 if key, ok := keyString.(string); ok {
                         fmt.Printf("secret_squirrel: Acquiring credentials for %s\n", key)
 
                         // include any mechanism here to get from KMS, S3, VAULT, etc.
                         os.Setenv(key, "SUPER_PRIVATE_PASSWORD")
                 } else {
                         fmt.Printf("secret_squirrel: Could not acquire credentials for %s\n", key)
                 }
         }
         
         // if no error, then proceed to list the contents/objects inside the bucket
         syscall.Exec(executable, os.Args, os.Environ())
}
