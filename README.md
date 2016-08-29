# Launch

A simple mechanism for deploying and running scripts or executables on remote hosts.

Launch is developed in Go so is able to use concurrency to run things simultaneously where appropriate.

Each instance of Launch will take on the role of an agent when started as a daemon, else will be used to interact with one or more remote instances.

A message queue is used to communicate instructions and recieve status updates to and from each of the Launch daemons.

## Launchd

The launch daemon or program

