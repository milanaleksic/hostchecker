# Host Checker

[![Build Status](https://semaphoreci.com/api/v1/milanaleksic/hostchecker/branches/master/badge.svg)](https://semaphoreci.com/milanaleksic/hostchecker)

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
    ]
  },
 // ... other servers to be checked
]
```