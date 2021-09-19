# diane
DIANE stands for "DNS is apparently not easy" and was created as a personal 
project. It is a simple application written in [Go](https://golang.org/)
because I wanted to learn more about Go as well as using 
[Docker](https://www.docker.com), [Prometheus](https://prometheus.io/), 
and [Grafana](https://grafana.com/).

Currently, DIANE is peforming the following functionality:
* Reporting on WHOIS information for configured domains
* More to come!


## Build and Test
To build and test this application, run the `make` command for these functions:

* `make build`  
  Builds the application with a couple additional options..
* `make test`  
  Runs the `go test` command with a couple additional options.
* `make review`  
  Runs the `go test cover` command followed by opening your browser to review code coverage.
* `make container`  
  Runs the `docker build` command with a designated Dockerfile and additional build options.
* `make local`  
  Runs the `docker compose` command to spin up the application in a docker image along
  with Prometheus and Grafana. 
* `make clean-local`  
  Runs the `docker compose` command to tear down the spun up containers and network.

## Run Locally
That `make local` command will spin up the application in a container along with 
Prometheus and Grafana containers associated with it. The output of the command will
show the running local containers along with the URL with dynamic port for Grafana. Once
Grafana is available, log in and look for the DIANE dashboard. Cheers!

## Configuration
The configuration file is in YAML format and exists as `configs/diane.yaml` for the time being. The structure is as follows:

* _domains_  
  Array of domain names and will be queried with the whois protocol.




## References
1. https://github.com/golang-standards/project-layout
1. https://www.viget.com/articles/two-ways-to-share-git-hooks-with-your-team/

