package main

import (
         "fmt"
         "github.com/mitchellh/goamz/aws"
         "github.com/mitchellh/goamz/s3"
         "os"
         "syscall"
         "strings"
         "container/list"
)

func main() {

         executable := "/bin/dockerstart"
         bucketPrefix := "my-secrets-bucket-"
         bucketSuffixEnvKey := "BUCKETSUFFIX"
         credentialsEnvKeyPrefix := "CREDENTIALS_"


         if _, err := os.Stat(executable); os.IsNotExist(err) {
                 fmt.Println("secret_squirrel: " + executable + " does not exist")
                 os.Exit(1)
         }

         environment, present := os.LookupEnv(bucketSuffixEnvKey)
         if !present {
                 fmt.Println("secret_squirrel: " + bucketSuffixEnvKey + " not set")
                 os.Exit(1)
         }

         // Get an auth from the iam_instance_profile attached to the running ec2 instance
         AWSAuth, _ := aws.GetAuth("", "") 
     
         // Get the correct region
         region := aws.EUWest // TODO: use Regions() instead which is a map

         connection := s3.New(AWSAuth, region)

         credentialsBucket := bucketPrefix + environment
         bucket := connection.Bucket(credentialsBucket)
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
                         response, err := bucket.Get(key)

                         fmt.Printf("secret_squirrel: Acquiring credentials for %s\n", key)
                         if err != nil {
                                 fmt.Printf("secret_squirrel: Error: %s\n", err)
                                 os.Exit(1)
                         }
                         value := strings.TrimSpace(string(response))
                         os.Setenv(key, value)
                 } else {
                         fmt.Printf("secret_squirrel: Could not acquire credentials for %s\n", key)
                 }
         }
         
         // if no error, then proceed to list the contents/objects inside the bucket
         syscall.Exec(executable, os.Args, os.Environ())
}
