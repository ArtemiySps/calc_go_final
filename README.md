# Калькулятор на языке Golang

Калькулятор основан на взаимодействии двух серверов: оркестратора и агента, где оркестратор делит входное математическое выражение на отдельные задачи, а агент эти задачи вычисляет.

Взаимодействие оркестратора и агента происходит по gRPC, а связь пользователя с оркестратором - по HTTP. 

Все выражения сохраняются в базе данных SQL, доступ к ним доступен после перезагрузки калькулятора.

## Запуск

1. Клонируйте библиотеку через git clone 

2. Перейдите в папку с программой (calc_go_final) в двух cmd-терминалах:
```
cd calc_go_final
```

3.  В одном из них введите команду, которая запустит gRPC-сервер агента:
```
go run ./cmd/agent/main.go
```

4. В другом введите команду, которая подключит оркестратор к gRPC-серверу агента и запустит http-сервер для взаимодействия с пользователем:
```
go run ./cmd/orkestrator/main.go
```

4. Откройте cmd-терминал и введите команду:
```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1+2+3+4+5\"}"
```

6. Для того, чтобы вычислять одновременно два и более выражений, каждое новое следует запускать командой в отдельном терминале (либо, если вычисление предыдущего выражения завершилось - в том же):
```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"5-4-3-2-1\"}"
```

7. Чтобы получить все выражения, используйте в отдельном терминале команду:
```
curl -X POST http://localhost:8081/api/v1/expressions
```

8. Чтобы получить выражение по его ID, используйте команду:
```
curl -X POST http://localhost:8081/api/v1/expression/(любой ID, полученный из expressions)
```

9. Чтобы проверить сохранность выражений после перезагрузки калькулятора, рекомендуется завершить процесс (Ctrl+C) в окнах, где запускались сервера, а затем запустить их снова и повторно получить выражения

## Примеры запросов
```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1+2+3\"}"
```
Результат: 6

```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1*2+(3+4-5+(6/2))\"}"
```
Результат: 1

```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1+2/0\"}"
```
Результат: division by zero

```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1+2+a\"}"
```
Результат: unexpected symbol

```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1++2\"}"
```
Результат: incorrect expression


## Принцип работы (not updated)

В калькуляторе взаимодействуют пользователь, оркестратор и агент.
[[readme_assets/Scheme.png]]

Оркестратор принимает от пользователя выражение, добавляет его в хранилище, присваивая ID, представляет в виде обратной польской записи. Затем начинается деление выражения на простейшие задачи (задача представлена структурой Task, имеющей поля ID, Status, Arg1, Arg2, Operation, Operation_time, Result, Error). Только что сформированный task отправляется в хранилище задач. Функция деления на задачи уходит в ожидание до тех пор, пока статус у отправленной им задачи не поменяется на "completed".

Всё это время, сразу после запуска сервера агента, он отправляет оркестратору GET-запросы каждые 500 ms (можно изменить в .env), желая получить задачу. При каждом запросе идёт проверка на наличие нерешённых задач (с Status == "need to send") в хранилище задач. Если такие имеются, то агенту выдаётся первая нерешенная задача из стека задач. 

Попав к агенту, задача решается определённым количеством параллельно запущенных воркеров. Их количество задаётся переменной среды COMPUTING_POWER. В калькуляторе имитируются долгие вычисления, поэтому переменными среды задано время выполнения для каждой математической операции. Как только один из воркеров завершит вычисление задачи, останавливается работа всех остальных, а результат (либо ошибка, если таковая обнаружилась в процессе вычисления) отправляется обратно оркестратору. Следующим GET-запросом агент берет следующую нерешённую задачу.

Получив результат, оркестратор изменяет вычисленную задачу в хранилище, добавляя в соответствующие поля результат/ошибку, а также меняет статус на "completed". Функция деления на задачи выходит из режима ожидания и анализирует результаты: если ошибки не возникло - подставляем результат вычислений в следующую задачу, если ошибка всё же возникла - возвращаем её пользователю.

Как только всё выражение будет вычислено без ошибок - выводим результат пользователю.

В процессе работы сервера можно также просмотреть хранилище выражений, и найти выражение по ID с помощью запросов end-поинтами "/api/v1/expressions" и "/api/v1/expression/:id" соответственно.


## Структура проекта (not updated)
```
calc_go_2.0
├── cmd
│   ├── agent
│   │   └── main.go
│   └── orkestrator
│       └── main.go
├── env 
│   └── .env
├── internal
│   ├── agent
│   │   ├── config
│   │   │   └── config.go
│   │   └── service
│   │       ├── agent.go
│   │       └── agent_test.go
│   └── orkestrator
│       ├── config
│       │   └── config.go
│       ├── http
│       │   └── http.go
│       └── service
│           ├── orkestrator.go
│           └── orkestrator_test.go
├── pkg
│   └── models
│       ├── errors.go
│       ├── models.go
│       └── operations.go
├── go.mod
└── go.sum
```

#### cmd
- agent/main.go - запуска сервера агента
- orkestrator/main.go - запуск сервера оркестратора

#### internal/agent - файлы агента
- config/config.go - конфигурирование агента
- service/agent.go - реализация агента

#### internal/orkestrator - файлы оркестратора
- config/config.go - конфигурирование оркестратора
- http/http.go - реализация хендлеров
- service/orkestrator.go - функции оркестратора (деление выражения на задачи, добавление выражений и задач в хранилище, изменение параметров выражений и задач, получение всех выражений и задач из хранилища)

#### env
- .env - переменные среды

#### pkg/models
- errors.go - тексты ошибок
- models.go - структуры
- operations.go - функции создания ID через uuid, преобразования выражения в обратную польскую запись

## Переменные среды

Переменные среды и их значения по умолчанию (значения можно изменить в .env):
```
TIME_ADDITION_MS=3000           #
TIME_SUBTRACTION_MS=3000        # времена выполнения
TIME_MULTIPLICATION_MS=3000     # математических операций
TIME_DIVISION_MS=2000           #

COMPUTING_POWER=1               # количество запускаемых воркеров

PORT_ORKESTRATOR=8081           # порт для запуска http сервера-оркестратора

PORT_AGENT=8080                 # порт для запуска grpc сервера-агента
HOST_AGENT=localhost            # хост для запуска grpc сервера-агента



curl -X POST http://localhost:8081/api/v1/register -H "Content-Type:application/json" -d "{\"login\":\" \",\"password\":\" \"}"

curl -X POST http://localhost:8081/api/v1/login -H "Content-Type:application/json" -d "{\"login\":\" \",\"password\":\" \"}"

curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -H "Authorization:<token>" -d "{\"expression\":\" \"}" 

curl -X POST http://localhost:8081/api/v1/expressions -H "Authorization:<token>"

curl -X POST http://localhost:8081/api/v1/expression/<id> -H "Authorization:<token>"

curl -X POST http://localhost:8081/api/v1/clear -H "Authorization:<token>"