1. 
### description: 
    supposing in a random container:
    triggerA,triggerB all have variable $1;
    first called:B not ok;A is ok with $1=a
    second called:B is ok with $1=b
    as A is ok ,so the container is ok
    but A is ok when $1=a
    as for $1=b,that is unknown

### solution:
    let all nodes possess their own stateMap
    if a given state(variable values) can be found in the stateMap,
    then the node is ok