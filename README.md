
# build
```shell
docker build -t yangzxi/film-probe:latest .
```

# run
```shell
docker run -p 8801:8080 --name film-probe -d yangzxi/film-probe:latest
```