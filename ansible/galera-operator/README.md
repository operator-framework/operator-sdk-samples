WIP, not ready for general consumption.

# TODO

1. Add a finalizer:
   * Don't remove PVCs if a certain option is set
1. Load balancer
   * Add ability to specify load balancer kind (HAproxy, or native LoadBalancer type) or to disable it
1. Figure out if any nodes are up and set it as a fact
1. Move pod creation to a StatefulSet
   * Start up bootstrap pod only if no nodes are running
   * Kill bootstrap pod once the cluster is up?
1. Enhance pods
   * Add readiness checks.
1. Add more checks for variable changes:
   * mariadb_version change to upgrade pods.
1. Make more things configurable:
   * PV size requested
   * my.cnf options?
   * MySQL stuff:
     1. setting root password
     1. creating new user with permissions
     1. creating database

# BUGS

1. If node 1 is killed, it restarts with bootstrap options (see TODO above).
1. When shrinking the cluster, the HAProxy config is not changed.
1. When shrinking the cluster, nodes should be removed from HAProxy first to prevent connection hangs/timeouts.

# Questions / Stretch Goals

1. Create a Galera arbiter if the requested cluster size is an even number? Alternative would be to set the weight on the first node to 2 to give an odd quorum number.
