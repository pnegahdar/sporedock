SporeDock
=========

SporeDock is an opinionated Docker orchestration tool that makes it easy to scale and distribute Docker apps on a multi-node/host cluster. There are lots of tools around Docker Orchestration right now (see: Deis, Flynn, Shipyard, etcd, Coreos, Fleet, Serf, Bob, vlucand, hipache, Dokku, Mesos, Mesosphere, etc) but figuring out which to use and making them work together is difficult. SporeDock makes those decisions for you.


## Architecture

SporeDock has 3 major components:

  **Etcd** : Internode communication, Discovery, Process synchronization, Primary election

  **Docker Client** : Managing Docker daemon

  **vulcand** : Dynamic http load balancer


  Together they give you a highly dynamic self-discovering deployment stack powered by docker.

 All the tools are written in Go to produce a single distributable binary that acts as both the server and client. Making cloud formations of these binaries are very easy.


## Interface

Set up the discovery url for nodes to discover one another on their private network

     sporedock init <discovery_url>

Begin sporedock daemon on the node:

    sporedock start


**Thats it, node is now part of cluster**

Logs all the nodes into the docker registry where images will be retrieved

    sporedock registry login <user> <pass>

Set up environments

    sporedock env create <env_name>
    sporedock env add -env <env_name> -key <key> -value <value>
    sporedock env load -env <env_name> -file <file>


Distribute Apps. Gets the apps ready to deploy on all the nodes

    sporedock apps add -name <app_name> -image <docker_image> -binds <optinal:hostname> -env <env_name>

List Apps

    sporedock apps list

Scale/Stop Apps. Automatically distributing them accross the cluster

    appbox deploy <app_name> 10
    appbox stop <app_name>
    appbox deploy <app_name> 5


Apps/Webapps are automatically moved around as needed to maximize their distribution. If a node binds to a webapp it will be added to Vulcan's load balancer dynamically and traffic will be routed to the correct place as the node moves around. Should a node die, all procs will be redistributed.