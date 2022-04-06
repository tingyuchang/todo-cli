# todo-cli

todo-cli is a commandline tool to manage your task. it will create a csv file to store data


## how to use

### build
```
go build -o todo
```
### insert data
```
./todo insert <TASKNAME>
```

###  search
search support search id and status (status current only support number)
```
./todo search id <ID>
./todo search status <STATUS>
```
### update
update task status
```
./todo update <ID> <STATUS>
```
### delete
delete task
```
./todo delete <ID>
```

### list
list all tasks
```
./todo list
```


### status
| number| status |
|---|---|
|0|backlog|
|1|in progress|
|2|review|
|3|done|