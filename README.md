# Abstract
The purpose of this piece of code is to compare Clickhouse performances on several kinds of storage.

# Disclaimer
This glorious software has been written in "Quick and Dirty" mode, without any pretentions. It just does the job.

# What it does
It launches a webserver with REST API and waits for HTTP GET command. The rest command creates a table named "perftest" (if it doesn't exsist) using TableMerge engine. Then the API call inserts several JSON records into this database. Table structure is quite simple: it contains a timestamp (always the same, that doesn't matter) and a JSON structure with workstation name and user name. User name us formatted as "user " + <iteration>. Workstation name is formatted ad "wks" + <iteration>.

# Configuration
## Clickhouse
Clickhouse has to be configured with at least2 storages:
- S3 : S3 disk. This disk is targeted with a specific storage policy.
- NAS : usual NAS mount. This disk is targeted with a specific storage policy.

## App
The app settings are defined in config.yaml. If the piece of software is executed in a container, a custom file should be mounted as /config.yaml
Settings are the following : 
- clickhouse/server : FQDN or IP address of CH server
- clickhouse/port : the port of Clickhouse native (not SQL) client
- clickhouse/login : login of Clickhouse connection
- clickhouse/password : password of Clickhouse connection

# Usage
## Golang nerd mode
go run *.go

## Docker mode
docker run -p 8080:8080 -v <your_yaml_file>:/bin/config.yaml sdmitriev/chbench:latest

## When it is launched
When the piece of software is launched (dockerised or purely local) it can be used with curl, httpie, postman or even a simple browser.

For single ingestion test, the URL is :
http://<destination>:8080/<name_of_the_policy>/<number_of_records_to_insert>

For repeated ingestions, the URL is :
http://<destination>:8080/batch/<name_of_the_policy>/<number_of_records_to_insert>/<number_of_iterations>

With:
- name_of_the_policy : the name of clickhouse storage policy. You can use any policy name, not only "s3" or "NAS". Really, whatever you need and what is defined in your Clickhouse XML settings file
- number_of_records_to_insert : the number of records that you need to insert in the database
- number_of_iterations : the number of successive ingestions to perform

## Results
The results are displayed as logs on docker console (or on STDOUT of the computer running the Go code).