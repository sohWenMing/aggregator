# The Aggregator Project #

## Purpose Of This Project ##
The purpose of this project was to practice working on creating a command line tool that worked with PostgreSQL to aggregate RSS feeds
from different sources. In the process of this project, I also had learnt how to deal with reading of files, unmarshalling and marshalling 
of data (both in XML and JSON format), configurations of goose and sqlc for the purposes of integrating with PostgreSQL and also installing 
and running an instance of PostgreSQL locally while interacting with the program through the psql command line tool. 


## Prerequisites ##

- either postgresql installed locally, or access to a postgresql instance. 
- a file ".gatorconfig.json" which is installed in the home folder of the system

### Contents of the .gatorconfig.json file ###
{
  "db_url": *\<enter db string here>*,  
  "current_user_name": "nindgabeet"
}


## Installation Instructions ##
After setting up of the .gatorconfig.json file, assuming that the connecting to the PostgreSQL instance is valid, the program should be ready
to run. Options are to either 
 - build a binary using the command go build -o *\<enter desired filename here>*
 - run **go install** from the root of the project