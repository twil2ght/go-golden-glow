## About Flow of a person's words

1. check sentence type
2. sentence type== if a person D
3. cond:add if A D
4. cond:add if A is a person
5. `[repo] [SSET] they -> A`
6. `[repo] [SSET] A is used`

```
if someone asks your name
then verify if you can tell them
if yes
then tell them

if dev asks Susie's name
check if Susie can tell dev Susie's name

if Susie can tell dev Susie's name
then Susie should tell dev Susie's name

some asks your name
verify if you can tell soemone-> how: who is someone:dev->verify if you can tell dev
you can tell someone
who is someone:dev
you should tell someone your name
who is someone: dev
you should tell dev your name
Warning: make someone the single:which means before this rond ends,you must not set someone;
when all states are reset,then you can set it again
```

```
how to verify if you know a person:
    verify if you know their name(if yes,then you know them)
        how to verify if you know a person's name:
            search their name in your database(if their name in your database then you know their name)
                how to search a person's name in Susie's database:
                    search the person's name in Susie's database  
                        [input] [repo] C @ the person -> B
                        [input] [repo] C @ B's name -> B1
                        [output] B's name in Susie's database
                a person's name in Susie's database
            Susie knows a person's name
        Susie knows a person   

Normal verison:           
    how to tell a person your name:
        check what your name is
        if your name is Susie(varible:Susie)
        then say "I am Susie"(check all Susie then to $Susie)
Easy version:
    how to tell a person your name:
        check what your name is
        if your name is $Susie          -> [input] Susie's name is A;[output] [repo] [SSET] $Susie -> A
        then say "I am $Susie" 
    real-use: check what Susie's name is 
                -> Susie's name is Susie
                -> [repo] [SSET] $Susie -> Susie
                -> $trigger: your name is $Susie
              say "I am $Susie"(check all variables and replace,this step can only be done at the low level api)   
```
```
first the "Susie" here is a placeholder;I have to check my name and the value is
Susie then I say I am Susie because I have to replace the placeholder with its true value

```