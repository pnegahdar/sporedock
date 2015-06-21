SporeDock
==========

**Status**: Not Runable (Design Phase). WIP.

SporeDock is an opinionated Docker orchestration tool that makes it easy to scale and distribute Docker apps on a multi-node/host cluster. There are lots of tools around Docker Orchestration right now (see: Deis, Flynn, Shipyard, etcd, Coreos, Fleet, Serf, Bob, vlucand, hipache, Dokku, Mesos, Mesosphere, etc) but figuring out which to use and making them work together is difficult. SporeDock makes those decisions for you.



## Interface

Set up the discovery url for nodes to discover one another on their private network

     sporedock init <discovery_url>

Begin sporedock daemon on the node:

    sporedock start


**Thats it, node is now part of cluster**

Set up environments

    sporedock env create <env_name>
    sporedock env add -env <env_name> -key <key> -value <value>
    sporedock env rm -env <env_name>


Distribute Apps. Gets the apps ready to deploy on all the nodes

    sporedock apps add -name <unique_app_name> -image <docker_image> -env <env_name>

List Apps

    sporedock apps list

Scale/Stop Apps. Automatically distributing them across the cluster

**Web Apps**

    sporedock launch_webapp <app_name> --scale <count> --host 'dev.daefs.apps.kensho.com' // Test the endpoint
    sporedock addhost <app_name> --host 'staging.kensho.com'
    sporedock rmhost <app_name> --host 'staging.kensho.com'

**Service Apps**

    sporedock launch <app_name> --scale <count>

**ALL**

    sporedock stop <app_name>


`Web|Workers` are automatically moved around as needed to maximize their distribution. If a node binds to a webapp it will be added to Vulcan's load balancer dynamically and traffic will be routed to the correct place as the node moves around. Should a node die, all procs will be redistributed.


