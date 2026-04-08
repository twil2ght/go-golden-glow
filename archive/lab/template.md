## template implement plan
1. 把AC的C换掉后，创建个新node，
2. 这个新node注册到nodePool，
3. containerOf需要重写
4. 创建新container，但是不能被repo存储
### steps
1. find the A C node
2. make new node
3. make new container
4. save in memory