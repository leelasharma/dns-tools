# dns-tools

[![Build Status](https://travis-ci.org/egymgmbh/dns-tools.svg?branch=master)](https://travis-ci.org/egymgmbh/dns-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/egymgmbh/dns-tools)](https://goreportcard.com/report/github.com/egymgmbh/dns-tools)

DNS tools that we use for SRE work.

## Introduction

    +-----------+       +-----------------+
    | +---------+-+     | Service Account |
    | | +---------+-+   | File (JSON)     |
    +-| |           |   +-----------------+
      | | Zone Data |            |                 ,--.
      +-| (YAML)    |            v             _.-(    )_
        |           |----->  dns-tools  <---->(_CloudDNS_)
        +-----------+            ^
                                 |
                        +-----------------+
                        | Configuration   |
                        | File (YAML)     |
                        +-----------------+


## Tools

* `dbcheck` Loads zone data from a local directory into a in-memory database and
  verifies that all configured managed zones are retrievable from that database.
  We use this tool to test our zone data changes in a continuous deployment
  pipeline.
* `mzcreate` Connects to a Google Cloud Platform project and creates all managed
  zones that are in the local configuration file but missing on CloudDNS.
  We use this tool to easily add new zones by just adding them in the YAML files
  and letting our continuous integration deployment take care of the rest.
* `mzdump` Fetches all CloudDNS managed zones from a Google Cloud Platform
  project and prints the associated meta data, such as nameservers, using a
  template.
  We used this tool to create [JIRA](https://www.atlassian.com/software/jira)
  tickets for nameserver changes using JIRA's CSV import function.
* `mzmon` Fetches all CloudDNS managed zones from a Google Cloud Platform
  project and uses DNS lookups to check if delegations are correctly implemented
  at the corresponding Top Level Domain (TLD) nameservers. The results are then
  written to an InfluxDB time series database in a configurable interval.
  We use this tool to continuously monitor the delegations of our
  business-critical domains.
* `rrlookup` Loads zone data from a local directory into a in-memory database
  and looks up all records of all configured managed zones and compares the
  results with the expected values from the database.
  We use this tool to check for DNS errors and to verify our deployments went
  well.
* `rrpush` Loads zone data from a local directory into a in-memory database
  and deploys all records of all configured managed zones to CloudDNS.
  We use this tool to rapidly deploy changes to our zone data repository into
  production.


## Configuration

## Zone Data

## License

    Copyright 2017 eGym GmbH <support@egym.de>
    Copyright 2017 Dan Luedtke <dan.luedtke@egym.de>

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
