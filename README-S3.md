secret-squirrel
===============

Secret Squirrel is a proof-of-concept project which provides secret access credentials to docker containers without docker leakage of the credentials, and with the credentials only existing in memory.

This proof-of-concept is designed to work around the following setup, but the approach could easily be extended to work with Vault or KMS or a secret-store of your choice. Please check the simpler README.md for a much simpler proof-of-concept:

* I am running an EC2 instance on AWS
* I have a secrets repository in an s3 bucket called s3://my-secrets-bucket-<Your Name>
* I have an IAM instance profile attached to the EC2 instance that allows reading of objects from this bucket
* I have an object in by bucket with name "SECRETPASSWORD" and content "PRIVATE99" - (unencrypted for the POC)
* I have a docker image I wish to run, whose entrypoint is /bin/dockerstart, (See notes at end for a more general case)
* When run, the resulting container wants access to the secret via environment variable `${SECRETPASSWORD}`

Prerequisites
=============

The build-phase of this project assumes you are running docker-machine with a "default" docker-machine

Building
========

Now build `secret_squirrel_s3` binary:

```
    $ ./build.sh
```

If successful, you should find the `secret_squirrel_s3` binary in the lib/ folder

The `secret_squirrel_s3` binary this project produces is a statically linked binary, and so is self-contained and should work on any linux system.

Setup
=====

* Put `secret_squirrel_s3` binary in /sbin on the docker host (probably docker-machine VM), and chown 755

```
    # YMMV with this >>

    $ cat lib/secret_squirrel_s3 | docker-machine ssh default sudo tee /sbin/secret_squirrel_s3 >/dev/null 2>&1
    $ docker-machine ssh default sudo chmod 755 /sbin/secret_squirrel_s3
```

Running
=======

This can be used within docker itself, for containers with an entrypoint of /bin/dockerstart as follows:

* Run your docker app as follows:

```
    docker run --name test_app \
               -v /sbin/secret_squirrel_s3:/sbin/secret_squirrel_s3 \
               --entrypoint /sbin/secret_squirrel_s3 \
               -e "CREDENTIALS_SECRETPASSWORD=true" \
               -e "BUCKETSUFFIX=<Your name>" \
               myapp:latest with_arg1 with_arg2 with_arg3
```

Or if you dont have a myapp to test, you can simply do:

```
    docker run --name test_app \
               -v /sbin/secret_squirrel_s3:/sbin/secret_squirrel_s3 \
               --entrypoint /sbin/secret_squirrel_s3 \
               -e "CREDENTIALS_SECRETPASSWORD=true" \
               -e "BUCKETSUFFIX=<Your name>" \
               -it \
               ubuntu:latest bash -c 'echo $SECRETPASSWORD && sleep 60'
```

If you have trouble with complaints about x509 certs, it may be because your image is not a full-flavoured linux, but a busybox/alpine image without the x509 root certs installed, in which case you may also need to volume mount those with:

```
    -v /etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt:ro
```

Notice that when your app is running, a docker inspect DOES NOT SHOW your SECRETPASSWORD:

```
    $ docker inspect --format '{{.Config.Env}}' test_app

    [BUCKETSUFFIX=gtmtech CREDENTIALS_SECRETPASSWORD=true]
```

Yet the output of the test_app (in the 2nd example) is to clearly display the password:

```
    [test_app] PRIVATE99
```

Here is a corresponding example for running on marathon/mesos:

```
    {
      "id": "myapp",
      "instances": 1,
      "args": ["with_arg1", "with_arg2", "with_arg2"],
      "container": {
        "type": "DOCKER",
        "docker": {
          "image": "myapp:latest",
          "network": "BRIDGE",
          "parameters": [
            {"key":"env", "value":"BUCKETSUFFIX=gtmtech"},
            {"key":"env", "value":"CREDENTIALS_SECRETPASSWORD=true"},
            {"key":"entrypoint", "value":"/sbin/secret_squirrel_s3"},
            {"key":"volume", "value":"/sbin/secret_squirrel_s3:/sbin/secret_squirrel_s3:ro"},
            {"key":"volume", "value":"/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt:ro"}
          ]
        }
      }
    }
```

To recap, process 1 in your container has the password in its memory, via an environment variable which docker itself does not know about.

Until docker provide a secrets-api or a secrets-driver, this allows the user the greatest flexibility to create injected secrets which:

* Do not ever get stored on disk
* Are only in memory
* Docker has no access to
* Containers need no work to interface with (but see Notes section below if entrypoint isnt `/bin/dockerstart`)
* Can be retrieved from virtually any destination

Explanation
===========

The proof-of-concept `secret_squirrel_s3` binary above analyses its own environment (environment vars) to determine which credentials to fetch from the s3 bucket via the iam_instance_profile enabled mechanism 

The following diagram explains how it might work in a marathon/mesos setup:

![Secret Squirrel Sequence Diagram](./secret_squirrel_s3.png?raw=true "Secret Squirrel Sequence Diagram")

* Firstly it uses the value of the environment variable `BUCKETSUFFIX` to determine which s3 bucket the credentials should be picked up from. The s3 bucket the credentials are fetched from is `s3://my-secrets-bucket-${BUCKETSUFFIX}`

* Next, it scans all environment vars for vars beginning with "CREDENTIALS_" . Any that it finds it attempts to fetch as bucket objects from the bucket (after removing the CREDENTIALS_ prefix first), and add the corresponding bucket object data to the current environment as a key/value pair.

Finally the `secret_squirrel_s3` binary sys-execs the `/bin/dockerstart` process with the new environment vars. For example:

```
    export BUCKETSUFFIX=gtmtech
    export CREDENTIALS_SECRETPASSWORD=true  # set this env var to tell secret_squirrel to fetch SECRETPASSWORD

    $ ./secret_squirrel
       => (retrieves bucket object) s3://my-secrets-buckets-gtmtech/SECRETPASSWORD
          with content "PRIVATE99"
       => (sets environment variable) SECRETPASSWORD=PRIVATE99
       => "exec"s /bin/dockerstart 
```

Further Notes
=============

If the docker image you wish to run does not have `/bin/dockerstart` as its entrypoint, you may be able to have some success, by passing the ENTRYPOINT through to a modified `secret_squirrel_s3` binary, in a 2-step process e.g.:

```
    $ ENTRYPOINT=$( docker inspect --format '{{.Config.Entrypoint}}' myapp:1 )
    $ docker run -v /sbin/secret-squirrel_s3:/sbin/secret_squirrel_s3:ro 
                 --entrypoint /sbin/secret_squirrel_s3
                 -e "ORIGINAL_ENTRYPOINT=$ENTRYPOINT" 
                 -e "CREDENTIALS_SECRETPASSWORD=true" 
                 -e "BUCKETSUFFIX=gtmtech" 
                 myapp:latest with_arg1 with_arg2 with_arg3

This could be done ahead of pushing to marathon. Note the `secret_squirrel_s3` go code would need to be modified to exec the value of ORIGINAL_ENDPOINT env var, if it exists, instead of `/sbin/secret_squirrel_s3` but this is trivial.

Other security notes
====================

* As the script is available on the host, it could be run accidentally on the host. Because `/bin/dockerstart` does not exist on the host, it is not available and secret_squirrel bails out early. An alternative would be to have a dumb `/bin/dockerstart` which does nothing but exit on the host too

* If the container was compromised by an RCE, it is possible that the `/sbin/secret_squirrel_s3` binary could be run again via another process in that container. However because it forcibly execs `/bin/dockerstart` afterwards, it would not be possible to capture the output. It would be possible however if (in the case where you override with `ORIGINAL_ENTRYPOINT` env var and the binary accepts that as its exec destination) that the binary could be made to divulge secrets, if the container was already exploited via an RCE. (For this reason it is better to have conventions about entrypoints hard-coded and only run images with that entrypoint in place which can be achieved with good Docker build practices)

