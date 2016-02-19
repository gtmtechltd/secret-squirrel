secret-squirrel
===============

Secret Squirrel is a proof-of-concept project which provides secret access credentials to docker containers without docker leakage of the credentials, and with the credentials only existing in memory.

There is a more advance proof-of-concept in README-S3.md, however this proof-of-concept will supply a private fixed password (hardcoded in the `secret_squirrel` binary) to any container. For non-hardcoded password check out the more advanced proof-of-concept.

Prerequisites
=============

The build-phase of this project assumes you are running docker-machine with a "default" docker-machine

Building secret-squirrel
========================

First build the test_app

```
    $ (cd test_app && ./build.sh)
```

This should give you a local docker_image called test_app:1

Now build `secret_squirrel` binary:

```
    $ ./build.sh
```

If successful, you should find the `secret_squirrel` binary in the lib/ folder (the other file is for the more advanced proof-of-concept)

The `secret_squirrel` binary this project produces is a statically linked binary, and so is self-contained and should work on any linux system.

Setup
=====

* Put `secret_squirrel` binary in /sbin on the docker host (probably docker-machine VM), and chown 755

```
    # YMMV with this >>

    $ cat lib/secret_squirrel | docker-machine ssh default sudo tee /sbin/secret_squirrel >/dev/null 2>&1 
    $ docker-machine ssh default sudo chmod 755 /sbin/secret_squirrel
```

Running
=======

The test_app is built with an `ENTRYPOINT` of ["/bin/dockerstart"]. The `secret_squirrel` binary assumes that will always be your `ENTRYPOINT` (see Notes at bottom for different entrypoints).

* Run your docker app as follows:

```
    docker run --name test_app1 \
               -v /sbin/secret_squirrel:/sbin/secret_squirrel \
               -e "CREDENTIALS_SECRETPASSWORD=true" \
               test_app:1 with_arg1 with_arg2 with_arg3

    I am /bin/dockerstart and I was called with args: with_arg1 with_arg2 with_arg3
    SECRETPASSWORD is 
```

As you can see, there is no SECRETPASSWORD set.

* Now run your docker app as follows:

```
    docker run --name test_app2 \
               --entrypoint /sbin/secret_squirrel \
               -v /sbin/secret_squirrel:/sbin/secret_squirrel:ro \
               -e "CREDENTIALS_SECRETPASSWORD=true" \
               test_app:1 with_arg1 with_arg2 with_arg3

    I am /bin/dockerstart and I was called with args: with_arg1 with_arg2 with_arg3
    SECRETPASSWORD is SUPER_PRIVATE_PASSWORD
```

Notice that whilst your app is running, a docker inspect DOES NOT SHOW the value of your SECRETPASSWORD:

```
    $ docker inspect --format '{{.Config.Env}}' test_app2

    [CREDENTIALS_SECRETPASSWORD=true]
```

Yet the output of the test_app2 (in the 2nd example) is to clearly display the password:

```
    I am /bin/dockerstart and I was called with args: with_arg1 with_arg2 with_arg3
    SECRETPASSWORD is SUPER_PRIVATE_PASSWORD
```

To recap, process 1 in your container has the password in its memory, via an environment variable which docker itself does not know about.

Until docker provide a secrets-api or a secrets-driver, this allows the user the greatest flexibility to create injected secrets which:

* Do not ever get stored on disk
* Are only in memory
* Docker has no access to
* Containers need no work to interface with (but see Notes section below if entrypoint isnt `/bin/dockerstart`)
* Can be retrieved from virtually any destination

Further Notes
=============

If the docker image you wish to run does not have `/bin/dockerstart` as its entrypoint, you may be able to have some success, by passing the ENTRYPOINT through to a modified secret_squirrel binary, in a 2-step process e.g.:

```
    $ ENTRYPOINT=$( docker inspect --format '{{.Config.Entrypoint}}' myapp:1 )
    $ docker run -v /sbin/secret_squirrel:/sbin/secret_squirrel:ro
                 --entrypoint /sbin/secret_squirrel 
                 -e "ORIGINAL_ENTRYPOINT=$ENTRYPOINT" 
                 -e "CREDENTIALS_SECRETPASSWORD=true" 
                 myapp:latest with_arg1 with_arg2 with_arg3
```

... but this may have security implications (see below)

Other security notes
====================

* As the script is available on the host, it could be run accidentally on the host. Because `/bin/dockerstart` does not exist on the host, it is not available and secret_squirrel bails out early. An alternative would be to have a dumb `/bin/dockerstart` which does nothing but exit on the host too

* If the container was compromised by an RCE, it is possible that the `/sbin/secret_squirrel` binary could be run again via another process in that container. However because it forcibly execs `/bin/dockerstart` afterwards, it would not be possible to capture the output. It would be possible however if (in the case where you override with `ORIGINAL_ENTRYPOINT` env var and the binary accepts that as its exec destination) that the binary could be made to divulge secrets, if the container was already exploited via an RCE. (For this reason it is better to have conventions about entrypoints hard-coded and only run images with that entrypoint in place which can be achieved with good Docker build practices)

Next step
=========

Now check out README-S3.md for a more complete example using S3 buckets.
