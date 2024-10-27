## Создание базового профиля
```bash
curl http://localhost:8080/debug/pprof/heap > base.pprof
```

## Создание обновленного профиля
```bash
curl http://localhost:8080/debug/pprof/heap > result.pprof
```

## Сравнение результата
```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof 
```
