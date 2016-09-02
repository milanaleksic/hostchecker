# Host Checker

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/hostchecker/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/hostchecker)
[![Docker Stars](https://img.shields.io/docker/stars/milanaleksic/hostchecker.svg?maxAge=2592000)]()
[![Docker Pulls](https://img.shields.io/docker/pulls/milanaleksic/hostchecker.svg?maxAge=2592000)]()
[![Docker Automated buil](https://img.shields.io/docker/automated/milanaleksic/hostchecker.svg?maxAge=2592000)]()

## What it's all about
 
This tiny CLI app verification of services. 

It was meant to block execution of tests if deployment wasn't a full success.
 
To run it, you have to have a file called `expectations.json` beside the app.

## expectations.json DSL

```txt
[
  {
    "server": "<server_location>:22",
    "user": "<ssh username>",
    "password": "<ssh password>",
    "upstart": [
      {
        "name": "<service_name>",
        "user": "<expected_service_app_executor>",
        "newerThanSeconds": 3600, // how old is service allowed to be 
        "ports": [
          1312 // which port should be occupied by the service?
        ]
      },
      // ... other upstart services
    ],
    "custom": [
      {
        "name": "Service name which is ignored in this case",
        "regex": "[s]ome process name", // how to identify process in PS listing
        "user": "<expected_service_app_executor>",
        "newerThanSeconds": 3600, // how old is service allowed to be 
        "ports": [
          9090// which port should be occupied by the service?
        ]
      }
    ],
    // which URLs should be checked
    "responses": [
      {
        "name": "Mongo response",
        "url": "http://mongo_site:27017/",
        // what codes are acceptable
        "codes": [
          200
        ],
        // optionally, check what is the response from the server
        "response": "It looks like you are trying to access MongoDB over HTTP on the native driver port.\n"
      }
      ],
      // freestyle shell commands: they have to return 0 to succeed
      "shell": [
        {
          "name": "Is Mongo operational?",
          // which command do we want to execute remotely?
          "cli": "mongo localhost/temp --eval \"print('searchOnNonExistingId='+db.mycollection.find({'_id':-1}).length())\" | grep searchOnNonExistingId",
          // optionally, make sure string stdout output of the command is identical to sth we expect
          "expected": "searchOnNonExistingId=0"
        }
      ]
  },
 // ... other servers to be checked
]
```

## How to run from Docker:

To run from within a docker container, you need to map local file into `expectations.json` inside the container, here is the magic:

    docker run --rm \
        -v ${PWD}/expectations.json:/go/src/app/expectations.json 
        milanaleksic/hostchecker:<version>