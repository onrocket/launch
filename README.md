# Status

this is prototype only, at present there is no published working code (yet)

Code Branches will be used to implement each 'stage' of deployment.

There are no stages yet but the first branch will be 'launch_config'

As each stage (branch) is complete it will be merged into 'master' 

# Launch

A simple mechanism for deploying and running scripts or executables on remote hosts.

Launch is developed in Go so is able to use concurrency to run things simultaneously where appropriate.

Each instance of Launch will take on the role of an agent when started as a daemon, else will be used to interact with one or more remote instances.

A message queue is used to communicate instructions and recieve status updates to and from each of the Launch daemons.

## Launcher

The launch daemon or program

### Launch Sequence

A set of commands make up a launch sequence. Some may be run concurrently,,
others in a given order.

Commands are any executable or script that is compatible with the target server.

### Launch Configuration

A launch configuration containing all information necessary to populate a set 
of launch commands in a launch sequence

### Command Sequencing

Launch commands may be preceded with a series of digits from 0..N where the 
lowest number is executed first and each in turn in ascending order there after.

for example :

    00_check_disks
    01_check_memory
    02_check_cpu

would in order check disk, memory and then CPU

### Command pre test

A command that has a test associated will run the test prior to running the 
command itself, so a command named `03_install_webservice_test` would be run 
before `03_install_webservice` and if the test command returns an exit code of 0
`03_install_webservice` will not be run.

This way, commands to 'enforce' the installation of packages for instance will
only be executed if the `_test` command of a given command sequence returns a 
non zero value.


